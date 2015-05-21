package fs

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"path"
	"time"
)

type MongoBlog struct {
	session *mgo.Session
}

var ErrReadOnly = errors.New("this supports read only")
var ErrNotImplemented = errors.New("yeah, we didn't implement this")
var ErrIsDirectory = errors.New("this supports write only")

var OPEN_WRITE_MODE = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

func NewMongoBlog() *MongoBlog {
	session, err := mgo.Dial("localhost")
	if err != nil {
		fmt.Printf("error connecting to mongo: %s", err.Error())
		return nil
	}
	fmt.Println("now connected")
	return &MongoBlog{session}
}

func (mb *MongoBlog) Open(name string) (afero.File, error) {
	return mb.OpenFile(name, os.O_RDONLY, 0644)
}

func (mb *MongoBlog) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	switch flag {
	case os.O_RDONLY:
		{
			return mb.openReadFile(name, flag, perm)
		}
	case OPEN_WRITE_MODE:
		{
			return mb.openWriteFile(name, flag, perm)
		}
	default:
		return nil, afero.ErrFileNotFound
	}
}

func (mb *MongoBlog) openReadFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	posts := mb.session.DB("blog").C("posts")
	switch name {
	case "/posts/":
		{
			return NewQueryFile(name, flag, os.ModeDir|0755, posts), nil
		}
	case "/posts/new":
		{
			return &StaticRead{
				"new",
				bytes.NewReader([]byte(NEW_TEMPLATE)),
			}, nil
		}
	default:
		pname, fname := path.Split(name)
		if pname == "/posts/" {
			post := new(Post)
			err := posts.Find(bson.M{"slug": fname}).One(post)
			if err == nil {
				return &StaticRead{
					fname,
					bytes.NewReader([]byte(post.Raw)),
				}, nil
			}
		}
		return nil, afero.ErrFileNotFound
	}
}

func (mb *MongoBlog) openWriteFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	posts := mb.session.DB("blog").C("posts")
	switch name {
	case "/posts/new":
		{
			return NewWriteFile(name, flag, perm, posts), nil
		}
	default:
		pname, _ := path.Split(name)
		if pname == "/posts/" {
			return NewWriteFile(name, flag, perm, posts), nil
		}
		return nil, afero.ErrFileNotFound
	}
}

func (mb *MongoBlog) Create(name string) (afero.File, error)       { return nil, nil }
func (mb *MongoBlog) Mkdir(path string, perm os.FileMode) error    { return nil }
func (mb *MongoBlog) MkdirAll(path string, perm os.FileMode) error { return nil }
func (mb *MongoBlog) Remove(name string) error                     { return afero.ErrFileNotFound }
func (mb *MongoBlog) RemoveAll(path string) error                  { return afero.ErrFileNotFound }
func (mb *MongoBlog) Rename(old, new string) error                 { return afero.ErrFileNotFound }
func (mb *MongoBlog) Stat(name string) (os.FileInfo, error)        { return nil, afero.ErrFileNotFound }
func (mb *MongoBlog) Name() string                                 { return "goshawk" }
func (mb *MongoBlog) Chmod(name string, mode os.FileMode) error    { return afero.ErrFileNotFound }
func (mb *MongoBlog) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return afero.ErrFileNotFound
}
