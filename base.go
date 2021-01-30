package tcp_client

import (
	"context"
	"io"
	"net"
	"time"
)

// Connecter抽象连接器，拥有面向连接的交互方法: Request
type Connecter interface {
	IsConnected() bool
	Request(reqMsg ReqMessager) (RspMessager, error)
	Connect(context.Context, string)
	OnServerChange(addr string)
}
type ReqMessager interface {
	Messager
}
type RspMessager interface {
	Messager
}
type Messager interface {
	GetDataLen() uint32
	//获取消息ID
	GetMsgID() uint32
	//获取消息内容
	GetData() []byte
	//设置消息数据段长度
	SetDataLen(len uint32)
	//设置消息ID
	SetMsgID(msgID uint32)
	//设置消息内容
	SetData(data []byte)
}

// DataPacker 是一个抽象打包器，可将Messager持有的消息打包、拆包
type DataPacker interface {
	//获取包头长度方法
	GetHeadLen() uint32
	//封包方法(压缩数据)
	Pack(msg Messager) ([]byte, error)
	//拆拆包方法包方法(解压数据)
	Unpack(binaryData []byte) (Messager, error)
}

type Client interface {
	OnOpen(f func())
	OnMessage(f func([]byte))
	OnError(f func(error))
	Listen(addr string)
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetReadTimeout(readTimeout time.Duration)
	io.WriteCloser
	SetWriteTimeout(writeTimeout time.Duration)
}
