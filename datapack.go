package tcpclient

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const _defMaxPacketSize uint32 = 2 << 26

//DataPack 封包拆包类，是TCP字节流与Messager协议包的转换器，
type DataPack struct {
	MaxPacketSize uint32
}

//封包拆包实例初始化方法
func NewDataPack() *DataPack {
	return &DataPack{
		MaxPacketSize: _defMaxPacketSize,
	}
}

//获取包头长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	//Id uint32(4字节) +  DataLen uint32(4字节)
	return 8
}

//封包方法(压缩数据)
func (dp *DataPack) Pack(msg Messager) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	//写dataLen
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	//写msgID
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgID()); err != nil {
		return nil, err
	}

	//写data数据
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

//拆拆包方法包方法(解压数据)
func (dp *DataPack) Unpack(binaryData []byte) (Messager, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	//读dataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	//读msgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}

	//判断dataLen的长度是否超出我们允许的最大包长度
	if dp.MaxPacketSize > 0 && msg.DataLen > dp.MaxPacketSize {
		return nil, fmt.Errorf("too large msg data received. %d > %d", msg.DataLen, dp.MaxPacketSize)
	}

	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}
