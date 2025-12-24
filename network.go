package main

import (
	"log"
	"net"
)

func listen(port string, connChan chan<- net.Conn) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("Ошибка запуска слушателя: %v", err)
		return
	}
	defer listener.Close()

	log.Printf("Ожидание подключения на порту %s...", port)
	conn, err := listener.Accept()
	if err != nil {
		log.Printf("Ошибка принятия соединения: %v", err)
		return
	}

	log.Printf("Подключение принято от %s", conn.RemoteAddr())
	connChan <- conn
}

func connect(addr string, connChan chan<- net.Conn) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("Не удалось подключиться к %s: %v", addr, err)
		return
	}

	log.Printf("Успешно подключились к %s", addr)
	connChan <- conn
}
