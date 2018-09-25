package client

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type Config struct {
	Tls    *tls.Config
	Dialer *net.Dialer
}

var (
	ErrFileNotFound       = errors.New("Server reported that the file did not exist")
	ErrClientIncompatible = errors.New("Server reported that the client is incompatible")
	ErrUnknownError       = errors.New("Server reported an error we didn't know existed")
)

type readcloser struct {
	r io.Reader
	c io.Closer
}

func (r *readcloser) Close() error {
	return r.Close()
}

func (r *readcloser) Read(b []byte) (int, error) {
	return r.r.Read(b)
}

func Do(network, address, file string, cfg *Config) (e error, r io.ReadCloser) {
	var c net.Conn

	if cfg.Tls == nil {
		// Non-tls
		if cfg.Dialer != nil {
			c, e = cfg.Dialer.Dial(network, address)
		} else {
			c, e = net.Dial(network, address)
		}
	} else {
		// tls
		if cfg.Dialer != nil {
			c, e = tls.DialWithDialer(cfg.Dialer, network, address, cfg.Tls)
		} else {
			c, e = tls.Dial(network, address, cfg.Tls)
		}
	}

	if e != nil {
		return
	}

	w := bufio.NewWriter(c)
	cVrFNLlen := binary.MaxVarintLen16 + binary.MaxVarintLen32
	cVrFNL := make([]byte, cVrFNLlen, cVrFNLlen)

	written := binary.PutUvarint(cVrFNL, 0)
	written += binary.PutUvarint(cVrFNL[written:], uint64(len(file)))

	_, e = w.Write(cVrFNL[:written])
	if e != nil {
		c.Close()
		return
	}

	_, e = w.WriteString(file)
	if e != nil {
		c.Close()
		return
	}

	br := bufio.NewReader(c)
	lenErr, e := binary.ReadVarint(br)
	if e != nil {
		c.Close()
		return
	}

	if lenErr < 0 {
		switch lenErr {
		case -1:
			e = ErrFileNotFound
		case -2:
			e = ErrClientIncompatible
		default:
			e = ErrUnknownError
		}
		c.Close()
		return
	}

	return nil, &readcloser{
		io.LimitReader(r, lenErr),
		c,
	}
}
