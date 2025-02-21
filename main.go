package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/shiro8613/minecraft-a-proxy/config"
	"github.com/shiro8613/minecraft-a-proxy/proxy"
)

const CONFIG_PATH = "./config.yml"

func init() {
	if err := config.Load(CONFIG_PATH); err != nil {
		log.Fatalln("[ERROR] ", err)
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

	proxy.StartWatching(ctx)

	server := proxy.NewServer()
	log.Printf("[INFO] server is running on %s\n", conf.Bind)
	log.Println("[INFO] server mappings")
	for k, v := range conf.Servers {
		log.Printf(" - %s -> %s", k, v)
	}
	bindHost, err := net.ResolveTCPAddr("tcp", conf.Bind)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("[INFO] server started")
	if err := server.Start(ctx, bindHost); err != nil {
		log.Fatalln(err)
	}
}