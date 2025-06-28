package main

import (
	"database/sql"
	"encoding/json"
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
					user, err = CreateUser(db, strconv.FormatInt(update.Message.Chat.ID, 10), update.Message.From.UserName)
					if err != nil {
						log.Fatal(err)
					}
					if err := EnsureStatTableReady(db, update.Message.Chat.ID); err != nil {
						log.Fatal(err)
					}
					m := "Добро пожаловать, для начала ответьте на вопрос."
					message, buttons, err := PrepareQuestionAndButtons(user, db, update.Message.Chat.ID, 12)
					if err != nil {
						log.Fatal(err)
					}
					m += message
					if err := saveUserLastMessage(user, db, m); err != nil {
						log.Printf("Ошибка при сохранении последнего сообщения: %v", err)
					}
					sendMessageWithKeyboard(bot, update.Message.Chat.ID, m, buttons)
					continue
				} else {
					log.Fatal(err)
				}
			}
			if "/repeat" == update.Message.Text {
				buttons, err := parseButtons(user.GetButtons())
				if err != nil {
					log.Fatal(err)
				}
				sendMessageWithKeyboard(bot, update.Message.Chat.ID, user.GetLastMessage(), buttons)
				continue
			}
			rightWord, rightWordStat, err := FindEngWord(db, update.Message.Chat.ID, user.GetAnswer())
			if err != nil {
				if err := EnsureStatTableReady(db, update.Message.Chat.ID); err != nil {
					log.Fatal(err)
				}
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
				rightWordStat.SetWin(rightWordStat.GetWin() + 1)
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
				rightWordStat.SetLos(rightWordStat.GetLos() + 1)
			}
			rightWord.Update(db)
			rightWordStat.Update(db, update.Message.Chat.ID)
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
							m += fmt.Sprintf("<b>%d. %s: %.2f%%</b>\n", place, usr.GetUsername(), float64(user.GetTotalTrue())/float64(user.GetTotalCount())*100)
							topFive = true
						} else {
							m += fmt.Sprintf("%d. %s: %.2f%%\n", place, usr.GetUsername(), usr.GetTotalRate())
						}
					} else if topFive {
						break
					} else if strconv.FormatInt(update.Message.Chat.ID, 10) == usr.GetID() {
						m += "--------------------------------\n"
						m += fmt.Sprintf("<b>%d. %s: %.2f%%</b>\n", place, usr.GetUsername(), usr.GetTotalRate())
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
			message, buttons, err := PrepareQuestionAndButtons(user, db, update.Message.Chat.ID, 12)
			if err != nil {
				log.Fatal(err)
			}
			m += message
			if err := saveUserLastMessage(user, db, m); err != nil {
				log.Printf("Ошибка при сохранении последнего сообщения: %v", err)
			}
			sendMessageWithKeyboard(bot, update.Message.Chat.ID, m, buttons)
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

func sendMessageWithKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string, buttons [][]tgbotapi.KeyboardButton) {
	keyboard := tgbotapi.NewReplyKeyboard(buttons...)
	keyboard.OneTimeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "HTML"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func PrepareQuestionAndButtons(user *User, db *DB, chatID int64, wordCount int) (string, [][]tgbotapi.KeyboardButton, error) {
	engWords, err := GetAllEngWordByStat(db, chatID, wordCount)
	if err != nil {
		return "", nil, err
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

	message := "\nВопрос: " + question + tran
	buttons := createTelegramButtons(engWords, buttonType)
	jsonButtons, err := json.Marshal(buttons)
	if err != nil {
		log.Printf("Ошибка сериализации кнопок: %v", err)
	} else {
		user.SetButtons(string(jsonButtons))
	}
	if err := user.Update(db); err != nil {
		return "", nil, err
	}
	return message, buttons, nil
}

func saveUserLastMessage(user *User, db *DB, message string) error {
	user.SetLastMessage(message)

	if err := user.Update(db); err != nil {
		log.Printf("Ошибка обновления пользователя: %v", err)
		return err
	}
	return nil
}

func parseButtons(buttons string) ([][]tgbotapi.KeyboardButton, error) {
	var rawButtons [][]map[string]string
	err := json.Unmarshal([]byte(buttons), &rawButtons)
	if err != nil {
		return nil, err
	}

	var keyboard [][]tgbotapi.KeyboardButton
	for _, row := range rawButtons {
		var buttonRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			if text, ok := btn["text"]; ok {
				buttonRow = append(buttonRow, tgbotapi.NewKeyboardButton(text))
			}
		}
		keyboard = append(keyboard, buttonRow)
	}

	return keyboard, nil
}
