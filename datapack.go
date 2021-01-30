package tcp_client

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// DataPack 将Messager数据打包成二进制数据
type DataPack struct {
	MaxPacketsize uint32
}

//NewDataPack 封包拆包实例初始化方法
func NewDataPack(maxPacketSize uint32) *DataPack {
	return &DataPack{
		MaxPacketsize: maxPacketSize}
}

//GetHeadLen 获取包头长度的方法
func (dp *DataPack) GetHeadLen() uint32 {
	//Id uint32(4字节) + Datalen uint32(4字节)
	return 8
}

//Pack 封包方法(压缩数据)
func (dp *DataPack) Pack(msg Messager) ([]byte, error) {
	// 创建1个存放bytes字节的緩冲
	dataBuff := bytes.NewBuffer([]byte{})
	// dataLen
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

//Unpack 拆包方法(解压数据)
func (dp *DataPack) Unpack(binaryData []byte) (Messager, error) {
	// 创建一个从输入二进制数据读取数据的ioReader
	dataBuff := bytes.NewReader(binaryData)
	msg := new(Message)

	// 读dataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}
	// 读msgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}
	// 判断dataLen是否超出最大包长度
	if dp.MaxPacketsize > 0 && msg.DataLen > dp.MaxPacketsize {
		return nil, errors.New("too large msg data received")
	}

	// 读data
	data := make([]byte, msg.DataLen)
	if err := binary.Read(dataBuff, binary.LittleEndian, &data); err != nil {
		return nil, err
	}
	msg.Data = data
	return msg, nil
}
