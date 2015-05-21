package fs

import (
	"bytes"
	"fmt"
	"github.com/spf13/hugo/parser"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"time"
)

type WriteFile struct {
	name string
	mode os.FileMode
	q    *mgo.Collection
	buf  *bytes.Buffer
}

func NewWriteFile(name string, flag int, mode os.FileMode, q *mgo.Collection) *WriteFile {
	buf := new(bytes.Buffer)
	return &WriteFile{
		name,
		mode,
		q,
		buf,
	}
}

func (r *WriteFile) WriteString(s string) (ret int, err error) {
	return r.buf.WriteString(s)
}

func (r *WriteFile) Write(p []byte) (n int, err error) {
	return r.buf.Write(p)
}

// okay, this is where the magic happens
func (r *WriteFile) Close() error {
	raw := r.buf.Bytes()
	rawr := bytes.NewReader(raw)
	page, err := parser.ReadFrom(rawr)
	if err != nil {
		return err
	}
	meta, err := page.Metadata()
	if err != nil {
		return err
	}
	metamap := meta.(map[string]interface{})
	post := FromMap(r.name, metamap)
	post.Raw = string(raw)
	post.Body = string(page.Content())
	fmt.Printf("closing: %v\n\n", post)
	_, err = r.q.Upsert(bson.M{"slug": post.Slug}, post)
	return err
}

func (r *WriteFile) Stat() (os.FileInfo, error) { return r, nil }
func (r *WriteFile) Truncate(size int64) error {
	r.buf.Truncate(int(size))
	return nil
}

func (r *WriteFile) Name() string {
	return r.name
}

// a file is its own stat for now
func (r *WriteFile) Size() int64 {
	return int64(r.buf.Len())
}

func (r *WriteFile) Read(p []byte) (n int, err error)               { return 0, os.ErrInvalid }
func (r *WriteFile) ReadAt(p []byte, off int64) (n int, err error)  { return 0, os.ErrInvalid }
func (r *WriteFile) Seek(offset int64, whence int) (int64, error)   { return 0, os.ErrInvalid }
func (r *WriteFile) WriteAt(p []byte, off int64) (n int, err error) { return 0, os.ErrInvalid }
func (r *WriteFile) Readdir(count int) ([]os.FileInfo, error)       { return nil, os.ErrInvalid }
func (r *WriteFile) Readdirnames(n int) ([]string, error)           { return nil, os.ErrInvalid }

func (r *WriteFile) Mode() os.FileMode {
	return r.mode
}
func (r *WriteFile) ModTime() time.Time {
	return time.Now() // TODO: from the data
}
func (r *WriteFile) IsDir() bool {
	return r.mode.IsDir()
}
func (r *WriteFile) Sys() interface{} { return nil }
