package main

import (
	"fmt"
	"log"
	"math/rand"
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
			// Клавиатура исчезнет после нажатия
			m := ""
			user, err := getUserByID(db, update.Message.Chat.ID)
			if err != nil {
				log.Fatal(err)
			}
			rightWord, err := FindEngWord(db, user.getAnswer())
			if err != nil {
				log.Fatal(err)
			}
			ansTran := ""
			if rightWord.GetEng() == user.getAnswer() {
				ansTran = rightWord.GetTran()
			}
			m += "Ответ: " + user.getAnswer() + ansTran + "\n"
			if update.Message.Text == user.getAnswer() {
				m += "Правильно!✅"
				rightWord.SetWin(rightWord.GetWin() + 1)
			} else {
				m += "Неправильно!!!!!⛔️"
				rightWord.SetLos(rightWord.GetLos() + 1)
			}
			rightWord.Update(db)
			m += "\n\n"
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
			engWords, err := GetAllEngWord(db, 12)
			if err != nil {
				log.Fatal(err)
			}
			answerWord := getAnswerWord(engWords)
			buttonType := "eng"
			answer := answerWord.Eng
			question := answerWord.Rus
			tran := ""
			if rand.Intn(2) == 1 {
				buttonType = "rus"
				answer = answerWord.Rus
				question = answerWord.Eng
				tran = " " + answerWord.Tran
			}
			user.SetAnswer(answer)
			user.SetQuestion(question)
			m += "\nВопрос: " + question + tran

			user.Update(db)

			buttons := createTelegramButtons(engWords, buttonType)

			// Создаём клавиатуру
			keyboard := tgbotapi.NewReplyKeyboard(buttons...)
			keyboard.OneTimeKeyboard = true
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
			msg.ReplyMarkup = keyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)
		}
	}
}

func createTelegramButtons(engWords []EngWord, buttonType string) [][]tgbotapi.KeyboardButton {
	var buttons [][]tgbotapi.KeyboardButton

	// Количество кнопок в строке
	buttonsPerRow := 3

	// Создаем кнопки
	for i := 0; i < len(engWords); i += buttonsPerRow {
		var row []tgbotapi.KeyboardButton
		for j := 0; j < buttonsPerRow && i+j < len(engWords); j++ {
			button := tgbotapi.NewKeyboardButton(engWords[i+j].Eng)
			if buttonType == "rus" {
				button = tgbotapi.NewKeyboardButton(engWords[i+j].Rus)
			}
			row = append(row, button)
		}
		buttons = append(buttons, row)
	}

	return buttons
}

func getAnswerWord(engWords []EngWord) EngWord {
	randomIndex := rand.Intn(len(engWords))
	return engWords[randomIndex]
}
