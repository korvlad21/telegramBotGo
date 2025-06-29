package main

import (
	"encoding/json"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type User struct {
	id          string
	username    string
	cycleTrue   uint
	cycleCount  uint
	totalTrue   uint
	totalCount  uint
	question    string
	answer      string
	buttons     string
	lastMessage string
	sets        uint
	level       uint
	totalRate   float64
}

func (e *User) GetID() string {
	return e.id
}

func (e *User) GetUsername() string {
	return e.username
}

func (e *User) GetCycleTrue() uint {
	return e.cycleTrue
}

func (e *User) GetCycleCount() uint {
	return e.cycleCount
}

func (e *User) GetTotalTrue() uint {
	return e.totalTrue
}

func (e *User) GetTotalRate() float64 {
	return e.totalRate
}

func (e *User) GetTotalCount() uint {
	return e.totalCount
}

func (e *User) GetQuestion() string {
	return e.question
}

func (e *User) GetAnswer() string {
	return e.answer
}

func (e *User) GetButtons() string {
	return e.buttons
}

func (e *User) GetLastMessage() string {
	return e.lastMessage
}

func (e *User) GetSets() uint {
	return e.sets
}

func (e *User) GetLevel() uint {
	return e.level
}

func (e *User) SetID(id string) {
	e.id = id
}

func (e *User) SetUsername(username string) {
	e.username = username
}

func (e *User) SetCycleTrue(cycleTrue uint) {
	e.cycleTrue = cycleTrue
}

func (e *User) SetCycleCount(cycleCount uint) {
	e.cycleCount = cycleCount
}

func (e *User) SetTotalTrue(totalTrue uint) {
	e.totalTrue = totalTrue
}

func (e *User) SetTotalCount(totalCount uint) {
	e.totalCount = totalCount
}

func (e *User) SetTotalRate(totalRate float64) {
	e.totalRate = totalRate
}

func (e *User) SetQuestion(question string) {
	e.question = question
}

func (e *User) SetAnswer(answer string) {
	e.answer = answer
}

func (e *User) SetButtons(buttons string) {
	e.buttons = buttons
}

func (e *User) SetLastMessage(lastMessage string) {
	e.lastMessage = lastMessage
}

func (e *User) SetSets(sets uint) {
	e.sets = sets
}

func (e *User) SetLevel(level uint) {
	e.level = level
}

func getUsers(db *DB) ([]User, error) {
	query := `
		SELECT id, username, 
		       CASE WHEN total_count = 0 THEN 0 ELSE 100 * (total_true / total_count) END AS total_rate
		FROM users
		ORDER BY total_rate DESC`
	rows, err := db.Connection.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.id, &user.username, &user.totalRate); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func getUserByID(db *DB, id int64) (*User, error) {
	query := `SELECT id, username, cycle_true, cycle_count, total_true, total_count, question, answer, buttons, last_message, sets, level, 
	                 CASE WHEN total_true = 0 THEN 0 ELSE total_count / total_true END AS ratio
	          FROM users
	          WHERE id = ?`
	row := db.Connection.QueryRow(query, id)

	var user User
	if err := row.Scan(&user.id, &user.username, &user.cycleTrue, &user.cycleCount, &user.totalTrue, &user.totalCount, &user.question, &user.answer, &user.buttons, &user.lastMessage, &user.sets, &user.level, &user.totalRate); err != nil {
		return nil, err
	}

	return &user, nil
}

func (e *User) Update(db *DB) error {
	query := `UPDATE users SET username = ?, cycle_true = ?, cycle_count = ?, total_true = ?, total_count = ?, 
				question = ?, answer = ?, buttons = ?, last_message = ?, sets = ?, level = ? WHERE id = ?`
	_, err := db.Connection.Exec(query, e.username, e.cycleTrue, e.cycleCount, e.totalTrue, e.totalCount,
		e.question, e.answer, e.buttons, e.lastMessage, e.sets, e.level, e.id)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func CreateUser(db *DB, chatId string, username string) (*User, error) {
	query := `INSERT INTO users (id, username, cycle_true, cycle_count, total_true, total_count, question, answer, buttons, last_message, sets, level)
	          VALUES (?, ?, 0, 0, 0, 0, '', '', '', '', 0, 1)`
	_, err := db.Connection.Exec(query, chatId, username)
	if err != nil {
		return nil, err
	}

	user := &User{
		id:          chatId,
		username:    username,
		cycleTrue:   0,
		cycleCount:  0,
		totalTrue:   0,
		totalCount:  0,
		question:    "",
		answer:      "",
		buttons:     "",
		lastMessage: "",
		sets:        0,
		level:       1,
		totalRate:   0,
	}

	return user, nil
}

func (b *Bot) saveUserLastMessage(user *User, message string, buttons [][]tgbotapi.KeyboardButton) error {
	user.SetLastMessage(message)
	jsonButtons, _ := json.Marshal(buttons)
	user.SetButtons(string(jsonButtons))
	return user.Update(b.db)
}

func (e *User) DebugSQL() string {
	return fmt.Sprintf(
		`UPDATE users SET username = '%s', cycle_true = %d, cycle_count = %d, total_true = %d, total_count = %d, 
question = '%s', answer = '%s', buttons = '%s', last_message = '%s', sets = %d, level = %d WHERE id = '%s'`,
		e.username, e.cycleTrue, e.cycleCount, e.totalTrue, e.totalCount,
		e.question, e.answer, e.buttons, e.lastMessage, e.sets, e.level, e.id,
	)
}
