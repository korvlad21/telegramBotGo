package main

import (
	"fmt"
	"log"
)

type WordStat struct {
	ID     uint64
	WordId uint64
	Win    uint64
	Los    uint64
}

// Геттеры
func (e *WordStat) GetID() uint64 {
	return e.ID
}

func (e *WordStat) GetWordId() uint64 {
	return e.WordId
}

func (e *WordStat) GetWin() uint64 {
	return e.Win
}

func (e *WordStat) GetLos() uint64 {
	return e.Los
}

// Сеттеры
func (e *WordStat) SetID(id uint64) {
	e.ID = id
}

func (e *WordStat) SetWordId(wordId uint64) {
	e.WordId = wordId
}

func (e *WordStat) SetWin(win uint64) {
	e.Win = win
}

func (e *WordStat) SetLos(los uint64) {
	e.Los = los
}

func CreateStatTable(db *DB, chatId int64) error {
	// Имя таблицы
	tableName := fmt.Sprintf("%d_word_stat", chatId)

	// SQL запрос для создания таблицы
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id BIGINT(20) UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		word_id BIGINT(20) UNSIGNED NOT NULL,
		win BIGINT(20) UNSIGNED DEFAULT 0 NOT NULL,
		los BIGINT(20) UNSIGNED DEFAULT 0 NOT NULL,
		INDEX idx_word_id (word_id),
		FOREIGN KEY (word_id) REFERENCES eng_word(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`, tableName)

	// Выполняем запрос
	_, err := db.Connection.Exec(query)
	if err != nil {
		return fmt.Errorf("ошибка при создании таблицы %s: %v", tableName, err)
	}

	log.Printf("Таблица %s успешно создана с внешним ключом на eng_word!", tableName)
	return nil
}

func FillStatTableFromEngWords(db *DB, chatId int64) error {
	// Получаем все английские слова
	words, err := GetAllEngWord(db)
	if err != nil {
		return fmt.Errorf("не удалось получить слова: %v", err)
	}

	// Имя таблицы
	tableName := fmt.Sprintf("%d_word_stat", chatId)

	// Подготавливаем SQL-запрос
	query := fmt.Sprintf(`INSERT INTO %s (word_id, win, los) VALUES (?, ?, ?)`, tableName)

	// Вставляем каждое слово
	for _, word := range words {
		_, err := db.Connection.Exec(query, word.ID, word.Win, word.Los)
		if err != nil {
			return fmt.Errorf("ошибка вставки слова ID %d: %v", word.ID, err)
		}
	}

	log.Printf("Таблица %s успешно заполнена из eng_word!", tableName)
	return nil
}

func IsStatTableEmpty(db *DB, chatId int64) (bool, error) {
	tableName := fmt.Sprintf("%d_word_stat", chatId)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tableName)

	var count int
	err := db.Connection.QueryRow(query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("не удалось проверить количество записей в таблице %s: %v", tableName, err)
	}

	return count == 0, nil
}

func EnsureStatTableReady(db *DB, chatId int64) error {
	empty, err := IsStatTableEmpty(db, chatId)
	if err != nil {
		log.Printf("Таблица не найдена, создаём заново: %v", err)

		err = CreateStatTable(db, chatId)
		if err != nil {
			return fmt.Errorf("ошибка создания таблицы: %v", err)
		}

		err = FillStatTableFromEngWords(db, chatId)
		if err != nil {
			return fmt.Errorf("ошибка заполнения таблицы: %v", err)
		}
	} else if empty {
		err = FillStatTableFromEngWords(db, chatId)
		if err != nil {
			return fmt.Errorf("ошибка заполнения таблицы: %v", err)
		}
	}
	return nil
}

func (ws *WordStat) Update(db *DB, chatId int64) error {
	tableName := fmt.Sprintf("%d_word_stat", chatId)

	query := fmt.Sprintf(`
		UPDATE %s 
		SET win = ?, los = ?
		WHERE id = ?
	`, tableName)

	_, err := db.Connection.Exec(query, ws.Win, ws.Los, ws.ID)
	if err != nil {
		return fmt.Errorf("не удалось обновить статистику слова с id %d: %v", ws.ID, err)
	}

	return nil
}
