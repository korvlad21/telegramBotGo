package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	log.Println("🤖 Бот начал обработку сообщений...")

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		}
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	log.Printf("📨 [%s] %s", message.From.UserName, message.Text)

	user, err := b.getOrCreateUser(message)
	if err != nil {
		log.Printf("❌ Ошибка получения пользователя: %v", err)
		b.sendErrorMessage(message.Chat.ID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if message.Text == "/repeat" {
		b.handleRepeatCommand(user, message.Chat.ID)
		return
	}

	b.processAnswerAndSendResponse(user, message)
}

func (b *Bot) getOrCreateUser(message *tgbotapi.Message) (*User, error) {
	user, err := getUserByID(b.db, message.Chat.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return b.createNewUser(message)
		}
		return nil, err
	}
	return user, nil
}

func (b *Bot) createNewUser(message *tgbotapi.Message) (*User, error) {
	log.Printf("👤 Создание нового пользователя: %s", message.From.UserName)

	user, err := CreateUser(b.db, strconv.FormatInt(message.Chat.ID, 10), message.From.UserName)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	if err := EnsureStatTableReady(b.db, message.Chat.ID); err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы статистики: %w", err)
	}

	welcomeMsg := "🎉 Добро пожаловать! Для начала ответьте на вопрос."
	questionMsg, buttons, err := b.prepareQuestion(user, message.Chat.ID)
	if err != nil {
		return nil, err
	}

	fullMessage := welcomeMsg + questionMsg
	if err := b.saveUserLastMessage(user, fullMessage); err != nil {
		log.Printf("⚠️ Ошибка при сохранении сообщения: %v", err)
	}

	b.sendMessageWithKeyboard(message.Chat.ID, fullMessage, buttons)
	return user, nil
}

