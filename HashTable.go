package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Определение структуры Node для хранения данных в хэш-таблице
type NodeHT struct {
	key   string
	value string
	next  *NodeHT
}

// Определение структуры HashTable для хэш-таблицы
type HashTable struct {
	data     []*NodeHT
	capacity int
	mu       sync.Mutex // Мьютекс для потокобезопасности
}

// Функция для вычисления хэша ключа
func (ht *HashTable) hash(key string) int {
	hash := 0
	for _, char := range key {
		hash += int(char)
	}
	return hash % ht.capacity
}

func (ht *HashTable) HSET(filename, key, value string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	// Проверяем уникальность ключа и хэша в памяти программы
	index := ht.hash(key)
	if ht.exists(key, index) {
		fmt.Println("Ключ уже существует в памяти программы:", key)
		return
	}

	// Проверяем уникальность ключа в файле
	if fileContainsKey(filename, key) {
		fmt.Println("Ключ уже существует в файле:", key)
		return
	}

	// Проверяем наличие элементов с тем же хешем в файле
	if fileContainsHash(filename, index) {
		fmt.Println("Коллизия: элемент с таким же хешем уже существует в файле:", key)
		return
	}

	newNode := &NodeHT{key: key, value: value}

	// Записываем значение в файл в формате "HashTable: {ключ:значение:хеш}"
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("HashTable: {%s:%s:%d}\n", key, value, index))
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
		return
	}

	// Добавляем элемент в хэш-таблицу
	if ht.data[index] == nil {
		ht.data[index] = newNode
	} else {
		current := ht.data[index]
		for current.next != nil {
			current = current.next
		}
		current.next = newNode
	}

	fmt.Println("Добавлено в HashTable:", key)
}

// Функция для проверки наличия элементов с тем же хешем в файле
func fileContainsHash(filename string, hash int) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false // Если ошибка при открытии файла, считаем, что элементов с таким хешем в файле нет
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "HashTable: {") && strings.HasSuffix(line, "}") {
			// Извлекаем содержимое внутри фигурных скобок
			content := line[len("HashTable: {") : len(line)-1]
			// Разбиваем содержимое по двоеточию
			parts := strings.Split(content, ":")
			if len(parts) == 3 {
				h, err := strconv.Atoi(parts[2])
				if err != nil {
					continue
				}
				if h == hash {
					return true // Найден элемент с таким же хешем
				}
			}
		}
	}

	return false // Элементов с таким хешем в файле нет
}

// Функция для проверки уникальности ключа и хэша в памяти программы
func (ht *HashTable) exists(key string, index int) bool {
	current := ht.data[index]
	for current != nil {
		if current.key == key {
			return true // Ключ уже существует
		}
		current = current.next
	}
	return false
}

// Функция для проверки наличия ключа в файле
func fileContainsKey(filename, key string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false // Если ошибка при открытии файла, считаем, что ключа в файле нет
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "HashTable: {") && strings.HasSuffix(line, "}") {
			// Извлекаем содержимое внутри фигурных скобок
			content := line[len("HashTable: {") : len(line)-1]
			// Разбиваем содержимое по двоеточию
			parts := strings.Split(content, ":")
			if len(parts) == 3 && parts[0] == key {
				return true // Ключ найден в файле
			}
		}
	}

	return false // Ключ не найден в файле
}

// Функция HGET для чтения значения по ключу из файла
func (ht *HashTable) HGET(filename, key string) (string, error) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	// Попробуем сначала найти ключ в файле
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return "", err
	}
	defer file.Close()

	formattedKey := fmt.Sprintf("HashTable: {%s", key)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, formattedKey) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return parts[1], nil
			}
		}
	}

	// Если ключ не найден в файле, попробуем найти в памяти программы
	index := ht.hash(key)
	current := ht.data[index]

	for current != nil {
		if current.key == key {
			return current.value, nil
		}
		current = current.next
	}

	// Если не найдено ни в файле, ни в памяти программы, возвращаем ошибку
	return "", fmt.Errorf("Ключ %s не найден", key)
}

func (ht *HashTable) HDEL(filename, key string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	// Удаляем из памяти программы
	index := ht.hash(key)
	current := ht.data[index]
	var prev *NodeHT

	for current != nil {
		if current.key == key {
			// Сохраняем значение перед удалением
			deletedValue := current.value

			if prev == nil {
				ht.data[index] = current.next
			} else {
				prev.next = current.next
			}

			fmt.Println("Удалено из хэш-таблицы:", deletedValue)

			break
		}
		prev = current
		current = current.next
	}

	// Теперь удалим из файла
	lines, err := readLines(filename)
	if err != nil {
		fmt.Println("Ошибка при чтении из файла:", err)
		return
	}

	updatedLines := make([]string, 0)

	for _, line := range lines {
		if strings.HasPrefix(line, "HashTable: {") && strings.HasSuffix(line, "}") {
			// Обрабатываем строку в формате "HashTable: {...}"
			parts := strings.Split(line, "{")
			if len(parts) != 2 {
				fmt.Println("Ошибка при обработке строки HashTable:", line)
				continue
			}

			// Получаем внутренний текст внутри фигурных скобок
			internalText := parts[1]

			// Разбиваем внутренний текст по запятой, чтобы получить ключи и значения
			keyValuePairs := strings.Split(internalText, ",")
			updatedKeyValuePairs := make([]string, 0)

			for _, kvPair := range keyValuePairs {
				parts := strings.SplitN(kvPair, ":", 2)
				if len(parts) != 2 {
					fmt.Println("Ошибка при обработке пары ключ-значение:", kvPair)
					continue
				}
				k := strings.TrimSpace(parts[0])

				if k != key {
					// Сохраняем все строки, кроме удаляемой
					updatedKeyValuePairs = append(updatedKeyValuePairs, kvPair)
				}
			}

			if len(updatedKeyValuePairs) > 0 {
				// Собираем обновленную строку HashTable: {...}
				updatedLine := "HashTable: {" + strings.Join(updatedKeyValuePairs, ", ") + "}"
				updatedLines = append(updatedLines, updatedLine)
			}
		} else {
			// Просто добавляем строки, не связанные с HashTable
			updatedLines = append(updatedLines, line)
		}
	}

	// Записываем обновленные строки в файл
	err = writeLines(filename, updatedLines)
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
	}
}
