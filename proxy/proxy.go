package proxy

import (
	"context"
	"log"
	"io"
	"net"
	"slices"

	"github.com/shiro8613/minecraft-a-proxy/packet"
	"github.com/shiro8613/minecraft-a-proxy/config"

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

	eg, eg_ctx := errgroup.WithContext(context.Background())
	eg_ctx, cancel := context.WithCancel(eg_ctx)

	var sConn *net.TCPConn
	addr := cConn.RemoteAddr()
	logged := 0
	
	eg.Go(func() error {
		buff := make([]byte, 0xFFFF)
		var e error
		for {
			n, err := cConn.Read(buff)
			if err != nil {
				e = err
				break
			}
			b := buff[:n]
			if 0 < len(b) {
				if logged != 2 { 
					b1 := slices.Clone(b)
					if sConn == nil {
						p := &packet.HelloPacket{}
						r, err := p.Read(b1)
						if err != nil {
							e = err
							break
						}
						
						if r {
							server, ok := config.GetConfig().Servers[p.Hostname]
							if !ok {
								if err := cConn.Close(); err != nil {
									e = err
								}
								break
							}

							if p.State == 1 {
								log.Printf("[INFO] %s is ping", addr)
							}

							serverIP, err := net.ResolveTCPAddr("tcp", server)
							if err != nil {
								e = err
								break
							}

							sConn, err = net.DialTCP("tcp", nil, serverIP)
							if err != nil {
								e = err
								break
							}

							logged = 1
							goto NEXT
						}
					}

					if sConn != nil && logged == 1 {
						p := &packet.LoginPacket{}
						r, err := p.Read(b1)
						if err != nil && err != io.EOF {
							e = err
							break
						}

						if r && 3 < p.Length {
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
						e = err
						break
					}
				}
		
			select {
			case <- eg_ctx.Done():
				break
			default:
				continue
			}
		}
		cancel()
		return e
	})
	
	eg.Go(func() error { 
		for {
			if sConn != nil {
				_, err := io.Copy(cConn, sConn)		
				if err != nil && err != io.EOF {
					return err
				}
			}

			select {
			case <- eg_ctx.Done():
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
