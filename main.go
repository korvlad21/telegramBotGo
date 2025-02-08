package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки файла .env")
	}
	// Инициализируем базу данных
	db := NewDB()
	defer db.Close() // Закрываем соединение после завершения работы
	// Получаем токен из переменных окружения (рекомендуется для безопасности)

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Не удалось найти токен бота в переменных окружения.")
	}

	// Создаём нового бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	// Включаем логирование взаимодействий с API Telegram
	bot.Debug = true

	log.Printf("Авторизован как: %s", bot.Self.UserName)
	log.Printf("Авторизован как: %s", bot.Self.ID)

	// Настраиваем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Основной цикл обработки сообщений
	for update := range updates {
		if update.Message != nil { // Если это сообщение
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			log.Printf("%s", update.Message.Text)
			// Создаём кнопки
			buttons := [][]tgbotapi.KeyboardButton{
				{
					tgbotapi.NewKeyboardButton("Кнопка 1"),
					tgbotapi.NewKeyboardButton("Кнопка 2"),
					tgbotapi.NewKeyboardButton("Кнопка 3"),
				},
				{
					tgbotapi.NewKeyboardButton("Кнопка 4"),
					tgbotapi.NewKeyboardButton("Кнопка 5"),
					tgbotapi.NewKeyboardButton("Кнопка 6"),
				},
				{
					tgbotapi.NewKeyboardButton("Кнопка 7"),
					tgbotapi.NewKeyboardButton("Кнопка 8"),
					tgbotapi.NewKeyboardButton("Кнопка 9"),
				},
				{
					tgbotapi.NewKeyboardButton("Кнопка 10"),
					tgbotapi.NewKeyboardButton("Кнопка 11"),
					tgbotapi.NewKeyboardButton("Кнопка 12"),
				},
			}

			// Создаём клавиатуру
			keyboard := tgbotapi.NewReplyKeyboard(buttons...)
			keyboard.OneTimeKeyboard = true // Клавиатура исчезнет после нажатия

			clients, err := getEngClients(db)
			if err != nil {
				log.Fatal(err)
			}
			m := ""
			place := 0
			for _, client := range clients {
				place++
				if place <= 3 {
					m += "Пользователь с именем " + client.Name()
				}
			}
			m += update.Message.Chat.UserName
			client, err := getEngClientByID(db, int(update.Message.Chat.ID))
			m += client.Name()
			// Отправляем сообщение с клавиатурой
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}
	}
}
