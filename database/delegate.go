package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func createDelegateTableIfNotExists(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS delegate (
		id INT AUTO_INCREMENT PRIMARY KEY,
		receiver_address VARCHAR(100) NOT NULL,
		txid VARCHAR(100) NOT NULL UNIQUE,
		undelegated BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_undelegated_created_at (undelegated, created_at)
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}
	log.Println("✅ delegate table ensured.")
	return nil
}

func InsertDelegateRecord(db *sql.DB, receiverAddress, txid string) error {
	query := `
	INSERT INTO delegate (receiver_address, txid)
	VALUES (?, ?);`
	_, err := db.Exec(query, receiverAddress, txid)
	return err
}

func GetUndelegatedBefore(db *sql.DB, before time.Time) ([]DelegateRecord, error) {
	query := `
	SELECT id, receiver_address, txid, undelegated, created_at
	FROM delegate
	WHERE undelegated = FALSE AND created_at < ?;`

	rows, err := db.Query(query, before)
	if err != nil {
		return nil, fmt.Errorf("query undelegated failed: %w", err)
	}
	defer rows.Close()

	var records []DelegateRecord
	for rows.Next() {
		var r DelegateRecord
		if err := rows.Scan(&r.ID, &r.ReceiverAddress, &r.TxID, &r.Undelegated, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan row failed: %w", err)
		}
		records = append(records, r)
	}
	return records, nil
}
