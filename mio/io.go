package mio

import (
	"net"
)

func WriteAll(conn net.Conn, buf []byte) error {
	//log.Print("写会浏览器")
	for writtenNum := 0; writtenNum != len(buf); {
		tempNum, err := conn.Write(buf[writtenNum:])
		if err != nil {
			return err
		}
		writtenNum += tempNum
	}
	return nil
}
