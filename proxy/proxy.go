package proxy

import (
	"context"
	"io"
	"log"
	"net"
	"runtime"
	"slices"

	"github.com/shiro8613/minecraft-a-proxy/config"
	"github.com/shiro8613/minecraft-a-proxy/eg"
	"github.com/shiro8613/minecraft-a-proxy/packet"
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
			log.Fatalln("[ERROR] ", err)
		}
	}()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			if err == net.ErrClosed {
				return nil
			}
			return err
		}

		go s.handler(ctx, conn)
	}
}

func (s *ProxyServer) handler(ctx context.Context, c *net.TCPConn) {
	defer c.Close()

	addr := c.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		log.Println("[ERROR] ", err)
		s = nil
		c = nil
		runtime.GC()
		return
	}

	if HasBannedIps(ip) {
		c.Close()
		log.Printf("[INFO] banned-ip %s blocked\n", ip)
		s = nil
		c = nil
	
		runtime.GC()
		return
	}

	p := &Proxy{}
	err = p.Start(ctx, c)
	if err != nil && err != io.EOF {
		p = nil
		s = nil
		c = nil
		log.Println("[ERROR] ", err)
	}

	p = nil
	s = nil
	c = nil
	runtime.GC()
}

type Proxy struct {}

func (pr *Proxy) Start(ctx context.Context, cConn *net.TCPConn) error {
	defer cConn.Close()

	group := eg.New()

	var sConn *net.TCPConn
	closed, cancel := context.WithCancel(ctx)
	defer cancel()

	group.Go(func() error {
		addr := cConn.RemoteAddr()
		logged := 0	
		buff := make([]byte, 0xFFFF)

		for {
			n, err := cConn.Read(buff)
			if err != nil {
				buff = nil
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
							p = nil
							b1 = nil
							buff = nil
							b = nil
							return err
						}

						if r {
							server, ok := config.GetConfig().Servers[p.Hostname]
							if !ok {
								p = nil
								b1 = nil
								buff = nil
								b = nil
								return io.EOF
							}

							serverIP, err := net.ResolveTCPAddr("tcp", server)
							if err != nil {
								p = nil
								b1 = nil
								buff = nil
								b = nil
								serverIP = nil
								return err
							}

							sConn, err = net.DialTCP("tcp", nil, serverIP)
							if err != nil {
								p = nil
								b1 = nil
								buff = nil
								b = nil
								serverIP = nil
								sConn = nil
								return err
							}

							serverIP = nil

							if p.State == 1 {
								log.Printf("[INFO] %s is ping\n", addr)
							}

							if l != nil {
								log.Printf("[INFO] player is connected %s <- [%s]%s(%s)\n", server, addr.String(), l.Name, l.Uuid)
								b1 = nil
								l = nil
								logged = 2
								goto NEXT	
							}

							logged = 1
							p = nil
							b1 = nil

							goto NEXT
						}
					}

					if logged == 1 {
						p := &packet.LoginPacket{}
						r, err := p.Read(b1)
						if err != nil {
							buff = nil
							b = nil
							b1 = nil
							p = nil
							return err
						}

						if r {
							log.Printf("[INFO] player is connected (parse failed) <- [%s]%s(%s)\n", addr.String(), p.Name, p.Uuid)
							logged = 2
							goto NEXT
						}

						b1 = nil
						p = nil
					}
				} else if logged == 2 {
					addr = nil
					logged = 3
				}
			}

			NEXT:
				if sConn != nil {
					_, err = sConn.Write(b)
					if err != nil {
						buff = nil
						b = nil
						return err
					}
				} else {
					return io.EOF
				}
		
			b = nil

			select {
			case <- closed.Done():
				buff = nil
				return nil
			default:
			}
		}
	})
	
	group.Go(func() error {
		b := make([]byte, 0xFFFF)
		for {
			if sConn != nil {
				n, err := io.CopyBuffer(cConn, sConn, b)
				if err != nil {
					b = nil
					return err
				}

				if n < 0 {
					b = nil
					return io.EOF
				}
			}

			select {
			case <- closed.Done():
				b = nil
				return nil
			default:
			}
		}
	})

	err := group.Wait()
	cConn.Close()
	cConn = nil
	if sConn != nil {
		sConn.Close()
	}
	
	return err
}
