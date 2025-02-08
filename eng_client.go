package main

import (
	"database/sql"
)

type EngClient struct {
	id         string
	name       string
	cycleTrue  int
	cycleCount int
	totalTrue  int64
	totalCount int64
	question   sql.NullString
	answer     sql.NullString
	sets       int
	level      int
	totalRate  float64
}

// Геттеры
func (e *EngClient) ID() string {
	return e.id
}

func (e *EngClient) Name() string {
	return e.name
}

func (e *EngClient) CycleTrue() int {
	return e.cycleTrue
}

func (e *EngClient) CycleCount() int {
	return e.cycleCount
}

func (e *EngClient) TotalTrue() int64 {
	return e.totalTrue
}

func (e *EngClient) TotalRate() float64 {
	return e.totalRate
}

func (e *EngClient) TotalCount() int64 {
	return e.totalCount
}

func (e *EngClient) Question() sql.NullString {
	return e.question
}

func (e *EngClient) Answer() sql.NullString {
	return e.answer
}

func (e *EngClient) Sets() int {
	return e.sets
}

func (e *EngClient) Level() int {
	return e.level
}

// Сеттеры
func (e *EngClient) SetID(id string) {
	e.id = id
}

func (e *EngClient) SetName(name string) {
	e.name = name
}

func (e *EngClient) SetCycleTrue(cycleTrue int) {
	e.cycleTrue = cycleTrue
}

func (e *EngClient) SetCycleCount(cycleCount int) {
	e.cycleCount = cycleCount
}

func (e *EngClient) SetTotalTrue(totalTrue int64) {
	e.totalTrue = totalTrue
}

func (e *EngClient) SetTotalCount(totalCount int64) {
	e.totalCount = totalCount
}

func (e *EngClient) SetTotalRate(totalRate float64) {
	e.totalRate = totalRate
}

func (e *EngClient) SetQuestion(question sql.NullString) {
	e.question = question
}

func (e *EngClient) SetAnswer(answer sql.NullString) {
	e.answer = answer
}

func (e *EngClient) SetSets(sets int) {
	e.sets = sets
}

func (e *EngClient) SetLevel(level int) {
	e.level = level
}

// Функция для получения списка клиентов
func getEngClients(db *DB) ([]EngClient, error) {
	query := `SELECT id, name, cycle_true, cycle_count, total_true, total_count, question, answer, sets, level,
	                 CASE WHEN total_true = 0 THEN 0 ELSE total_count / total_true END AS total_rate
	          FROM eng_client
	          ORDER BY total_rate`
	rows, err := db.Connection.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []EngClient

	for rows.Next() {
		var client EngClient
		if err := rows.Scan(&client.id, &client.name, &client.cycleTrue, &client.cycleCount, &client.totalTrue, &client.totalCount, &client.question, &client.answer, &client.sets, &client.level, &client.totalRate); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

func getEngClientByID(db *DB, id int) (*EngClient, error) {
	query := `SELECT id, name, cycle_true, cycle_count, total_true, total_count, question, answer, sets, level, 
	                 CASE WHEN total_true = 0 THEN 0 ELSE total_count / total_true END AS ratio
	          FROM eng_client
	          WHERE id = ?`
	row := db.Connection.QueryRow(query, id)

	var client EngClient
	if err := row.Scan(&client.id, &client.name, &client.cycleTrue, &client.cycleCount, &client.totalTrue, &client.totalCount, &client.question, &client.answer, &client.sets, &client.level, &client.totalRate); err != nil {
		return nil, err
	}

	return &client, nil
}
