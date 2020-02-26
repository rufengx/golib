package sendfile

import (
	"fmt"
	"net"
	"os"
)

func Sendfile(tcpConn *net.TCPConn, file *os.File) (int64, error) {
	fstat, err := file.Stat()
	if nil != err {
		return 0, err
	}

	size := fstat.Size()
	if size < 0 {
		return 0, fmt.Errorf("file size is negative, filename: %v", file.Name())
	}

	if size != int64(int(size)) {
		return 0, fmt.Errorf("file size too large, filename: %v", file.Name())
	}

	// ReadFrom use syscall splice and sendfile to send data.
	return tcpConn.ReadFrom(file)
}
