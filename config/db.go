package config

import (
	"database/sql"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var SQLDB *sql.DB

func ConnectDB() {
	dsn := "root:@tcp(127.0.0.1:3306)/dbcloud?parseTime=true"

	// GORM
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal connect GORM")
	}
	DB = gormDB

	// SQL (buat auth lama)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Gagal connect SQL")
	}
	SQLDB = sqlDB
}