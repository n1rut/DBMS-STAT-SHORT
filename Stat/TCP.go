package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Чтение данных из подключения
	scanner := bufio.NewScanner(conn)
	var data []byte
	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
		if len(scanner.Bytes()) == 0 {
			break
		}
	}

	// Преобразование данных в строку
	request := strings.TrimSpace(string(data))

	// Обработка запросов
	if request == "CHECKJSON" {
		// Отправка запроса SENDJSON СУБД
		existingEntries := reportData.Entries
		sendJSONRequest(existingEntries)
		// Отправка ответа клиенту после выполнения
		fmt.Fprintln(conn, "CHECKJSON выполнен.")
	} else if strings.HasPrefix(request, "REPORT") {
		// Разделение аргументов по пробелу
		args := strings.Fields(request)[1:]
		// Генерация отчета
		report := generateReport(args)
		// Сохранение отчета в файл
		saveReportToFile(report, "reportStat.json")
		// Отправка ответа после выполнения
		fmt.Fprintln(conn, "Отчет сгенерирован и сохранен в файл reportStat.json.")
		// Вывод отчета клиенту
		fmt.Fprintln(conn, "Отчет:")
		printJSONToConn(report, conn)
	} else {
		// Разбор JSON данных
		var newReportData ReportData
		if err := json.Unmarshal(data, &newReportData); err != nil {
			log.Println("Ошибка при разборе JSON:", err)
			return
		}

		reportData.Entries = append(reportData.Entries, newReportData.Entries...)

		// Вывод данных отчета
		fmt.Println("Получены данные от БД:")
		for _, entry := range newReportData.Entries {
			fmt.Printf("ID: %d, URL: %s (%s), SourceIP: %s, Count: %d\n", entry.ID, entry.OriginalURL, entry.ShortURL, entry.SourceIP, entry.Count)
		}
	}
}
