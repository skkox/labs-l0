package main

import "fmt"

func main() {
	fmt.Println("Сервис заказов L0")
	fmt.Println("====================")
	fmt.Println("План выполнения:")
	fmt.Println("1. Настройка PostgreSQL")
	fmt.Println("2. Создание репозитория")
	fmt.Println("3. Создание Go-проекта")
	fmt.Println("4. Структуры данных")
	fmt.Println("5. Подключение к PostgreSQL")
	fmt.Println("6. Подключение к NATS Streaming")
	fmt.Println("7. Сохранение данных")
	fmt.Println("8. HTTP-сервер")
	fmt.Println("9. Тестирование")
	fmt.Println("")
	
	// Тестируем структуры данных
	fmt.Println("Тестирование структур данных...")
	cache := NewOrderCache()
	fmt.Printf("Кэш создан: %+v\n", cache)
	
	fmt.Println("Готов к следующему шагу - подключение к PostgreSQL")
}
