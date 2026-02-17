package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/vterry/food-ordering/catalog/internal/infra/config"
)

func NewDataBase(cfg config.DbConfig) (*sql.DB, error) {
	driverConfig := mysql.Config{
		User:                 cfg.User,
		Passwd:               cfg.Password,
		Addr:                 cfg.Address,
		DBName:               cfg.Name,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	db, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed into open connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
