package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Определение структуры Node для элементов очереди
type Node struct {
	data string
	next *Node
}

// Определение структуры Queue для очереди
type Queue struct {
	head *Node
	tail *Node
}

// Функция QPUSH для добавления элемента в очередь и записи в файл
func (q *Queue) QPUSH(filename, val string) {

	if strings.HasPrefix(val, "Queue: {") && strings.HasSuffix(val, "}") {
		newNode := &Node{data: val, next: nil}
		if q.tail == nil {
			q.head = newNode
			q.tail = newNode
			fmt.Println("Добавлено в очередь:", val)
		} else {
			q.tail.next = newNode
			q.tail = newNode
			fmt.Println("Добавлено в очередь:", val)
		}
	} else {
		// Преобразовываем значение в формат "Queue: {...}"
		formattedVal := fmt.Sprintf("Queue: {%s}", val)

		// Записываем значение в файл
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Ошибка при открытии файла:", err)
			return
		}
		defer file.Close()

		_, err = file.WriteString(formattedVal + "\n")
		if err != nil {
			fmt.Println("Ошибка при записи в файл:", err)
			return
		}

		// Добавляем значение в очередь
		newNode := &Node{data: formattedVal, next: nil}
		if q.tail == nil {
			q.head = newNode
			q.tail = newNode
		} else {
			q.tail.next = newNode
			q.tail = newNode
		}
		fmt.Println("Добавлено в очередь:", formattedVal)
	}
}

// Функция QPOP для удаления и возврата элемента из очереди и удаления из файла
func (q *Queue) QPOP(filename string) (string, error) {

	// Проверяем наличие строк в формате "Queue: {...}" в файле
	lines, err := readLines(filename)
	if err != nil {
		fmt.Println("Ошибка при чтении из файла:", err)
		return "", err
	}

	var found bool
	var deletedLine string // Для хранения удаленной строки

	// Создаем слайс для обновленных строк
	updatedLines := make([]string, 0)

	for _, line := range lines {
		if !found && strings.HasPrefix(line, "Queue: {") && strings.HasSuffix(line, "}") {
			deletedLine = line // Сохраняем удаленную строку
			found = true
			continue
		}
		updatedLines = append(updatedLines, line)
	}

	if !found {
		fmt.Println("Очередь пуста.")
		return "", nil
	}

	err = writeLines(filename, updatedLines)
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
	}

	// Удаляем элемент из памяти программы
	if q.head == nil {
		return "", errors.New("Очередь пуста") // Возвращаем ошибку, если очередь пуста
	}
	data := q.head.data
	q.head = q.head.next
	if q.head == nil {
		q.tail = nil // Если удалили последний элемент, обнуляем и tail
	}

	// Выводим сообщение об удалении строки из файла
	fmt.Printf("Удалено из очереди: %s\n", deletedLine)

	return data, nil
}
