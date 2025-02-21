package proxy

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"slices"

	"github.com/shiro8613/minecraft-a-proxy/config"
	"github.com/shiro8613/minecraft-a-proxy/packet"

	"golang.org/x/sync/errgroup"
)

type ProxyServer struct {}

func NewServer() *ProxyServer {
	return &ProxyServer{}
}

func (s *ProxyServer) Start(ctx context.Context, addr *net.TCPAddr) error {
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	defer l.Close()

	go func () {
		<- ctx.Done()
		if err := l.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}

		go s.handler(ctx, conn)
	}
}

func (s *ProxyServer) handler(ctx context.Context, c *net.TCPConn) {
	defer c.Close()

	p := &Proxy{}
	err := p.Start(ctx, c)
	if err != nil && err != io.EOF {
		log.Printf("[error] %s\n", err)
	}
}

type Proxy struct {}

func (p *Proxy) Start(ctx context.Context, cConn *net.TCPConn) error {
	defer cConn.Close()

	var eg errgroup.Group

	var sConn *net.TCPConn
	addr := cConn.RemoteAddr()
	logged := 0
	
	eg.Go(func() error {
		buff := make([]byte, 0xFFFF)
		for {
			n, err := cConn.Read(buff)
			if err != nil {
				return err
			}
			b := buff[:n]
			if 0 < len(b) {
				if logged != 2 { 
					b1 := slices.Clone(b)
					if logged == 0 {
						p := &packet.HelloPacket{}
						r, l, err := p.Read(b1)
						if err != nil {
							return err
						}
						
						if r {
							server, ok := config.GetConfig().Servers[p.Hostname]
							if !ok {
								return cConn.Close()
							}

							if p.State == 1 {
								log.Printf("[INFO] %s is ping", addr)
							}

							serverIP, err := net.ResolveTCPAddr("tcp", server)
							if err != nil {
								return err
							}

							sConn, err = net.DialTCP("tcp", nil, serverIP)
							if err != nil {
								return err
							}

							if l != nil {
								log.Printf("[INFO] player is connected [%s]%s(%s)", addr.String(), l.Name, l.Uuid)
								logged = 2
								goto NEXT	
							}

							logged = 1
							goto NEXT
						}
					}

					if logged == 1 {
						p := &packet.LoginPacket{}
						r, err := p.Read(b1)
						if err != nil && err != io.EOF {
							return err
						}

						if r {
							log.Printf("[INFO] player is connected [%s]%s(%s)", addr.String(), p.Name, p.Uuid)
							logged = 2
							goto NEXT
						}
					}
				}
			}

			NEXT:
				if sConn != nil {
					n, err = sConn.Write(b)
					if err != nil {
						return err
					}
				}
		
			select {
			case <- ctx.Done():
				break
			default:
				continue
			}
		}
	})
	
	eg.Go(func() error { 
		for {
			if sConn != nil {
				n, err := io.Copy(cConn, sConn)
				if err != nil {
					return err
				}

				if 0 < n {
					return io.EOF
				}
			}

			select {
			case <- ctx.Done():
				return nil
			default:
				continue
			}
		}
	})

	err := eg.Wait()
	if sConn != nil {
	  sConn.Close()
	}

	return err
}
