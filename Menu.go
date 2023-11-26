package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func processCommand(command string, set *Set, stack *Stack, queue *Queue, ht *HashTable, filename string) string {
	// Разбиваем команду на токены, разделенные пробелами
	tokens := strings.Fields(command)

	if len(tokens) < 1 {
		return "Пустая команда"
	}

	cmd := tokens[0]

	switch cmd {
	case "SENDJSON":
		// Обработка команды SENDJSON
		sendReportsJSONToStatisticService()
		return "SENDJSON выполнено"

	case "SHORTLINK":
		// Обработка команды SHORTLINK
		if len(tokens) < 3 {
			return "Недостаточно аргументов для команды SHORTLINK"
		}
		shortLink := tokens[1]
		originalLink := strings.Join(tokens[2:], " ")
		linkMap[shortLink] = originalLink
		saveLinksToFile()

		return fmt.Sprintf("SHORTLINK выполнено: %s -> %s", shortLink, originalLink)

	case "SADD":
		// Обработка команды SADD
		if len(tokens) < 2 {
			return "Недостаточно аргументов для команды SADD"
		}
		key := tokens[1]
		set.SADD(filename, key)
		return "SADD выполнено"

	case "SREM":
		// Обработка команды SREM
		if len(tokens) < 2 {
			return "Недостаточно аргументов для команды SREM"
		}
		key := tokens[1]
		set.SREM(filename, key)
		return "SREM выполнено"

	case "SISMEMBER":
		// Обработка команды SISMEMBER
		if len(tokens) < 2 {
			return "Недостаточно аргументов для команды SISMEMBER"
		}
		key := tokens[1]
		if set.SISMEMBER(filename, key) {
			return fmt.Sprintf("%s присутствует в множестве", key)
		} else {
			return fmt.Sprintf("%s не найден в множестве", key)
		}

	case "SPUSH":
		// Обработка команды SPUSH
		if len(tokens) < 2 {
			return "Недостаточно аргументов для команды SPUSH"
		}
		val := tokens[1]
		stack.SPUSH(filename, val)
		return "SPUSH выполнено"

	case "SPOP":
		// Обработка команды SPOP
		val, err := stack.SPOP(filename)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("SPOP выполнено: %s", val)

	case "QPUSH":
		// Обработка команды QPUSH
		if len(tokens) < 2 {
			return "Недостаточно аргументов для команды QPUSH"
		}
		val := tokens[1]
		queue.QPUSH(filename, val)
		return "QPUSH выполнено"

	case "QPOP":
		// Обработка команды QPOP
		val, err := queue.QPOP(filename)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("QPOP выполнено: %s", val)

	case "HSET":
		// Обработка команды HSET
		if len(tokens) < 3 {
			return "Недостаточно аргументов для команды HSET"
		}
		key := tokens[1]
		value := tokens[2]
		ht.HSET(filename, key, value)
		return "HSET выполнено"

	case "HDEL":
		// Обработка команды HDEL
		if len(tokens) < 2 {
			return "Недостаточно аргументов для команды HDEL"
		}
		key := tokens[1]
		ht.HDEL(filename, key)
		return "HDEL выполнено"

	case "HGET":
		// Обработка команды HGET
		if len(tokens) < 2 {
			return "Недостаточно аргументов для команды HGET"
		}
		key := tokens[1]
		value, err := ht.HGET(filename, key)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("HGET выполнено: %s", value)

	default:
		return "Неизвестная команда"
	}
}

// Функция для чтения строк из файла
func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Функция для записи строк в файл
func writeLines(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
