package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
)

type ReportEntry struct {
	ID           int    `json:"Id"`
	PID          *int   `json:"Pid"`
	OriginalURL  string `json:"OriginalURL,omitempty"`
	ShortURL     string `json:"ShortURL,omitempty"`
	SourceIP     string `json:"SourceIP"`
	TimeInterval string `json:"TimeInterval"`
	Count        int    `json:"Count"`
}

type ReportData struct {
	Entries []ReportEntry `json:"entries"`
}

type DetailReport struct {
	Count   int                      `json:"Count,omitempty"`
	Details map[string]*DetailReport `json:"Details,omitempty"`
}

var mu sync.Mutex
var reportData ReportData

func main() {
	go startStatisticService()
	startUserInterface()
}

func startStatisticService() {
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Сервис статистики запущен. Ожидание подключений...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Ошибка при подключении клиента:", err)
			continue
		}

		fmt.Println("Подключение от", conn.RemoteAddr().String())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Чтение данных из подключения
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Println("Ошибка при чтении данных:", err)
		return
	}

	// Разбор JSON данных
	var newReportData ReportData
	if err := json.Unmarshal(data, &newReportData); err != nil {
		log.Println("Ошибка при разборе JSON:", err)
		return
	}

	// Добавление новых данных к существующим
	mu.Lock()
	reportData.Entries = append(reportData.Entries, newReportData.Entries...)
	mu.Unlock()

	// Вывод данных отчета
	fmt.Println("Получены данные от БД:")
	for _, entry := range newReportData.Entries {
		fmt.Printf("ID: %d, URL: %s (%s), SourceIP: %s, Count: %d\n", entry.ID, entry.OriginalURL, entry.ShortURL, entry.SourceIP, entry.Count)
	}
}

func startUserInterface() {
	for {
		displayMenu()

		choice := getUserChoice()

		switch choice {
		case 1:
			sendJSONRequest()
		case 2:
			fmt.Println("Генерация отчета. Введите порядок детализаций (например, \"SourceIP TimeInterval URL\"): ")

			// Запрос у пользователя каждого элемента последовательности
			var detailsOrder []string
			for i := 0; i < 3; i++ {
				var detail string
				fmt.Scan(&detail)
				detailsOrder = append(detailsOrder, detail)
			}

			report := generateReport(detailsOrder)
			fmt.Println("Отчет:")
			printJSON(report)
			// Сохранение отчета в файл
			saveReportToFile(report, "reportStat.json")
		case 3:
			fmt.Println("Выход из программы.")
			return
		default:
			fmt.Println("Некорректный выбор. Пожалуйста, введите число от 1 до 3.")
		}
	}
}

func displayMenu() {
	fmt.Println("\nМеню:")
	fmt.Println("1. Подгрузить данные с сервера (SENDJSON)")
	fmt.Println("2. Отчет")
	fmt.Println("3. Выход из программы")
	fmt.Print("Выберите опцию: ")
}

func getUserChoice() int {
	var choice int
	fmt.Scan(&choice)
	return choice
}

func sendJSONRequest() {
	fmt.Println("Отправка запроса SENDJSON на сервер БД...")

	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Ошибка подключения к серверу БД:", err)
		return
	}
	defer conn.Close()

	// Отправка запроса SENDJSON
	conn.Write([]byte("SENDJSON\n"))

	// Чтение ответа от сервера БД
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка при чтении ответа от сервера БД:", err)
		return
	}

	fmt.Println("Ответ от сервера БД:", strings.TrimSpace(response))
}

func generateReport(detailsOrder []string) DetailReport {
	mu.Lock()
	defer mu.Unlock()

	// Создаем карту для хранения данных отчета
	report := DetailReport{Count: 0}

	// Заполняем карту данными из reportData
	for _, entry := range reportData.Entries {
		currLevel := &report
		currLevel.Count += entry.Count

		for _, level := range detailsOrder {
			switch level {
			case "SourceIP":
				currLevel = currLevel.getOrCreateDetail(entry.SourceIP)
			case "TimeInterval":
				currLevel = currLevel.getOrCreateDetail(entry.TimeInterval)
			case "URL":
				currLevel = currLevel.getOrCreateDetail(fmt.Sprintf("%s (%s)", entry.OriginalURL, entry.ShortURL))
			}

			currLevel.Count += entry.Count
		}
	}

	return report
}

func (dr *DetailReport) getOrCreateDetail(key string) *DetailReport {
	if dr.Details == nil {
		dr.Details = make(map[string]*DetailReport)
	}

	if _, ok := dr.Details[key]; !ok {
		dr.Details[key] = &DetailReport{}
	}

	return dr.Details[key]
}

func saveReportToFile(report DetailReport, filename string) {
	// Преобразование структуры в JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Println("Ошибка при маршалинге в JSON:", err)
		return
	}

	// Запись в файл
	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
		return
	}

	fmt.Printf("Отчет сохранен в файл %s.\n", filename)
}

func printJSON(report DetailReport) {
	// Преобразование структуры в JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Println("Ошибка при маршалинге в JSON:", err)
		return
	}

	fmt.Println(string(jsonData))
}
