package config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbPasswordRedactor = regexp.MustCompile(`password=\S+`)

func sanitizeDBError(err error) string {
	if err == nil {
		return ""
	}
	return dbPasswordRedactor.ReplaceAllString(err.Error(), "password=REDACTED")
}

func NewPostgresWrite() *gorm.DB {
	return newPostgres("write")
}

func NewPostgresRead() *gorm.DB {
	return newPostgres("read")
}

func newPostgres(role string) *gorm.DB {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, sslmode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("postgres connect failed role=%s: %s", role, sanitizeDBError(err))
		panic("postgres connect failed")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("sql.DB handle failed role=%s: %s", role, sanitizeDBError(err))
		panic("sql.DB handle failed")
	}
	if role == "read" {
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(25)
	} else {
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(12)
	}
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db
}
