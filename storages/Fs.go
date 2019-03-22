package storages

import (
	"io"
	"os"
)

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	io.WriterAt
	Name() string
	Readdir(count int) ([]os.FileInfo, error)
	Readdirnames(n int) ([]string, error)
	Stat() (os.FileInfo, error)
	Sync() error
	Truncate(size int64) error
	WriteString(s string) (ret int, err error)
}

type Fs interface {
	Open(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
}

type OsFs struct {}

func (fs *OsFs) Open(name string) (File, error) {
	return os.Open(name)
}

func (fs *OsFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func NewOsFs() *OsFs {
	return &OsFs{}
}
