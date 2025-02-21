package proxy

import (
	"bufio"
	"context"
	"log"
	"os"
	"slices"
	"sync"

	"github.com/fsnotify/fsnotify"
)


var (
	mu sync.Mutex
	internal_blocked_ips []string = make([]string, 0)
)

const BANNED_IP_PATH = "./banned-ips.txt"

func StartWatching(ctx context.Context) {
	go startWatching(ctx)
}

func startWatching(ctx context.Context) {
	log.Println("[INFO] banned-ips is loading..")
	if err := read(BANNED_IP_PATH); err != nil {
		log.Fatalln(err)	
	}
	log.Println("[INFO] banned-ips is loaded")

	watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

	go func() {
		for {
            select {
			case <- ctx.Done():
				return
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                if event.Has(fsnotify.Write) {
					log.Println("[INFO] banned-ips is reloading")
					if err := read(BANNED_IP_PATH); err != nil {
						log.Println("[ERROR] ", err)
						return
					}
					log.Println("[INFO] banned-ips is reloaded")
				}
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("[ERROR] ", err)
            }
        }
	}()

	err = watcher.Add(BANNED_IP_PATH)
    if err != nil {
        log.Fatalln("[ERROR] ", err)
    }

	<- ctx.Done()
}

func read(path string) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	mu.Lock()
	for scanner.Scan() {
		internal_blocked_ips = append(internal_blocked_ips, scanner.Text())
	}
	mu.Unlock()
	if err := scanner.Err(); err != nil {
		return err
	}
	
	return nil
}

func HasBannedIps(ip string) bool {
	mu.Lock()
	b := slices.Contains(internal_blocked_ips, ip)
	mu.Unlock()

	return b
}