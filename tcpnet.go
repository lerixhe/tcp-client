package tcpclient

import (
	"context"
	"time"
)

// Client
type Client interface {
	IsConnected() bool
	// 发送请求消息，并获取对应的响应，需提供对应响应的ID
	Request(reqMsg ReqMessager, rspID uint32) (rspMsg RspMessager, err error)
	// AddHandler 监听指定ID的消息，并使用指定的handler处理这些消息
	AddHandler(msgID uint32, handler Handler)

	Connect(context.Context, string)
	OnServerChange(addr string)
}

type Handler interface {
	Serve(ReqMessager, RspMessagerWriter)
}
type ReqMessager interface {
	Messager
}
type RspMessager interface {
	Messager
}
type RspMessagerWriter interface {
	Write(msg Messager) (int, error)
}

type Messager interface {
	GetDataLen() uint32
	//获取消息ID
	GetMsgID() uint32
	//获取消息内容
	GetData() []byte
	//设置消息数据段长度
	SetDataLen(len uint32)
	//设计消息ID
	SetMsgID(msgID uint32)
	//设计消息内容
	SetData(data []byte)
}

// DataPacker 是一个抽象打包器，将TCP字节流进行打包、拆包。
type DataPacker interface {
	//获取包头长度方法
	GetHeadLen() uint32
	//封包方法
	Pack(msg Messager) ([]byte, error)
	//拆包方法
	Unpack(binaryData []byte) (Messager, error)
}

type Connetion interface {
	// hook函数
	OnOpen(f func())
	OnMessage(f func(Messager))
	OnError(f func(error))

	Dial(addr string)
	Write(msg Messager) (int, error)
	Close() error
	SetReadTimeout(readTimeout time.Duration)
	SetWriteTimeout(writeTimeout time.Duration)

	RemoteAddr() string
	LocalAddr() string
}
