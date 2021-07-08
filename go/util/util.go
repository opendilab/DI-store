package util

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func CreateMemFileFromBytes(b []byte) (string, error) {
	fd, err := unix.MemfdCreate("", 0)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/proc/self/fd/%d", fd)
	err = os.WriteFile(path, b, 0755)
	if err != nil {
		return "", err
	}
	return path, nil
}
