package tcpclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

const (
	_defMaxReconnectCycle = 30 * time.Second
	_defRecvTimeout       = 10 * time.Second
	_defConnTimeout       = 180 * time.Second
)

type ClientOption interface {
	// 最大重连间隔
	MaxReconnectCycleOpt(defValue time.Duration) time.Duration
	// 响应超时
	RecvTimeoutOpt(defValue time.Duration) time.Duration
	// 连接超时
	ConnTimeoutOpt(defValue time.Duration) time.Duration
}

var _ Client = (*TCPClient)(nil)

type TCPClient struct {
	// 接收由对端发来的消息
	MessageChanMap map[uint32]chan Messager
	chanMapMu      sync.RWMutex

	// 当前客户端持有的连接
	conn Connetion
	// Client可选配置
	option ClientOption
	//当前连接的状态
	isConnected  bool
	connStatusMu sync.RWMutex
	retryTimes   int
	wait         time.Duration
	cancle       context.CancelFunc
	wg           *sync.WaitGroup
}

func NewTCPClient(conn Connetion, opt ClientOption) *TCPClient {
	if conn == nil || opt == nil {
		log.Println("one of client/dp/opt is nil ")
		return nil
	}
	connTimeout := opt.ConnTimeoutOpt(_defConnTimeout)
	conn.SetReadTimeout(connTimeout)
	conn.SetWriteTimeout(connTimeout)
	return &TCPClient{
		// RspMessageChan:    make(chan []byte),
		MessageChanMap: make(map[uint32]chan Messager, 0),
		conn:           conn,
		option:         opt,
		wg:             &sync.WaitGroup{},
	}
}

// IsConnected 是否已连接
func (c *TCPClient) IsConnected() bool {
	c.connStatusMu.RLock()
	is := c.isConnected
	c.connStatusMu.RUnlock()
	return is
}

// OnServerChange 通过此方法实现Notifier
func (c *TCPClient) OnServerChange(addr string) {
	log.Println("server changed!")
	//对于ip+port处于连接状态的，需要先关闭
	if c.IsConnected() {
		c.conn.Close()
	}
	// 原连接失败时会无限重试，此处进行中断
	if c.cancle != nil {
		c.cancle()
		log.Println("wait the last conn for disconnecting totally......")
		c.wg.Wait()
	}
	// 对新ip+port发起TCP连接，并返回取消方法，以便未来取消
	log.Println("wait finish.disconnect from the old server already because of server change")
	ctx, cancel := context.WithCancel(context.Background())
	c.cancle = cancel
	c.wg.Add(1)
	go c.Connect(ctx, addr)
}

