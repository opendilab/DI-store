package util

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"unsafe"
)

type BytesReadWriter struct {
	_rw  *bufio.ReadWriter
	conn net.Conn
}

func NewBytesReadWriter(conn net.Conn) *BytesReadWriter {
	return &BytesReadWriter{
		_rw:  bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		conn: conn,
	}
}

func (rw *BytesReadWriter) Write(p []byte) (int, error) {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(p)))
	n1, err := rw._rw.Write(bs)
	if err != nil {
		return n1, err
	}
	n2, err := rw._rw.Write(p)
	if err != nil {
		return n1 + n2, err
	}
	err = rw._rw.Flush()
	return n1 + n2, err
}

func (rw *BytesReadWriter) Read(buff []byte) ([]byte, error) {
	bs := make([]byte, 4)
	_, err := io.ReadFull(rw._rw, bs)
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
	_, err = io.ReadFull(rw._rw, data)
	return data, err
}

func (rw *BytesReadWriter) Close() error {
	return rw.conn.Close()
}

func BytesWithoutCopy(p unsafe.Pointer, size int) []byte {
	return (*[1 << 28]byte)(p)[:size:size]
}
