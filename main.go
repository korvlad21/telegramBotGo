package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки файла .env")
	}
	bot, err := NewBot()
	if err != nil {
		log.Fatal("❌ Ошибка инициализации бота:", err)
	}
	defer bot.Close()

	log.Printf("Авторизован как: %s", bot.api.Self.UserName)
	log.Printf("Авторизован как: %s", bot.api.Self.ID)

	bot.Run()
}
