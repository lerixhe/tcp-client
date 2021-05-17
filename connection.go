package tcpclient

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	_defReadTimeout  = 180 * time.Second
	_defWriteTimeout = 180 * time.Second
)

var _ Connetion = (*TCPConnection)(nil)

// TCPConnection is a long connection that implements the Connectionn interface
type TCPConnection struct {
	// 对TCP字节流的 拆封包协议，防止粘包、分包的问题
	packer            DataPacker
	onOpenCallback    func()
	onMessageCallback func(msg Messager)
	onErrorCallback   func(err error)
	Conn              net.Conn
	readTimeout       time.Duration
	writeTimeout      time.Duration
	Address           string
}

func (c *TCPConnection) OnOpen(f func()) {
	c.onOpenCallback = f
}

func (c *TCPConnection) OnMessage(f func(msg Messager)) {
	c.onMessageCallback = f
}

func (c *TCPConnection) OnError(f func(err error)) {
	c.onErrorCallback = f
}

func (c *TCPConnection) Close() error {
	return c.Conn.Close()
}
func (c *TCPConnection) RemoteAddr() string {
	if c.Conn != nil {
		return c.Conn.RemoteAddr().String()
	}
	return c.Address
}
func (c *TCPConnection) LocalAddr() string {
	if c.Conn != nil {
		return c.Conn.LocalAddr().String()
	}
	return "localAddr"
}
func (c *TCPConnection) SetReadTimeout(readTimeout time.Duration) {
	c.readTimeout = readTimeout
}
func (c *TCPConnection) SetWriteTimeout(writeTimeout time.Duration) {
	c.writeTimeout = writeTimeout
}
func (c *TCPConnection) Write(msg Messager) (int, error) {
	var n int
	var err error
	//将data封包，并且发送
	message, err := c.packer.Pack(msg)
	if err != nil {
		return 0, errors.New("c.packer.Pack error")
	}

	c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	n, err = c.Conn.Write(message)
	if err != nil {
		err = fmt.Errorf("TCPConn Write:%w", err)
		c.Close()
		c.onErrorCallback(err)
		return n, err
	}
	return n, nil
}

func (c *TCPConnection) Dial(addr string) {
	c.Address = addr
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		c.Conn = nil
		c.onErrorCallback(fmt.Errorf("net.DialTimeout:%w", err))
		return
	} else {
		defer conn.Close()
		//disable tcp keepalive
		conn.(*net.TCPConn).SetKeepAlive(false)
		c.Conn = conn
		c.onOpenCallback()
		c.read()
	}
}

func (c *TCPConnection) read() {
	for {
		// Read the header first when data flows in,
		// Then read the message body according to the length in the header.
		// conn will disconnect if any error occor
		c.Conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		header := make([]byte, c.packer.GetHeadLen())
		_, err := io.ReadFull(c.Conn, header)
		if err != nil {
			c.Close()
			c.onErrorCallback(fmt.Errorf("io.ReadFull(c.Conn, header):%w", err))
			return
		}
		msgWithoutBody, err := c.packer.Unpack(header)
		if err != nil {
			c.Close()
			c.onErrorCallback(fmt.Errorf("c.packer.Unpack:%w", err))
			return
		}
		var body []byte
		if bodyLen := msgWithoutBody.GetDataLen(); bodyLen > 0 {
			body = make([]byte, bodyLen)
			if _, err := io.ReadFull(c.Conn, body); err != nil {
				c.Close()
				c.onErrorCallback(fmt.Errorf("io.ReadFull(c.Conn, body):%w", err))
				return
			}
		}
		msgWithoutBody.SetData(body)
		msg := msgWithoutBody
		c.onMessageCallback(msg)
	}
}

func NewTCPConnection() *TCPConnection {
	conn := &TCPConnection{
		packer:       NewDataPack(),
		readTimeout:  _defReadTimeout,
		writeTimeout: _defWriteTimeout,
	}
	conn.OnOpen(func() {})
	conn.OnError(func(err error) {})
	conn.OnMessage(func(msg Messager) {})

	return conn
}
