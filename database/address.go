package database

import (
	"database/sql"
	"fmt"
	"log"
)

func createAddressTableIfNotExists(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS address (
		id INT AUTO_INCREMENT PRIMARY KEY,
		address VARCHAR(100) NOT NULL UNIQUE,
		hex_address VARCHAR(100) NOT NULL UNIQUE,
		platform VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}
	log.Println("✅ address table ensured.")
	return nil
}

func InsertAddress(db *sql.DB, addr string, hexAddr string, platform string) error {
	query := `INSERT INTO address(address, hex_address, platform) VALUES (?, ?, ?)`
	_, err := db.Exec(query, addr, hexAddr, platform)
	return err
}

func GetAddressesByPlatform(db *sql.DB, platform string) (map[string]struct{}, error) {
	query := `SELECT address FROM address WHERE platform = ?`

	rows, err := db.Query(query, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addressSet := make(map[string]struct{})
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, err
		}
		addressSet[addr] = struct{}{}
	}
	return addressSet, nil
}
