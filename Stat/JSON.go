package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
)

func sendJSONRequest(existingEntries []ReportEntry) {
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

	// Разбор JSON данных
	var newReportData ReportData
	if err := json.Unmarshal([]byte(response), &newReportData); err != nil {
		fmt.Println("Ошибка при разборе JSON:", err)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Проверка наличия данных перед добавлением
	for _, entry := range newReportData.Entries {
		if !containsEntryByID(existingEntries, entry.ID) {
			reportData.Entries = append(reportData.Entries, entry)
		}
	}
}

// containsEntry проверяет, содержится ли запись в слайсе по ID
func containsEntryByID(entries []ReportEntry, id int) bool {
	for _, e := range entries {
		if e.ID == id {
			return true
		}
	}
	return false
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

func printJSONToConn(report DetailReport, conn net.Conn) {
	// Преобразование структуры в JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Println("Ошибка при маршалинге в JSON:", err)
		return
	}

	// Отправка JSON данных клиенту
	lines := strings.Split(string(jsonData), "\n")
	for _, line := range lines {
		fmt.Fprintln(conn, line)
	}
	fmt.Println("JSON отправлен клиенту.")
}
