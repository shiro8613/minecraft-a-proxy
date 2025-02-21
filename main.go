package main

import (
	"context"
	"log"
	"net"
	"net/netip"
	"os"
	"os/signal"

	"github.com/shiro8613/minecraft-a-proxy/config"
	"github.com/shiro8613/minecraft-a-proxy/proxy"
)

const CONFIG_PATH = "./config.yml"

func init() {
	if err := config.Load(CONFIG_PATH); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	conf := config.GetConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<- ctx.Done()
		log.Println("[INFO] shutdown...")
	}()
	defer stop()

	server := proxy.NewServer()
	log.Printf("server is running on %s", conf.Bind)
	log.Println("server mappings")
	for k, v := range conf.Servers {
		log.Printf(" - %s -> %s", k, v)
	}
	log.Println("server started")
	if err := server.Start(ctx, net.TCPAddrFromAddrPort(netip.MustParseAddrPort(conf.Bind))); err != nil {
		log.Fatalln(err)
	}
}