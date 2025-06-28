package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	ButtonsPerRow    = 3
	WordsPerQuestion = 12
	CycleLength      = 100
)

type Bot struct {
	api         *tgbotapi.BotAPI
	db          *DB
	redisClient *Redis
}

func NewBot() (*Bot, error) {
	db := NewDB()
	redisClient := NewRedis()

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("токен бота не найден в переменных окружения")
	}

	api, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %w", err)
	}

	api.Debug = true

	return &Bot{
		api:         api,
		db:          db,
		redisClient: redisClient,
	}, nil
}

func (b *Bot) Close() {
	if b.db != nil {
		b.db.Close()
	}
	if b.redisClient != nil {
		b.redisClient.Close()
	}
}

func (b *Bot) GetBotInfo() (string, int64) {
	return b.api.Self.UserName, b.api.Self.ID
}

func (b *Bot) createTelegramButtons(engWords []EngWord, buttonType string) [][]tgbotapi.KeyboardButton {
	var buttons [][]tgbotapi.KeyboardButton

	for i := 0; i < len(engWords); i += ButtonsPerRow {
		var row []tgbotapi.KeyboardButton
		for j := 0; j < ButtonsPerRow && i+j < len(engWords); j++ {
			var buttonText string
			if buttonType == "rus" {
				buttonText = engWords[i+j].Rus
			} else {
				buttonText = engWords[i+j].Eng
			}
			row = append(row, tgbotapi.NewKeyboardButton(buttonText))
		}
		buttons = append(buttons, row)
	}

	return buttons
}

func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, buttons [][]tgbotapi.KeyboardButton) {
	keyboard := tgbotapi.NewReplyKeyboard(buttons...)
	keyboard.OneTimeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "HTML"

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки сообщения с клавиатурой: %v", err)
	}
}

func (b *Bot) parseButtons(buttons string) ([][]tgbotapi.KeyboardButton, error) {
	if buttons == "" {
		return nil, fmt.Errorf("пустая строка кнопок")
	}

	var rawButtons [][]map[string]string
	if err := json.Unmarshal([]byte(buttons), &rawButtons); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
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
