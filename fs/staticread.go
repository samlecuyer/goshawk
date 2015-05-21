package fs

import (
    "bytes"
    "os"
    "time"
)

const (
    NEW_TEMPLATE = `---
title: A New Post
---
# New Post

This is a new post on your blog.
You should probably change the text.
`
)

type StaticRead struct {
    name string
    buf *bytes.Reader
}

func (r *StaticRead) Close() error { return nil }
func (r *StaticRead) Read(p []byte) (n int, err error) { return r.buf.Read(p) }
func (r *StaticRead) ReadAt(p []byte, off int64) (n int, err error) { return r.buf.ReadAt(p, off) }
func (r *StaticRead) Seek(offset int64, whence int) (int64, error) { return r.buf.Seek(offset, whence) }
func (r *StaticRead) Write(p []byte) (n int, err error) { return 0, ErrReadOnly }
func (r *StaticRead) WriteAt(p []byte, off int64) (n int, err error) { return 0, ErrReadOnly }

func (r *StaticRead) Stat() (os.FileInfo, error) {
    return r, nil
}
func (r *StaticRead) Readdir(count int) ([]os.FileInfo, error) {
    return nil, os.ErrInvalid
}
func (r *StaticRead) Readdirnames(n int) ([]string, error) {
    return nil, os.ErrInvalid
}
func (r *StaticRead) WriteString(s string) (ret int, err error) {
    return 0, ErrReadOnly
}
func (r *StaticRead) Truncate(size int64) error {
    return os.ErrInvalid
}
 
func (r *StaticRead) Name() string {
    return r.name
}

// a file is its own stat for now
func (r *StaticRead) Size() int64 {
    return int64(r.buf.Len())
}

func (r *StaticRead) Mode() os.FileMode {
    return 0644 // TODO: we should set this from the data
}

func (r *StaticRead) ModTime() time.Time {
    return time.Now() // TODO: from the data
}

func (r *StaticRead) IsDir() bool {
    return false // TODO: directory files
}

func (r *StaticRead) Sys() interface{} { return nil }