// Connect 客户端建立并保持到指定addr的长连接,可接收断开信号主动断开，此时不会自动重连。
// 当连接遇到错误时，会间歇式重试建连。
func (c *TCPClient) Connect(ctx context.Context, addr string) {
	defer c.wg.Done()
	log.Println("Connect start")
	select {
	case <-ctx.Done():
		log.Printf("Connect Stop :%s\n", addr)
		return
	default:
		// 未来定会产生一个connect goroutinue,且必须等待它完成，提前加1
		c.wg.Add(1)
		log.Println("Connect continue,because no cancle")
	}
	c.conn.OnOpen(func() {
		c.connStatusMu.Lock()
		c.isConnected = true
		c.connStatusMu.Unlock()
		log.Printf("%s——>%s Server connect successfully\n", c.conn.LocalAddr(), c.conn.RemoteAddr())
		// 重试次数与等待时间重置
		c.retryTimes = 0
		c.wait = 0
	})

	c.conn.OnError(func(err error) {
		log.Printf("%s——>%s :Server disConnected on error:%s\n", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
		c.retryTimes++
		c.connStatusMu.Lock()
		c.isConnected = false
		c.connStatusMu.Unlock()
		maxReconnectCycle := c.option.MaxReconnectCycleOpt(_defMaxReconnectCycle)
		if c.wait < maxReconnectCycle {
			// 2进制指数退避策略进行重试
			c.wait = time.Duration(int64(math.Exp2(float64(c.retryTimes)))) * time.Second
			if c.wait >= maxReconnectCycle {
				c.wait = maxReconnectCycle
			}
		} else {
			c.wait = maxReconnectCycle
		}
		log.Printf("%s——>%s :Server disconnect, reestablish connection:%s，reconnect times:%d,next try after %s\n", c.conn.LocalAddr(), c.conn.RemoteAddr(), err, c.retryTimes, c.wait)
		time.Sleep(c.wait)
		// 递归调用修改为协程调用。防止递归导致的栈溢出
		go c.Connect(ctx, addr)
	})
	c.conn.OnMessage(func(msg Messager) {
		log.Printf("%s<——%s :onmessage:receive from  server %+v\n", c.conn.LocalAddr(), c.conn.RemoteAddr(), msg)
		msgChan := c.GetOrMakeMsgChan(msg.GetMsgID())
		msgChan <- msg
		log.Printf("recieved msg,id is %d\n", msg.GetMsgID())
	})
	c.conn.Dial(addr)
	log.Printf("finish dial addr:%s\n", addr)
}

// Request 发送请求消息，并获取对应的响应，需提供对应响应的ID
func (c *TCPClient) Request(reqMsg ReqMessager, rspID uint32) (rspMsg RspMessager, err error) {
	RspChan := c.GetOrMakeMsgChan(rspID)
	// 发请求前把消息管道清空，以防收到响应与请求错位
	cleanChan(RspChan)
	c.conn.Write(reqMsg)
	log.Printf("%s——>%s :request:write to server %+v\n", c.conn.LocalAddr(), c.conn.RemoteAddr(), reqMsg)
	// 超时时间内，取出响应数据
	recvTimeout := c.option.RecvTimeoutOpt(_defRecvTimeout)
	timer := time.NewTimer(recvTimeout)
	defer timer.Stop()

	select {
	case msg := <-RspChan:
		return msg, nil
	case <-timer.C:
		return nil, fmt.Errorf("receive message timeout because no msg,current conf:%s", recvTimeout)
	}

}

// AddHandler 监听指定ID的消息，并使用指定的handler处理这些消息
func (c *TCPClient) AddHandler(msgID uint32, handler Handler) {
	msgChan := c.GetOrMakeMsgChan(msgID)
	go func() {
		for msg := range msgChan {
			handler.Serve(msg, c.conn)
		}
	}()
}

// DeleteHandler 关闭删除监听指定ID的handler
func (c *TCPClient) DeleteHandler(msgID uint32) error {
	c.chanMapMu.Lock()
	defer c.chanMapMu.Unlock()
	msgChan, ok := c.MessageChanMap[msgID]
	if !ok || msgChan == nil {
		return errors.New("msgChan is nil or not exist")
	}
	close(msgChan)
	delete(c.MessageChanMap, msgID)
	return nil
}

// GetOrMakeMsgChan 取得对应类型的msgchan,不存在则创建
func (c *TCPClient) GetOrMakeMsgChan(msgID uint32) chan Messager {
	c.chanMapMu.Lock()
	defer c.chanMapMu.Unlock()
	msgChan, ok := c.MessageChanMap[msgID]
	if !ok || msgChan == nil {
		msgChan = make(chan Messager, 10)
		c.MessageChanMap[msgID] = msgChan
	}
	return msgChan
}

// GetConnnection 取得tcpclient中的连接
func (c *TCPClient) GetConnnection() Connetion {
	return c.conn
}

var once sync.Once
var client Client

// GetClient 获取一个单例的Client
func GetClient(opt ClientOption) Client {
	once.Do(
		func() {
			conn := NewTCPConnection()
			client = NewTCPClient(conn, opt)
		})
	return client
}

func cleanChan(c chan Messager) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
