package pkg

import (
	"encoding/binary"
	"io"
	"net"
)

func SendData(conn net.Conn, data []byte) error {
	// 计算消息长度并将其转换成4个字节的二进制数据
	length := uint32(len(data))
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, length)

	// 发送消息头部和消息内容
	_, err := conn.Write(append(header, data...))
	return err
}

func RecvData(conn net.Conn) ([]byte, error) {
	// 先读取4个字节的消息头部，获取消息长度
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header)

	// 根据消息长度读取对应的消息内容
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}
	return data, nil
}
