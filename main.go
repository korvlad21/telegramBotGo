package main

import (
	"database/sql"
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
	db := NewDB()
	defer db.Close()
	redisClient := NewRedis()

	defer redisClient.Close() // Закроем соединение при завершении

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
				if err == sql.ErrNoRows {
					continue
				}
				log.Fatal(err)
			}
			empty, err := IsStatTableEmpty(db, update.Message.Chat.ID)
			if err != nil {
				// Если ошибка — таблица, вероятно, не существует => создаём и заполняем
				log.Printf("Таблица не найдена, создаём заново: %v", err)

				err = CreateStatTable(db, update.Message.Chat.ID)
				if err != nil {
					log.Fatal(err)
				}

				err = FillStatTableFromEngWords(db, update.Message.Chat.ID)
				if err != nil {
					log.Fatal(err)
				}
				log.Fatal("2")
			} else if empty {
				// Таблица есть, но она пустая — просто заполняем
				err = FillStatTableFromEngWords(db, update.Message.Chat.ID)
				if err != nil {
					log.Fatal(err)
				}
				log.Fatal("3")
			}
			log.Fatal("4")
			rightWord, err := FindEngWord(db, user.GetAnswer())
			if err != nil {
				log.Fatal(err)
			}
			ansTran := ""
			if rightWord.GetEng() == user.GetAnswer() {
				ansTran = rightWord.GetTran()
			}
			m += "Ответ: " + user.GetAnswer() + ansTran + "\n"
			user.SetCycleCount(user.GetCycleCount() + 1)
			user.SetTotalCount(user.GetTotalCount() + 1)
			if update.Message.Text == user.GetAnswer() {
				m += "<b>Правильно!</b>✅"
				rightWord.SetWin(rightWord.GetWin() + 1)
				user.SetCycleTrue(user.GetCycleTrue() + 1)
				user.SetTotalTrue(user.GetTotalTrue() + 1)
			} else {
				m += "<b>Неправильно!!!!!</b>⛔️"
				key := fmt.Sprintf("%s", user.GetID())
				newEntry := fmt.Sprintf("%s - %s", rightWord.GetEng(), rightWord.GetRus())

				// Добавляем новый элемент в список
				if err := redisClient.ListPush(key, newEntry); err != nil {
					log.Printf("Ошибка добавления данных в Redis: %v", err)
				}
				rightWord.SetLos(rightWord.GetLos() + 1)
			}
			rightWord.Update(db)
			m += "\n"
			m += fmt.Sprintf("Сет(%d): %.2f%%\n", user.GetCycleCount(), float64(user.GetCycleTrue())/float64(user.GetCycleCount())*100)
			m += fmt.Sprintf("Всего(%d): %.2f%%\n", user.GetTotalCount(), float64(user.GetTotalTrue())/float64(user.GetTotalCount())*100)
			if 100 == int(user.GetCycleCount()) {
				m += "\n\n"
				user.SetCycleCount(0)
				user.SetCycleTrue(0)
				users, err := getUsers(db)
				if err != nil {
					log.Fatal(err)
				}
				place := 0
				topFive := false
				for _, usr := range users {
					place++
					if place <= 5 {
						if strconv.FormatInt(update.Message.Chat.ID, 10) == usr.GetID() {
							m += fmt.Sprintf("<b>%d. %s: %.2f%%</b>\n", place, usr.GetName(), usr.GetTotalRate())
							topFive = true
						} else {
							m += fmt.Sprintf("%d. %s: %.2f%%\n", place, usr.GetName(), usr.GetTotalRate())
						}
					} else if topFive {
						break
					} else if strconv.FormatInt(update.Message.Chat.ID, 10) == usr.GetID() {
						m += "--------------------------------\n"
						m += fmt.Sprintf("<b>%d. %s: %.2f%%</b>\n", place, usr.GetName(), usr.GetTotalRate())
						topFive = true
					}
				}
				str, err := redisClient.ListGetAllAsString(fmt.Sprintf("%s", user.GetID()))
				if err != nil {
					log.Println("Ошибка при получении данных из Redis:", err)
				} else {
					m += "\nСлова, в которых вы допустили ошибку:\n" + str + "\n\n"
				}
				redisClient.Delete(fmt.Sprintf("%s", user.GetID()))
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
