package main

import (
	"fmt"
	"os"
)

// Определение структуры Set для множества
type Set struct {
	head *SetNode
}

// Определение структуры SetNode для элементов множества
type SetNode struct {
	key  string
	next *SetNode
}

func (s *Set) SADD(filename, key string) {
	// Проверяем, есть ли элемент в множестве
	if s.SISMEMBER(filename, key) {
		return
	}

	// Преобразовываем значение в формат "Set: {...}"
	formattedKey := fmt.Sprintf("Set: {%s}", key)

	newNode := &SetNode{key: formattedKey}

	// Если множество пусто, делаем newNode его головой
	if s.head == nil {
		s.head = newNode
	} else {
		current := s.head
		for current.next != nil {
			current = current.next
		}
		current.next = newNode
	}

	fmt.Println("Добавлено в множество:", key)

	// Записываем значение в файл
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(formattedKey + "\n")
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
	}
}

// Функция SREM для удаления элемента из множества и удаления из файла
func (s *Set) SREM(filename, key string) {
	// Преобразовываем значение в формат "Set: {...}"
	formattedKey := fmt.Sprintf("Set: {%s}", key)

	// Переменная для отслеживания, было ли значение удалено из множества
	removedFromSet := false

	// Если значение было удалено из файла, удалите его из множества
	lines, err := readLines(filename)
	if err != nil {
		fmt.Println("Ошибка при чтении из файла:", err)
		return
	}

	updatedLines := make([]string, 0)
	for _, line := range lines {
		if line == formattedKey {
			removedFromSet = true
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	err = writeLines(filename, updatedLines)
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
	}

	// Если значение было удалено из файла, удалите его из множества в памяти
	if s.head == nil {
		fmt.Println("Множество пусто.")
		return
	}

	// Если удаляемый элемент - голова множества
	if s.head.key == formattedKey {
		s.head = s.head.next
		fmt.Println("Удалено из множества:", key)
		return
	}

	current := s.head
	var prev *SetNode

	for current != nil {
		if current.key == formattedKey {
			prev.next = current.next
			fmt.Println("Удалено из множества:", key)
			break
		}
		prev = current
		current = current.next
	}

	if !removedFromSet {
		fmt.Println("Значение не найдено в множестве.")
	}
}

// Функция SISMEMBER для поиска значения в файле
func (s *Set) SISMEMBER(filename, key string) bool {
	// Преобразовываем значение в формат "Set: {...}"
	formattedKey := fmt.Sprintf("Set: {%s}", key)

	lines, err := readLines(filename)
	if err != nil {
		fmt.Println("Ошибка при чтении из файла:", err)
		return false
	}

	for _, line := range lines {
		if line == formattedKey {
			return true
		}
	}

	return false
}
