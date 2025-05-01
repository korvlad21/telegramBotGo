package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // Подключаем драйвер для MySQL
)

// DB структура для хранения подключения
type DB struct {
	Connection *sql.DB
}

// NewDB функция для инициализации подключения к базе данных
func NewDB() *DB {
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_DATABASE")
	dbPort := os.Getenv("DB_PORT")

	port, err := strconv.Atoi(dbPort)
	if err != nil {
		log.Fatalf("Неверный формат порта: %v", err)
	}

	// Формируем DSN с указанием порта
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbUser, dbPassword, dbHost, port, dbName)

	connection, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Проверяем соединение
	if err := connection.Ping(); err != nil {
		log.Fatalf("Не удалось установить соединение с базой данных: %v", err)
	}

	log.Println("Успешное подключение к базе данных!")
	return &DB{Connection: connection}
}

// Close закрывает соединение с базой данных
func (db *DB) Close() {
	err := db.Connection.Close()
	if err != nil {
		log.Printf("Ошибка при закрытии соединения с базой данных: %v", err)
	}
}
