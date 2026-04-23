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

	ensureFileSystemColumns()
	ensureUserProfileColumns()
}

func ensureFileSystemColumns() {
	if !columnExists("dbcloud", "tb_files", "parent_id") {
		if _, err := SQLDB.Exec("ALTER TABLE tb_files ADD COLUMN parent_id BIGINT UNSIGNED NULL AFTER user_id"); err != nil {
			log.Fatal("Gagal menambahkan kolom parent_id pada tb_files")
		}
	}

	if !columnExists("dbcloud", "tb_files", "is_folder") {
		if _, err := SQLDB.Exec("ALTER TABLE tb_files ADD COLUMN is_folder TINYINT(1) NOT NULL DEFAULT 0 AFTER mime_type"); err != nil {
			log.Fatal("Gagal menambahkan kolom is_folder pada tb_files")
		}
	}
}

func columnExists(schemaName, tableName, columnName string) bool {
	const query = `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`

	var count int
	if err := SQLDB.QueryRow(query, schemaName, tableName, columnName).Scan(&count); err != nil {
		log.Fatal("Gagal cek struktur tabel")
	}

	return count > 0
}

func ensureUserProfileColumns() {
	if !columnExists("dbcloud", "tb_users", "name") {
		if _, err := SQLDB.Exec("ALTER TABLE tb_users ADD COLUMN name VARCHAR(100) NULL AFTER id"); err != nil {
			log.Fatal("Gagal menambahkan kolom name pada tb_users")
		}
	}

	if !columnExists("dbcloud", "tb_users", "phone") {
		if _, err := SQLDB.Exec("ALTER TABLE tb_users ADD COLUMN phone VARCHAR(30) NULL AFTER email"); err != nil {
			log.Fatal("Gagal menambahkan kolom phone pada tb_users")
		}
	}

	if !columnExists("dbcloud", "tb_users", "avatar_path") {
		if _, err := SQLDB.Exec("ALTER TABLE tb_users ADD COLUMN avatar_path VARCHAR(255) NULL AFTER phone"); err != nil {
			log.Fatal("Gagal menambahkan kolom avatar_path pada tb_users")
		}
	}
}
