package database

import (
	"database/sql"
	"time"
)

func createEventHistoryTableIfNotExists(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS event_history (
		id INT AUTO_INCREMENT PRIMARY KEY,
		transactionHash VARCHAR(100) NOT NULL,
		usdt VARCHAR(100) NOT NULL,
		fromAddress VARCHAR(100) NOT NULL,
		toAddress VARCHAR(100) NOT NULL,
		platform VARCHAR(50) NOT NULL,
		walletUsdt VARCHAR(100),
		orderId VARCHAR(100),
		askEnergySuccess BOOLEAN DEFAULT FALSE,
		createTime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_to_address (toAddress),
		INDEX idx_create_ask_platform (createTime, askEnergySuccess, platform)
	);
	`
	_, err := db.Exec(query)
	return err
}

func InsertEventHistory(db *sql.DB, ev EventHistory) error {
	query := `
	INSERT INTO event_history (
		transactionHash, usdt, fromAddress, toAddress,
		platform, walletUsdt, orderId, askEnergySuccess
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query,
		ev.TransactionHash, ev.USDT, ev.FromAddress, ev.ToAddress,
		ev.Platform, ev.WalletUSDT, ev.OrderID, ev.AskEnergySuccess)
	return err
}

func GetTodayEventCountGroupByPlatform(db *sql.DB) (map[string]int, error) {
	localEnd := time.Now()
	localStart := localEnd.Add(-24 * time.Hour)

	query := `
	SELECT platform, COUNT(*)
	FROM event_history
	WHERE createTime >= ? AND createTime < ?
	AND	askEnergySuccess = true
	GROUP BY platform
	`

	rows, err := db.Query(query, localStart, localEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var platform string
		var count int
		if err := rows.Scan(&platform, &count); err != nil {
			return nil, err
		}
		result[platform] = count
	}
	return result, nil
}

func GetEventsByToAddress(db *sql.DB, to string) ([]EventHistory, error) {
	query := `
	SELECT transactionHash, usdt, fromAddress, toAddress,
	       platform, walletUsdt, orderId, askEnergySuccess
	FROM event_history
	WHERE toAddress = ?
	ORDER BY createTime DESC
	`

	rows, err := db.Query(query, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []EventHistory
	for rows.Next() {
		var ev EventHistory
		if err := rows.Scan(
			&ev.TransactionHash, &ev.USDT, &ev.FromAddress, &ev.ToAddress,
			&ev.Platform, &ev.WalletUSDT, &ev.OrderID, &ev.AskEnergySuccess); err != nil {
			return nil, err
		}
		result = append(result, ev)
	}
	return result, nil
}
