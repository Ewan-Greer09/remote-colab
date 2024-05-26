package db

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string `gorm:"default:NULL"`

	DisplayName string `gorm:"default:NULL"`

	Rooms []ChatRoom `gorm:"many2many:user_rooms;"`
	Teams []Team     `gorm:"many2many:user_teams;"`
}

type ChatRoom struct {
	gorm.Model
	UID     uuid.UUID
	Name    string `gorm:"unique"`
	Members []User `gorm:"many2many:user_rooms;"`
}

type Team struct {
	gorm.Model
	Name    string `gorm:"unique"`
	Members []User `gorm:"many2many:user_teams;"`
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
	tx := db.conn.Take(&u, "email = ?", email)
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

	err = db.AutoMigrate(User{}, ChatRoom{}, Team{})
	if err != nil {
		panic(err)
	}

	return db, nil
}

func (db Database) GetChatRoomsByUser(email string) ([]ChatRoom, error) {
	var user User
	tx := db.conn.Find(&user, "email = ?", email)
	if tx.Error != nil {
		return nil, tx.Error
	}

	var rooms []ChatRoom
	err := db.conn.Model(&user).Association("Rooms").Find(&rooms)
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (db Database) CreateRoom(email string, roomName string) error {
	var user User
	tx := db.conn.Find(&user, "email = ?", email)
	if tx.Error != nil {
		return tx.Error
	}

	room := ChatRoom{
		UID:     uuid.New(),
		Name:    roomName,
		Members: []User{user},
	}

	log.Printf("Created: %s", room.Name)

	tx = db.conn.Create(&room)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
