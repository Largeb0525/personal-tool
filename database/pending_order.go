package database

import (
	"database/sql"
	"fmt"
	"log"
)

func createPendingOrderTableIfNotExists(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS pending_order (
		merchant_order_id VARCHAR(255) PRIMARY KEY,
		customer_username VARCHAR(255) NOT NULL,
		advertiser_username VARCHAR(255) NOT NULL,
		order_status VARCHAR(50) NOT NULL,
		display_fiat_amount DOUBLE NOT NULL,
		retries INT DEFAULT 0,
		original_chat_id BIGINT NOT NULL,
		reply_to_message_id BIGINT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("create table 'pending_order' failed: %w", err)
	}
	log.Println("✅ pending_order table ensured.")
	return nil
}

func InsertPendingOrder(db *sql.DB, order PendingOrder) error {
	query := `
	INSERT INTO pending_order (
		merchant_order_id, customer_username, advertiser_username,
		order_status, display_fiat_amount, retries, original_chat_id, reply_to_message_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		customer_username = VALUES(customer_username),
		advertiser_username = VALUES(advertiser_username),
		order_status = VALUES(order_status),
		display_fiat_amount = VALUES(display_fiat_amount),
		retries = VALUES(retries),
		original_chat_id = VALUES(original_chat_id),
		reply_to_message_id = VALUES(reply_to_message_id);
	`
	_, err := db.Exec(query,
		order.MerchantOrderID, order.CustomerUsername, order.AdvertiserUsername,
		order.OrderStatus, order.DisplayFiatAmount, order.Retries, order.OriginalChatID, order.ReplyToMessageID)
	return err
}

func GetPendingOrders(db *sql.DB) ([]PendingOrder, error) {
	query := `SELECT merchant_order_id, customer_username, advertiser_username, order_status, display_fiat_amount, retries, original_chat_id, reply_to_message_id, created_at, updated_at FROM pending_order`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []PendingOrder
	for rows.Next() {
		var order PendingOrder
		if err := rows.Scan(
			&order.MerchantOrderID, &order.CustomerUsername, &order.AdvertiserUsername,
			&order.OrderStatus, &order.DisplayFiatAmount, &order.Retries, &order.OriginalChatID, &order.ReplyToMessageID,
			&order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func UpdatePendingOrderStatus(db *sql.DB, merchantOrderID string, newStatus string) error {
	query := `UPDATE pending_order SET order_status = ?, updated_at = CURRENT_TIMESTAMP WHERE merchant_order_id = ?`
	_, err := db.Exec(query, newStatus, merchantOrderID)
	return err
}

func IncrementPendingOrderRetries(db *sql.DB, merchantOrderID string) error {
	query := `UPDATE pending_order SET retries = retries + 1, updated_at = CURRENT_TIMESTAMP WHERE merchant_order_id = ?`
	_, err := db.Exec(query, merchantOrderID)
	return err
}

func DeletePendingOrder(db *sql.DB, merchantOrderID string) error {
	query := `DELETE FROM pending_order WHERE merchant_order_id = ?`
	_, err := db.Exec(query, merchantOrderID)
	return err
}
