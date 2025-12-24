package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func handleChat(conn net.Conn) {
	defer conn.Close()

	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Соединение разорвано")
				os.Exit(0)
			}
			fmt.Print(fmt.Sprintf("\nСообщение от %s: %s> ", conn.RemoteAddr().String(), msg))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		msg := scanner.Text()
		if strings.TrimSpace(msg) == "/exit" {
			return
		}
		_, err := fmt.Fprintf(conn, "%s\n", msg)
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err)
			return
		}
		fmt.Print("> ")
	}
}
