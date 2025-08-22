package database

import (
	"database/sql"
	"fmt"
	"log"
)

func createChatTableIfNotExists(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS chat (
		id BIGINT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("create table 'chat' failed: %w", err)
	}
	log.Println("✅ chat table ensured.")
	return nil
}

func InsertOrUpdateChat(db *sql.DB, chatID int64, title string) error {
	query := `
	INSERT INTO chat (id, title) VALUES (?, ?)
	ON DUPLICATE KEY UPDATE title = ?;`
	_, err := db.Exec(query, chatID, title, title)
	return err
}

func DeleteChat(db *sql.DB, chatID int64) error {
	query := `DELETE FROM chat WHERE id = ?`
	_, err := db.Exec(query, chatID)
	return err
}

func GetAllChats(db *sql.DB) ([]Chat, error) {
	query := `SELECT id, title FROM chat`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var chat Chat
		if err := rows.Scan(&chat.ID, &chat.Title); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	return chats, nil
}

func GetChatByTitle(db *sql.DB, searchString string) ([]Chat, error) {
	query := `SELECT id, title FROM chat WHERE title LIKE ?`
	rows, err := db.Query(query, "%"+searchString+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var chat Chat
		if err := rows.Scan(&chat.ID, &chat.Title); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	return chats, nil
}
