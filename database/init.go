package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func SetDB(d *sql.DB) {
	db = d
}

func GetDB() *sql.DB {
	return db
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func loadConfig() DBConfig {
	return DBConfig{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetString("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
	}
}

func InitDatabase() *sql.DB {
	config := loadConfig()

	// 先嘗試連接目標 DB
	dbConn, err := connectWithRetry(config, true)

	// 若失敗且是 "Unknown database" 則建立
	if err != nil && strings.Contains(err.Error(), "Unknown database") {
		log.Printf("Database %s not found. Creating...\n", config.DBName)
		var sysDB *sql.DB
		sysDB, err = connectWithRetry(config, false)
		if err != nil {
			log.Fatalf("Failed to connect to MySQL system DB: %v", err)
		}
		defer sysDB.Close()

		_, err = sysDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", config.DBName))
		if err != nil {
			log.Fatalf("Failed to create DB: %v", err)
		}
		log.Printf("Created DB: %s\n", config.DBName)

		dbConn, err = connectWithRetry(config, true)
	}

	// 若還是有錯，直接失敗
	if err != nil {
		log.Fatalf("Final DB connect error: %v", err)
	}

	log.Printf("Connected to database: %s\n", config.DBName)

	// 一次寫 ensureTables()
	if err := ensureTables(dbConn); err != nil {
		log.Fatalf("Failed to ensure tables: %v", err)
	}
	SetDB(dbConn)
	return db
}

func connectWithRetry(cfg DBConfig, connectToTarget bool) (*sql.DB, error) {
	dsn := buildDSN(cfg, connectToTarget)
	var err error
	var dbconn *sql.DB
	for i := 0; i < 10; i++ {
		dbconn, err = sql.Open("mysql", dsn)
		if err == nil {
			err = dbconn.Ping()
		}
		if err == nil {
			return dbconn, nil
		}
		log.Printf("Retry %d: DB not ready yet... dsn: %s, err: %v", i+1, dsn, err)
		if strings.Contains(err.Error(), "Unknown database") {
			break
		}
		time.Sleep(3 * time.Second)
	}

	return nil, err
}

func buildDSN(cfg DBConfig, useTarget bool) string {
	dbName := "sys"
	if useTarget {
		dbName = cfg.DBName
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FTaipei",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, dbName)
}

func ensureTables(db *sql.DB) (err error) {
	err = createEventHistoryTableIfNotExists(db)
	if err != nil {
		return
	}
	err = createDelegateTableIfNotExists(db)
	if err != nil {
		return
	}
	err = createAddressTableIfNotExists(db)
	if err != nil {
		return
	}
	err = createChatTableIfNotExists(db)
	return
}
