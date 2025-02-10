package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
			m := ""

			users, err := getUsers(db)
			if err != nil {
				log.Fatal(err)
			}
			place := 0
			topFive := false
			for _, usr := range users {
				place++
				if place <= 5 {
					if strconv.FormatInt(update.Message.Chat.ID, 10) == usr.getID() {
						m += fmt.Sprintf("<b>%d. %s: %.2f%%</b>\n", place, usr.getName(), usr.getTotalRate())
						topFive = true
					} else {
						m += fmt.Sprintf("%d. %s: %.2f%%\n", place, usr.getName(), usr.getTotalRate())
					}
				} else if topFive {
					break
				} else if strconv.FormatInt(update.Message.Chat.ID, 10) == usr.getID() {
					m += "--------------------------------\n"
					m += fmt.Sprintf("<b>%d. %s: %.2f%%</b>\n", place, usr.getName(), usr.getTotalRate())
					topFive = true
				}
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
			msg.ReplyMarkup = keyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)
		}
	}
}
