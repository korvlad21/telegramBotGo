package main

import "database/sql"

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

func GetAllEngWord(db *DB, limit int) ([]EngWord, error) {
	var (
		query string
		rows  *sql.Rows
		err   error
	)

	if limit == 0 {
		query = `SELECT id, eng, tran, rus, win, los FROM eng_word`
		rows, err = db.Connection.Query(query)
	} else {
		query = `SELECT id, eng, tran, rus, win, los FROM eng_word ORDER BY RAND() * ((los + 1) / (win + 1)) DESC LIMIT ?`
		rows, err = db.Connection.Query(query, limit)
	}

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

func FindEngWord(db *DB, word string) (*EngWord, error) {
	query := `SELECT id, eng, tran, rus, win, los FROM eng_word WHERE eng = ? OR rus = ? LIMIT 1`
	row := db.Connection.QueryRow(query, word, word)

	var engWord EngWord
	if err := row.Scan(&engWord.ID, &engWord.Eng, &engWord.Tran, &engWord.Rus, &engWord.Win, &engWord.Los); err != nil {
		return nil, err
	}

	return &engWord, nil
}

func (e *EngWord) Update(db *DB) error {
	query := `UPDATE eng_word SET eng = ?, tran = ?, rus = ?, win = ?, los = ? WHERE id = ?`
	_, err := db.Connection.Exec(query, e.Eng, e.Tran, e.Rus, e.Win, e.Los, e.ID)
	return err
}
