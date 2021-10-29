package util

import (
	"encoding/binary"
	"io"
	"net"
	"unsafe"
)

type BytesReadWriter struct {
	conn net.Conn
}

func NewBytesReadWriter(conn net.Conn) *BytesReadWriter {
	return &BytesReadWriter{
		conn: conn,
	}
}

func (rw *BytesReadWriter) Write(p []byte) (int, error) {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(p)))
	n1, err := rw.conn.Write(bs)
	if err != nil {
		return n1, err
	}
	n2, err := rw.conn.Write(p)
	return n1 + n2, err
}

func (rw *BytesReadWriter) Read(buff []byte) ([]byte, error) {
	bs := make([]byte, 4)
	_, err := io.ReadFull(rw.conn, bs)
	if err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint32(bs)

	var data []byte
	if buff == nil {
		data = make([]byte, size)
	} else {
		if int(size) != len(buff) {
			panic("buff size mismatched")
		}
		data = buff
	}

	if size == 0 {
		return data, nil
	}
	_, err = io.ReadFull(rw.conn, data)
	return data, err
}

func (rw *BytesReadWriter) Close() error {
	return rw.conn.Close()
}

func BytesWithoutCopy(p unsafe.Pointer, size int) []byte {
	return (*[1 << 28]byte)(p)[:size:size]
}