func (b *Bot) prepareQuestion(user *User, chatID int64) (string, [][]tgbotapi.KeyboardButton, error) {
	engWords, err := GetAllEngWordByStat(b.db, chatID, WordsPerQuestion)
	if err != nil {
		return "", nil, fmt.Errorf("ошибка получения слов: %w", err)
	}

	if len(engWords) == 0 {
		return "", nil, fmt.Errorf("нет доступных слов для вопроса")
	}

	answerWord := b.getRandomWord(engWords)
	buttonType, answer, question, translation := b.determineQuestionType(answerWord)

	user.SetAnswer(answer)
	user.SetQuestion(question)

	if err := user.Update(b.db); err != nil {
		return "", nil, fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	message := fmt.Sprintf("\n❓ Вопрос: %s%s", question, translation)
	buttons := b.createTelegramButtons(engWords, buttonType)

	return message, buttons, nil
}

func (b *Bot) determineQuestionType(word EngWord) (buttonType, answer, question, translation string) {
	if rand.Intn(2) == 1 {
		return "rus", word.Rus, word.Eng, " (" + word.Tran + ")"
	}
	return "eng", word.Eng, word.Rus, ""
}

func (b *Bot) handleRepeatCommand(user *User, chatID int64) {
	log.Printf("🔄 Повтор сообщения для пользователя %s", user.GetUsername())

	buttons, err := b.parseButtons(user.GetButtons())
	if err != nil {
		log.Printf("❌ Ошибка парсинга кнопок: %v", err)
		b.sendErrorMessage(chatID, "Ошибка при восстановлении клавиатуры.")
		return
	}
	b.sendMessageWithKeyboard(chatID, user.GetLastMessage(), buttons)
}

func (b *Bot) processAnswerAndSendResponse(user *User, message *tgbotapi.Message) {
	rightWord, rightWordStat, err := FindEngWord(b.db, message.Chat.ID, user.GetAnswer())
	if err != nil {
		log.Printf("❌ Ошибка поиска слова: %v", err)
		b.sendErrorMessage(message.Chat.ID, "Произошла ошибка при поиске слова.")
		return
	}

	response := b.buildAnswerResponse(user, message.Text, rightWord, rightWordStat, message.Chat.ID)

	nextQuestion, buttons, err := b.prepareQuestion(user, message.Chat.ID)
	if err != nil {
		log.Printf("❌ Ошибка подготовки вопроса: %v", err)
		b.sendPlainMessage(message.Chat.ID, response+"Ошибка при подготовке следующего вопроса.")
		return
	}

	fullResponse := response + nextQuestion
	log.Printf("⚠️ %v", fullResponse)
	if err := b.saveUserLastMessage(user, fullResponse); err != nil {
		log.Printf("⚠️ Ошибка сохранения сообщения: %v", err)
	}

	jsonButtons, _ := json.Marshal(buttons)
	user.SetButtons(string(jsonButtons))

	b.sendMessageWithKeyboard(message.Chat.ID, fullResponse, buttons)
}

func (b *Bot) buildAnswerResponse(user *User, userAnswer string, rightWord *EngWord, stat *WordStat, chatID int64) string {
	var response string

	answerTranslation := ""
	if rightWord.GetEng() == user.GetAnswer() {
		answerTranslation = " (" + rightWord.GetTran() + ")"
	}
	response += fmt.Sprintf("💡 Ответ: %s%s\n", user.GetAnswer(), answerTranslation)

	user.SetCycleCount(user.GetCycleCount() + 1)
	user.SetTotalCount(user.GetTotalCount() + 1)

	if userAnswer == user.GetAnswer() {
		response += "<b>✅ Правильно!</b>\n"
		stat.SetWin(stat.GetWin() + 1)
		user.SetCycleTrue(user.GetCycleTrue() + 1)
		user.SetTotalTrue(user.GetTotalTrue() + 1)
		log.Printf("✅ Правильный ответ от %s", user.GetUsername())
	} else {
		response += "<b>❌ Неправильно!</b>\n"
		b.saveWrongAnswer(user, rightWord)
		stat.SetLos(stat.GetLos() + 1)
		log.Printf("❌ Неправильный ответ от %s: %s (правильно: %s)",
			user.GetUsername(), userAnswer, user.GetAnswer())
	}

	if err := rightWord.Update(b.db); err != nil {
		log.Printf("⚠️ Ошибка обновления слова: %v", err)
	}
	if err := stat.Update(b.db, chatID); err != nil {
		log.Printf("⚠️ Ошибка обновления статистики: %v", err)
	}

	response += b.buildStatistics(user, chatID)
	return response
}

func (b *Bot) saveWrongAnswer(user *User, word *EngWord) {
	key := user.GetID()
	entry := fmt.Sprintf("%s - %s", word.GetEng(), word.GetRus())

	if err := b.redisClient.ListPush(key, entry); err != nil {
		log.Printf("⚠️ Ошибка сохранения неправильного ответа в Redis: %v", err)
	}
}

func (b *Bot) buildStatistics(user *User, chatID int64) string {
	cycleRate := float64(user.GetCycleTrue()) / float64(user.GetCycleCount()) * 100
	totalRate := float64(user.GetTotalTrue()) / float64(user.GetTotalCount()) * 100

	stats := fmt.Sprintf("📊 Сет(%d): %.2f%%\n", user.GetCycleCount(), cycleRate)
	stats += fmt.Sprintf("📈 Всего(%d): %.2f%%\n", user.GetTotalCount(), totalRate)

	if user.GetCycleCount() == CycleLength {
		stats += b.buildCycleCompletionStats(user, chatID)
		user.SetCycleCount(0)
		user.SetCycleTrue(0)
		log.Printf("🏆 Пользователь %s завершил цикл", user.GetUsername())
	}

	if err := user.Update(b.db); err != nil {
		log.Printf("⚠️ Ошибка обновления пользователя: %v", err)
	}

	return stats
}

func (b *Bot) buildCycleCompletionStats(user *User, chatID int64) string {
	stats := "\n\n🏆 Цикл завершен! Рейтинг:\n"

	users, err := getUsers(b.db)
	if err != nil {
		log.Printf("❌ Ошибка получения пользователей: %v", err)
		return stats + "Ошибка получения рейтинга.\n"
	}

	currentUserID := strconv.FormatInt(chatID, 10)
	stats += b.buildLeaderboard(users, currentUserID, user)

	wrongAnswers, err := b.redisClient.ListGetAllAsString(user.GetID())
	if err != nil {
		log.Printf("⚠️ Ошибка получения неправильных ответов: %v", err)
	} else if wrongAnswers != "" {
		stats += "\n📝 Слова с ошибками:\n" + wrongAnswers + "\n"
	}

	if err := b.redisClient.Delete(user.GetID()); err != nil {
		log.Printf("⚠️ Ошибка очистки списка ошибок: %v", err)
	}

	return stats
}

func (b *Bot) buildLeaderboard(users []User, currentUserID string, currentUser *User) string {
	var leaderboard string
	place := 0
	topFive := false
	currentUserRate := float64(currentUser.GetTotalTrue()) / float64(currentUser.GetTotalCount()) * 100

	for _, usr := range users {
		place++
		if place <= 5 {
			if currentUserID == usr.GetID() {
				leaderboard += fmt.Sprintf("<b>🥇 %d. %s: %.2f%%</b>\n", place, usr.GetUsername(), currentUserRate)
				topFive = true
			} else {
				medal := b.getMedal(place)
				leaderboard += fmt.Sprintf("%s %d. %s: %.2f%%\n", medal, place, usr.GetUsername(), usr.GetTotalRate())
			}
		} else if topFive {
			break
		} else if currentUserID == usr.GetID() {
			leaderboard += "━━━━━━━━━━━━━━━━\n"
			leaderboard += fmt.Sprintf("<b>📍 %d. %s: %.2f%%</b>\n", place, usr.GetUsername(), usr.GetTotalRate())
			break
		}
	}

	return leaderboard
}

func (b *Bot) getMedal(place int) string {
	switch place {
	case 1:
		return "🥇"
	case 2:
		return "🥈"
	case 3:
		return "🥉"
	case 4, 5:
		return "🏅"
	default:
		return "📍"
	}
}
