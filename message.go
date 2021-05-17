package tcpclient

// Message 与agent交互的message格式
type Message struct {
	DataLen uint32 //消息体长度
	ID      uint32 //消息ID
	Data    []byte //消息体（protobuf）
}

//NewMsgPackage 创建一个Message消息包
func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		DataLen: uint32(len(data)),
		ID:      id,
		Data:    data,
	}
}

//GetDataLen 获取消息数据段长度
func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

//GetMsgID 获取消息ID
func (msg *Message) GetMsgID() uint32 {
	return msg.ID
}

//GetData 获取消息内容
func (msg *Message) GetData() []byte {
	return msg.Data
}

//SetDataLen 设置消息数据段长度
func (msg *Message) SetDataLen(len uint32) {
	msg.DataLen = len
}

//SetMsgID 设置消息ID
func (msg *Message) SetMsgID(msgID uint32) {
	msg.ID = msgID
}

//SetData 设置消息内容
func (msg *Message) SetData(data []byte) {
	msg.Data = data
}
