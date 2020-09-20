package sendfile

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestSendfile(t *testing.T) {
	file, err := os.Open("sendfile.go")
	if nil != err {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)

	// xio listen
	go func() {
		defer wg.Done()
		listen, err := net.Listen("tcp", "127.0.0.1:9999")
		if nil != err {
			panic(err)
		}

		// just do test, only accept one request.
		conn, err := listen.Accept()
		if nil != err {
			panic(err)
		}

		Sendfile(conn.(*net.TCPConn), file)
	}()

	// client dial
	go func() {
		time.Sleep(3 * time.Second)
		defer wg.Done()
		conn, err := net.Dial("tcp", "127.0.0.1:9999")
		if nil != err {
			t.Error(err)
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		var content string
		for i := 1; i <= 10; i++ {
			line, err := reader.ReadString(byte('\n'))
			if err != nil {
				t.Error(err)
			}
			content = content + line
		}
		fmt.Printf("receive data:\n%s \n...\n", content)
	}()
	wg.Wait()
}
