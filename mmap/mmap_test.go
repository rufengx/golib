package mmap

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestFile_Open(t *testing.T) {
	filename := "mmap_test.go"

	file, err := Open(filename)
	if nil != err {
		t.Error(err)
	}
	fmt.Printf("mmap data: \n%v \n", string(file.data))

	data, err := ioutil.ReadFile(filename)
	if nil != err {
		t.Error(err)
	}
	fmt.Printf("ioutil data: \n%v \n", string(data))
}
