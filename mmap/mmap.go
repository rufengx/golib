package mmap

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

type File struct {
	FileInfo os.FileInfo
	data     []byte
}

func Open(filename string) (*File, error) {
	file, err := os.Open(filename)
	if nil != err {
		return nil, err
	}

	fstat, err := file.Stat()
	if nil != err {
		return nil, err
	}

	size := fstat.Size()
	if size < 0 {
		return nil, fmt.Errorf("mmap: file %q has negative size", file.Name())
	}

	if size != int64(int(size)) {
		return nil, fmt.Errorf("mmap: file %q is too large", file.Name())
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if nil != err {
		return nil, err
	}

	f := &File{
		FileInfo: fstat,
		data:     data,
	}

	// help gc
	runtime.SetFinalizer(f, (*File).Close)
	return f, nil
}

func (f *File) Len() int {
	return len(f.data)
}

func (f *File) Close() error {
	if nil == f || nil == f.data {
		return nil
	}

	data := f.data
	f.data = nil

	// help gc
	runtime.SetFinalizer(f, nil)
	return syscall.Munmap(data)
}
