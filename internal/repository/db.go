package repository

import (
	"database/sql"
	"fmt"

	"game-server/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() error {
	cfg := config.GlobalConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
	)

	var err error
	DB, err = sql.Open(cfg.Driver, dsn)
	if err != nil {
		return err
	}

	if err := DB.Ping(); err != nil {
		return err
	}

	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(10)

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
