package proxy

import (
	"log"
	"io"
	"net"
	"net/netip"
	"slices"

	"github.com/shiro8613/minecraft-a-proxy/packet"
	"github.com/shiro8613/minecraft-a-proxy/config"
	"golang.org/x/sync/errgroup"
)

type ProxyServer struct {}

func NewServer() *ProxyServer {
	return &ProxyServer{}
}

func (s *ProxyServer) Start(addr *net.TCPAddr) error {
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	defer l.Close()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			return err
		}

		go s.handler(conn)
	}
}

func (s *ProxyServer) handler(c *net.TCPConn) {
	defer c.Close()

	p := &Proxy{}
	err := p.Start(c)
	if err != nil && err != io.EOF {
		log.Printf("[error] %s\n", err)
	}
}

type Proxy struct {}

func (p *Proxy) Start(cConn *net.TCPConn) error {
	defer cConn.Close()

	var eg errgroup.Group
	var sConn *net.TCPConn
	addr := cConn.RemoteAddr()
	
	eg.Go(func() error {
		buff := make([]byte, 0xFFFF)
		for {
			n, err := cConn.Read(buff)
			if err != nil {
				return err
			}
			b := buff[:n]
			if 0 < len(b) {
				if sConn == nil {
					b1 := slices.Clone(b)
					p := &packet.HelloPacket{}
					r, err := p.Read(b1)
					if err != nil {
						return err
					}
					
					if r {
						server, ok := config.GetConfig().servers[p.Hostname]
						if !ok {
							return cConn.Close()
						}

						serverIP := net.TCPAddrFromAddrPort(netip.MustParseAddrPort(server))
						sConn, err = net.DialTCP("tcp", nil, serverIP)
						if err != nil {
							return err
						}
					}
				}

				if sConn != nil && addr != nil {
					b1 := slices.Clone(b)
					p := &packet.LoginPacket{}
					r, err := p.Read(b1)
					if err != nil && err != io.EOF {
						return err
					}

					if r && 3 < p.Length {
						log.Printf("player is connected [%s]%s(%s)", addr.String(), p.Name, p.Uuid)
						addr = nil
					}
				}
			}

			if sConn != nil {
				n, err = sConn.Write(b)
				if err != nil {
					return err
				}
			}
		}
	})
	
	eg.Go(func() error { 
		buff := make([]byte, 0xFFFF)
		for {
			if sConn != nil {
				n,err := sConn.Read(buff)
				if err != nil {
					return err
				}
	
				b := buff[:n]
	
				n, err = cConn.Write(b)
				if err != nil {
					return err
				}
			}
		}
	})

	err := eg.Wait()
	if sConn != nil {
	  sConn.Close()
	}

	return err
}
