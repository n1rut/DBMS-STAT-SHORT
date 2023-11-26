package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type NodeS struct {
	data string
	next *NodeS
}

type Stack struct {
	head *NodeS
}

// Функция SPUSH для добавления элемента в стек и записи в файл
func (s *Stack) SPUSH(filename, val string) {

	if strings.HasPrefix(val, "Stack: {") && strings.HasSuffix(val, "}") {
		newNode := &NodeS{data: val, next: s.head}
		s.head = newNode
		fmt.Println("Добавлено в стек:", val)
	} else {
		// Преобразовываем значение в формат "Stack: {...}"
		formattedVal := fmt.Sprintf("Stack: {%s}", val)

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

		// Добавляем значение в стек
		newNode := &NodeS{data: formattedVal, next: s.head}
		s.head = newNode
		fmt.Println("Добавлено в стек:", formattedVal)
	}
}

// Функция SPOP для удаления и вывода значения из стека и удаления из файла
func (s *Stack) SPOP(filename string) (string, error) {

	// Удаляем значение из файла, если оно соответствует формату "Stack: {...}"
	lines, err := readLines(filename)
	if err != nil {
		fmt.Println("Ошибка при чтении из файла:", err)
		return "", nil
	}

	var deletedLine string // Для хранения удаленной строки
	found := false

	// Создаем слайс для обновленных строк
	updatedLines := make([]string, 0)

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if !found && strings.HasPrefix(line, "Stack: {") && strings.HasSuffix(line, "}") {
			deletedLine = line // Сохраняем удаленную строку
			found = true
			continue
		}
		updatedLines = append(updatedLines, line)
	}

	if !found {
		return "", errors.New("Строка в формате 'Stack: {...}' не найдена в файле")
	}

	// Разворачиваем обновленные строки, чтобы вернуть их в исходный порядок
	for i, j := 0, len(updatedLines)-1; i < j; i, j = i+1, j-1 {
		updatedLines[i], updatedLines[j] = updatedLines[j], updatedLines[i]
	}

	err = writeLines(filename, updatedLines)
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
	}

	// Удаляем элемент из памяти программы
	if s.head == nil {
		return "", errors.New("Стек пуст") // Возвращаем ошибку, если стек пуст
	}
	data := s.head.data
	s.head = s.head.next
	fmt.Println("Удалено из стека:", data)
	fmt.Printf("Удалено из файла: %s\n", deletedLine)

	return data, nil
}
