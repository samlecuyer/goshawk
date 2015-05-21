package scp

import (
	"fmt"
	"github.com/samlecuyer/goshawk/fs"
	"github.com/spf13/afero"
	"io"
	"os"
	"path"
)

type ScpReader struct {
	c  io.ReadWriter // this is actually a channel
	fs FsProxy       // for testing, this should be a muxer
}

func (s *ScpReader) Source(f *flags) {
	for _, name := range f.targs {
		ofd, err := s.fs.Fs().Open(name)
		if err != nil {
			s.run_err("%s: %s", name, err.Error())
			return
		}
		defer ofd.Close()
		stat, err := ofd.Stat()
		if err != nil {
			s.run_err("%s: %s", name, err.Error())
			return
		}
		mode := stat.Mode()
		switch {
		case mode.IsRegular():
			{
				break
			}
		case mode.IsDir() && f.r:
			{
				// s.run_err("yeah, we didn't implement this yet")
				s.rsource(ofd, f)
				continue
			}
		default:
			s.run_err("%s: is not a regular file", name)
			continue
		}
		size := stat.Size()
		// send the first line
		_, fname := path.Split(name)
		buf := fmt.Sprintf("C%04o %d %s\n", mode, size, fname)
		fmt.Println("Sending file modes: ", buf)
		s.c.Write([]byte(buf))
		s.response()
		// send the actual file
		i, err := io.CopyN(s.c, ofd, size)
		// check for response it
		if i != 1 && err != nil {
			s.fatal("could not write successfully")
		} else {
			s.okay()
		}
		s.response()
	}
}

func (s *ScpReader) rsource(dir afero.File, f *flags) {
	names, err := dir.Readdirnames(-1)
	if err != nil {
		s.run_err("couldn't read dir: %s", err.Error())
		return
	}
	name := dir.Name()
	pname, fname := path.Split(name)
	fmt.Printf("(%s)(%s)", pname, fname)
	stat, err := dir.Stat()
	mode := stat.Mode() & os.ModePerm
	fmt.Fprintf(s.c, "D%04o %d %s\n", mode, 0, fname)
	if s.response() != nil {
		return
	}
	nflags := &flags{f.f, f.t, f.r, f.d, f.v, make([]string, len(names))}
	for i, fname := range names {
		nflags.targs[i] = path.Join(name, fname)
	}
	s.Source(nflags)
	fmt.Fprint(s.c, "E\n")
	s.response()
}

func (s *ScpReader) Sink(f *flags) {
	var mode os.FileMode
	var size int64
	// acknowledge the req
	s.okay()
	var ofd afero.File
	//for first := true;; first = false {
	for {
		// allocate some space to read
		resp := make([]byte, 2048)
		// read the first line
		i, err := s.c.Read(resp)
		if err != nil {
			return
		}
		control := string(resp[:i])
		switch control[0] {
		case 'E':
			{
				s.okay()
				return
			}
		case 'C':
			{
				name := f.targs[0]
				fmt.Sscanf(control, "C%4o %10d %s\n", &mode, &size)
				ofd, err = s.fs.Fs().OpenFile(name, fs.OPEN_WRITE_MODE, mode)
				if err != nil {
					s.run_err(err.Error())
					continue
				}
			}
		default:
			s.run_err(err.Error())
			continue
		}
		fmt.Println(control)
		// parse the first line
		// acknowledge it
		s.okay()
		// actually read the data
		io.CopyN(ofd, s.c, size)
		ofd.Close()
		s.okay()

		// acknowledge it
		err = s.response()
		if err != nil {
			s.fatal(err.Error())
		} else {
			s.okay()
		}
	}
}

func (s *ScpReader) response() error {
	buf := make([]byte, 1)
	s.c.Read(buf)
	if buf[0] != 0x00 {
		buf = make([]byte, 1024)
		s.c.Read(buf)
		fmt.Printf("ERROR in response: `%s`", buf)
		return fmt.Errorf("error reading: %s", buf)
	}
	return nil
}

func (s *ScpReader) okay() {
	s.c.Write([]byte{0x00})
}

func (s *ScpReader) run_err(fmts string, v ...interface{}) {
	fmt.Printf(fmts, v)
	s.c.Write([]byte{0x01})
	fmt.Fprintf(s.c, "scp: ")
	fmt.Fprintf(s.c, fmts, v)
	s.c.Write([]byte("\n"))
}

func (s *ScpReader) fatal(msg string) {
	s.c.Write([]byte{0x02})
	s.c.Write([]byte(msg))
	s.c.Write([]byte("\n"))
}
