package db

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string
	Password string
}

type Database struct {
	conn *gorm.DB
}

func NewDatabase() *Database {
	db, err := newDBConn("user.db")
	if err != nil {
		panic(err)
	}
	return &Database{
		conn: db,
	}
}

func (db Database) CreateUser(data User) error {
	tx := db.conn.Create(&data)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (db Database) GetUser(email string) (*User, error) {
	var u *User
	tx := db.conn.First(&u, "email = ?", email)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return u, nil
}

func newDBConn(dbName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(User{})
	if err != nil {
		panic(err)
	}

	return db, nil
}
