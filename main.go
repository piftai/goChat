package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

// Структура для хранения информации о клиенте
type Client struct {
	conn net.Conn
	name string
}

var (
	clients   = make(map[net.Conn]Client) // Хранит всех подключенных клиентов
	clientsMu sync.Mutex                  // Мьютекс для безопасного доступа к clients
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Запрашиваем имя клиента
	conn.Write([]byte("Введите ваше имя: "))
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	name := scanner.Text()

	// Добавляем клиента в список
	clientsMu.Lock()
	clients[conn] = Client{conn: conn, name: name}
	clientsMu.Unlock()

	// Уведомляем всех о новом участнике
	broadcastMessage(fmt.Sprintf("%s присоединился к чату!", name), conn)

	// Читаем сообщения от клиента
	for scanner.Scan() {
		message := scanner.Text()
		broadcastMessage(fmt.Sprintf("%s: %s", name, message), conn)
	}

	// Удаляем клиента при отключении
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
	broadcastMessage(fmt.Sprintf("%s покинул чат.", name), conn)
}

// Рассылает сообщение всем клиентам, кроме отправителя
func broadcastMessage(message string, sender net.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for conn, _ := range clients {
		if conn != sender { // Не отправляем сообщение отправителю
			_, err := conn.Write([]byte(message + "\n"))
			if err != nil {
				log.Println("Ошибка отправки сообщения:", err)
				delete(clients, conn)
				conn.Close()
			}
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("Сервер запущен на localhost:8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Ошибка подключения:", err)
			continue
		}
		go handleConnection(conn)
	}
}
