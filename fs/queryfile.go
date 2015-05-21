package fs

import (
	"bytes"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"time"
)

type QueryFile struct {
	name string
	mode os.FileMode
	q    *mgo.Collection
	buf  *bytes.Buffer
}

func NewQueryFile(name string, flag int, mode os.FileMode, q *mgo.Collection) *QueryFile {
	buf := new(bytes.Buffer)
	return &QueryFile{
		name,
		mode,
		q,
		buf,
	}
}

func (r *QueryFile) Close() error {
	return nil
}
func (r *QueryFile) Read(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}
func (r *QueryFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}
func (r *QueryFile) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}
func (r *QueryFile) Write(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}
func (r *QueryFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}
func (r *QueryFile) Stat() (os.FileInfo, error) {
	return r, nil
}
func (r *QueryFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
func (r *QueryFile) Readdirnames(n int) ([]string, error) {
	if r.mode.IsDir() {
		query := r.q.Find(nil).Select(bson.M{"slug": 1})
		if n > -1 {
			query.Limit(n)
		}
		var result []struct {
			Slug string `json:"slug"`
		}
		err := query.All(&result)
		if err != nil {
			return nil, err
		}
		names := make([]string, len(result))
		for i, name := range result {
			names[i] = name.Slug
		}
		return names, nil
	}
	return nil, os.ErrInvalid
}
func (r *QueryFile) WriteString(s string) (ret int, err error) {
	return 0, os.ErrInvalid
}
func (r *QueryFile) Truncate(size int64) error {
	return os.ErrInvalid
}

func (r *QueryFile) Name() string {
	return r.name
}

// a file is its own stat for now
func (r *QueryFile) Size() int64 {
	return 0
}
func (r *QueryFile) Mode() os.FileMode {
	return r.mode
}
func (r *QueryFile) ModTime() time.Time {
	return time.Now() // TODO: from the data
}
func (r *QueryFile) IsDir() bool {
	return r.mode.IsDir()
}
func (r *QueryFile) Sys() interface{} { return nil }
