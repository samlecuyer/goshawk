package scp

import (
	"flag"
	"fmt"
	"github.com/samlecuyer/goshawk/fs"
	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
	"net"
	"strings"
)

type ScpServer struct {
	l  net.Listener
	c  *ssh.ServerConfig
	fs afero.Fs
}

type FsProxy interface {
	Fs() afero.Fs
}

func NewServer(listener net.Listener, conf *ssh.ServerConfig) *ScpServer {
	fs := fs.NewMongoBlog()
	return &ScpServer{listener, conf, fs}
}

func (ss *ScpServer) Fs() afero.Fs {
	return ss.fs
}

func (ss *ScpServer) ListenAndServe() {
	for {
		conn, err := ss.l.Accept()
		if err != nil {
			panic("failed to accept incoming connection")
		}
		go ss.handleConn(conn)
	}
}

func (ss *ScpServer) handleConn(conn net.Conn) {
	_, chans, reqs, err := ssh.NewServerConn(conn, ss.c)
	if err != nil {
		panic("failed to handshake")
	}
	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		ss.acceptScp(newChannel)
	}
}

func (ss *ScpServer) acceptScp(newChannel ssh.NewChannel) {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		panic("could not accept channel.")
	}

	go func(in <-chan *ssh.Request) {
		defer channel.CloseWrite()
		defer channel.Close()
		for req := range in {
			switch req.Type {
			case "exec":
				flags, err := parseRequest(string(req.Payload))
				if err != nil {
					fmt.Println(err)
					req.Reply(false, nil)
					break
				}
				req.Reply(true, nil)
				// we have to write an acknowledgement
				sr := &ScpReader{channel, ss}
				if flags.t {
					sr.Sink(flags)
				} else if flags.f {
					sr.response()
					sr.Source(flags)
				}
				return
			default:
				fmt.Printf("unknown: %s\n", req.Type)
				// we should handle env for utf8?
				req.Reply(false, nil)
			}
		}
	}(requests)
}

type flags struct {
	f     bool
	t     bool
	r     bool
	d     bool
	v     bool
	targs []string
}

func parseRequest(cmd string) (*flags, error) {
	parts := strings.Split(cmd, " ")
	pz_l := len(parts[0])
	pz := strings.TrimSpace(parts[0][pz_l-3:])
	if pz != "scp" {
		return nil, fmt.Errorf("exec must be 'scp', not '%x', %v", pz, len(pz))
	}
	flags := new(flags)
	fs := flag.NewFlagSet("scp", flag.ContinueOnError)
	fs.BoolVar(&flags.f, "f", false, "source (f for from)")
	fs.BoolVar(&flags.t, "t", false, "sink (t for to)")
	fs.BoolVar(&flags.r, "r", false, "recursive")
	fs.BoolVar(&flags.d, "d", false, "directory")
	fs.BoolVar(&flags.v, "v", false, "verbose")
	err := fs.Parse(parts[1:])
	if err != nil {
		return nil, err
	}
	flags.targs = fs.Args()
	if flags.t && flags.f {
		return nil, fmt.Errorf("cannot be both sink & source")
	}
	return flags, nil
}
