package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

var linkMap map[string]string
var characters string
var linkLength = 6

func init() {
	characters = generateCharacters()
}

func generateCharacters() string {
	var chars []rune
	for i := 'a'; i <= 'z'; i++ {
		chars = append(chars, i)
	}
	for i := 'A'; i <= 'Z'; i++ {
		chars = append(chars, i)
	}
	for i := '0'; i <= '9'; i++ {
		chars = append(chars, i)
	}
	return string(chars)
}

func main() {
	linkMap = make(map[string]string)
	rand.Seed(time.Now().UnixNano())

	go startDBServer()

	// Загрузка ссылок и счетчика
	loadLinksFromFile()
	loadCounterFromFile()

	// Маршрут для создания сокращенной ссылки
	http.HandleFunc("/shorten", shortenLink)

	// Маршрут для перенаправления
	http.HandleFunc("/redirect/", redirectLink)

	port := 8080
	fmt.Printf("Сервер слушает на порту %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Ошибка при запуске сервера: ", err)
	}
}

func startDBServer() {
	var mutex sync.Mutex
	fmt.Println("Запуск СУБД...")
	// Создаем слушатель
	ln, err := net.Listen("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Ошибка при запуске сервера СУБД:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Сервер СУБД запущен. Ожидание подключений...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка при подключении клиента СУБД:", err)
			continue
		}

		fmt.Println("Подключение от", conn.RemoteAddr().String())

		go handleConnection(conn, &mutex)
	}
}
