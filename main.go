package main

import (
	"log"
	"net"
	"net/netip"

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
	server := proxy.NewServer()
	log.Printf("server is running on %s", conf.Bind)
	log.Println("server mappings")
	for k, v := range conf.Servers {
		log.Printf(" - %s -> %s", k, v)
	}
	log.Println("server started")
	if err := server.Start(net.TCPAddrFromAddrPort(netip.MustParseAddrPort(conf.Bind))); err != nil {
		log.Fatalln(err)
	}
}