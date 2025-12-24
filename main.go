package main

import (
	"flag"
	"log"
	"net"
	"sync"
)

var (
	listenPort = flag.String("listen", "", "Порт для прослушивания (пример: :8080)")
	targetAddr = flag.String("target", "", "Адрес для подключения (пример: 192.168.1.2:8080)")
)

func main() {
	flag.Parse()

	if *listenPort == "" && *targetAddr == "" {
		log.Fatal("Укажите хотя бы один флаг: --listen или --target")
	}

	var wg sync.WaitGroup
	connChan := make(chan net.Conn, 1)

	if *listenPort != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listen(*listenPort, connChan)
		}()
	}

	if *targetAddr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			connect(*targetAddr, connChan)
		}()
	}

	var conn net.Conn
	select {
	case conn = <-connChan:
		log.Println("Соединение установлено!")
	case <-make(chan struct{}):
	}

	if conn != nil {
		handleChat(conn)
	}

	wg.Wait()
}
