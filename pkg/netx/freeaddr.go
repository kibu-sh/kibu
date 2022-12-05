package netx

import (
	"fmt"
	"net"
)

type Address interface {
	Network() string
	Host() string
	Port() string
	String() string
	HostPort() string
}

type TCPAddress struct {
	*net.TCPAddr
}

func (T TCPAddress) HostPort() string {
	return fmt.Sprintf("%s:%s", T.Host(), T.Port())
}

func (T TCPAddress) Host() string {
	return T.TCPAddr.IP.String()
}

func (T TCPAddress) Port() string {
	return fmt.Sprintf("%d", T.TCPAddr.Port)
}

func (T TCPAddress) String() string {
	return fmt.Sprintf("%s://%s", T.Network(), T.HostPort())
}

func (T TCPAddress) Underlying() net.Addr {
	return T.TCPAddr
}

// GetFreeAddr asks the kernel for a free open port that is ready to use.
func GetFreeAddr() (Address, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	defer l.Close()
	return TCPAddress{l.Addr().(*net.TCPAddr)}, nil
}
