package main

import "log"

type User struct {
	id         string
	name       string
	cycleTrue  uint
	cycleCount uint
	totalTrue  uint64
	totalCount uint64
	question   string
	answer     string
	sets       uint
	level      uint
	totalRate  float64
}

// Геттеры
func (e *User) getID() string {
	return e.id
}

func (e *User) getName() string {
	return e.name
}

func (e *User) getCycleTrue() uint {
	return e.cycleTrue
}

func (e *User) getCycleCount() uint {
	return e.cycleCount
}

func (e *User) getTotalTrue() uint64 {
	return e.totalTrue
}

func (e *User) getTotalRate() float64 {
	return e.totalRate
}

func (e *User) getTotalCount() uint64 {
	return e.totalCount
}

func (e *User) getQuestion() string {
	return e.question
}

func (e *User) getAnswer() string {
	return e.answer
}

func (e *User) getSets() uint {
	return e.sets
}

func (e *User) getLevel() uint {
	return e.level
}

// Сеттеры
func (e *User) SetID(id string) {
	e.id = id
}

func (e *User) SetName(name string) {
	e.name = name
}

func (e *User) SetCycleTrue(cycleTrue uint) {
	e.cycleTrue = cycleTrue
}

func (e *User) SetCycleCount(cycleCount uint) {
	e.cycleCount = cycleCount
}

func (e *User) SetTotalTrue(totalTrue uint64) {
	e.totalTrue = totalTrue
}

func (e *User) SetTotalCount(totalCount uint64) {
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

func (e *User) SetSets(sets uint) {
	e.sets = sets
}

func (e *User) SetLevel(level uint) {
	e.level = level
}

// Функция для получения списка клиентов
func getUsers(db *DB) ([]User, error) {
	query := `SELECT id, name, cycle_true, cycle_count, total_true, total_count, question, answer, sets, level,
	                 CASE WHEN total_true = 0 THEN 0 ELSE 100 * (total_true / total_count) END AS total_rate
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
		if err := rows.Scan(&user.id, &user.name, &user.cycleTrue, &user.cycleCount, &user.totalTrue, &user.totalCount, &user.question, &user.answer, &user.sets, &user.level, &user.totalRate); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func getUserByID(db *DB, id int64) (*User, error) {
	query := `SELECT id, name, cycle_true, cycle_count, total_true, total_count, question, answer, sets, level, 
	                 CASE WHEN total_true = 0 THEN 0 ELSE total_count / total_true END AS ratio
	          FROM users
	          WHERE id = ?`
	row := db.Connection.QueryRow(query, id)

	var user User
	if err := row.Scan(&user.id, &user.name, &user.cycleTrue, &user.cycleCount, &user.totalTrue, &user.totalCount, &user.question, &user.answer, &user.sets, &user.level, &user.totalRate); err != nil {
		return nil, err
	}

	return &user, nil
}

func (e *User) Update(db *DB) error {
	query := `UPDATE users SET name = ?, cycle_true = ?, cycle_count = ?, total_true = ?, total_count = ?, 
				question = ?, answer = ?, sets = ?, level = ? WHERE id = ?`
	_, err := db.Connection.Exec(query, e.name, e.cycleTrue, e.cycleCount, e.totalTrue, e.totalCount,
		e.question, e.answer, e.sets, e.level, e.id)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
