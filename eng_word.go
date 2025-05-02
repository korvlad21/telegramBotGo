package main

import (
	"database/sql"
	"fmt"
)

type EngWord struct {
	ID   uint64
	Eng  string
	Tran string
	Rus  string
	Win  uint64
	Los  uint64
}

// Геттеры
func (e *EngWord) GetID() uint64 {
	return e.ID
}

func (e *EngWord) GetEng() string {
	return e.Eng
}

func (e *EngWord) GetTran() string {
	return e.Tran
}

func (e *EngWord) GetRus() string {
	return e.Rus
}

func (e *EngWord) GetWin() uint64 {
	return e.Win
}

func (e *EngWord) GetLos() uint64 {
	return e.Los
}

// Сеттеры
func (e *EngWord) SetID(id uint64) {
	e.ID = id
}

func (e *EngWord) SetEng(eng string) {
	e.Eng = eng
}

func (e *EngWord) SetTran(tran string) {
	e.Tran = tran
}

func (e *EngWord) SetRus(rus string) {
	e.Rus = rus
}

func (e *EngWord) SetWin(win uint64) {
	e.Win = win
}

func (e *EngWord) SetLos(los uint64) {
	e.Los = los
}

func GetAllEngWord(db *DB) ([]EngWord, error) {
	var (
		query string
		rows  *sql.Rows
		err   error
	)
	query = `SELECT id, eng, tran, rus, win, los FROM eng_word`
	rows, err = db.Connection.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var engWords []EngWord
	for rows.Next() {
		var engWord EngWord
		if err := rows.Scan(&engWord.ID, &engWord.Eng, &engWord.Tran, &engWord.Rus, &engWord.Win, &engWord.Los); err != nil {
			return nil, err
		}
		engWords = append(engWords, engWord)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return engWords, nil
}

func GetAllEngWordByStat(db *DB, chatId int64, limit int) ([]EngWord, error) {
	tableName := fmt.Sprintf("%d_word_stat", chatId)

	query := fmt.Sprintf(`
		SELECT
			ew.id,
			ew.eng,
			ew.tran,
			ew.rus
		FROM
			eng_word AS ew
			LEFT JOIN %s AS ws ON ws.word_id = ew.id
		ORDER BY
			RAND() * ((IFNULL(ws.los, 0) + 1) / (IFNULL(ws.win, 0) + 1)) DESC
		LIMIT ?
	`, tableName)

	rows, err := db.Connection.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var engWords []EngWord
	for rows.Next() {
		var engWord EngWord
		if err := rows.Scan(&engWord.ID, &engWord.Eng, &engWord.Tran, &engWord.Rus); err != nil {
			return nil, err
		}
		engWords = append(engWords, engWord)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return engWords, nil
}

func FindEngWord(db *DB, chatId int64, word string) (*EngWord, *WordStat, error) {
	tableName := fmt.Sprintf("%d_word_stat", chatId)

	query := fmt.Sprintf(`
		SELECT
			ew.id,
			ew.eng,
			ew.tran,
			ew.rus,
			IFNULL(ws.win, 0),
			IFNULL(ws.los, 0),
			IFNULL(ws.id, 0)
		FROM
			eng_word AS ew
			LEFT JOIN %s AS ws ON ws.word_id = ew.id
		WHERE
			ew.eng = ?
			OR ew.rus = ?
		LIMIT 1
	`, tableName)

	row := db.Connection.QueryRow(query, word, word)

	var (
		engWord  EngWord
		wordStat WordStat
	)

	if err := row.Scan(&engWord.ID, &engWord.Eng, &engWord.Tran, &engWord.Rus, &wordStat.Win, &wordStat.Los, &wordStat.ID); err != nil {
		return nil, nil, err
	}

	wordStat.WordId = engWord.ID

	return &engWord, &wordStat, nil
}

func (e *EngWord) Update(db *DB) error {
	query := `UPDATE eng_word SET eng = ?, tran = ?, rus = ?, win = ?, los = ? WHERE id = ?`
	_, err := db.Connection.Exec(query, e.Eng, e.Tran, e.Rus, e.Win, e.Los, e.ID)
	return err
}
