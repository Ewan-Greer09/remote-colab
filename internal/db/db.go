package db

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email       string `gorm:"unique"`
	Password    string `gorm:"default:NULL"`
	DisplayName string `gorm:"default:NULL"`

	ChatRooms []ChatRoom `gorm:"many2many:chat_room_members;"`
}

type ChatRoom struct {
	gorm.Model
	UID      string `gorm:"unique"`
	Name     string
	Members  []User `gorm:"many2many:chat_room_members;"`
	Messages []Message
}

type Message struct {
	gorm.Model
	ChatRoomID string `gorm:"index;not null;type:uuid"`

	Author   string
	Content  string   `gorm:"type:text"`
	ChatRoom ChatRoom `gorm:"foreignKey:ChatRoomID"`
}

type Team struct {
	gorm.Model
	UID     uuid.UUID `gorm:"unique"`
	Name    string
	Members []User `gorm:"many2many:team_members;"`
}

type Database struct {
	conn *gorm.DB
}

func NewDatabase() *Database {
	db, err := newDBConn("project.db")
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

	tx := db.conn.Model(&User{}).First(&u, "email = ?", email)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return u, nil
}

func (db Database) AddUserToRoom(u User, roomId string) error {
	var room ChatRoom
	tx := db.conn.First(&room, "uid = ?", roomId)
	if tx.Error != nil {
		return tx.Error
	}

	err := db.conn.Model(&room).Association("Members").Append(&u)
	if err != nil {
		return err
	}

	return nil
}

func newDBConn(dbName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(User{}, ChatRoom{}, Team{}, Message{})
	if err != nil {
		panic(err)
	}

	return db, nil
}

func (db Database) GetChatRoomsByUser(email string) ([]ChatRoom, error) {
	//TODO: this should probably all be concatinated into a single query, to save on db stress.
	var user User
	tx := db.conn.Find(&user, "email = ?", email)
	if tx.Error != nil {
		return nil, tx.Error
	}

	var rooms []ChatRoom
	err := db.conn.Model(&user).Association("ChatRooms").Find(&rooms)
	if err != nil {
		return nil, err
	}

	var members []User
	for i := range rooms {
		err = db.conn.Model(&rooms[i]).Association("Members").Find(&members)
		if err != nil {
			return nil, err
		}
		rooms[i].Members = members
	}

	return rooms, nil
}

func (db Database) CreateRoom(room ChatRoom, email string) error {
	var user User
	tx := db.conn.Find(&user, "email = ?", email)
	if tx.Error != nil {
		return tx.Error
	}

	room.Members = append(room.Members, user)

	tx = db.conn.Create(&room)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (db Database) GetAllRooms() ([]ChatRoom, error) {
	var rooms []ChatRoom
	tx := db.conn.Find(&rooms)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return rooms, nil
}

func (db Database) CreateMessage(roomId string, author string, content string) error {
	var room ChatRoom
	tx := db.conn.First(&room, "uid = ?", roomId)
	if tx.Error != nil {
		return tx.Error
	}

	msg := Message{
		ChatRoom:   room,
		ChatRoomID: roomId,
		Author:     author,
		Content:    content,
	}

	tx = db.conn.Create(&msg)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (db Database) GetMessagesByRoomUID(roomUID string) ([]Message, error) {
	var room ChatRoom
	tx := db.conn.First(&room, "uid = ?", roomUID)
	if tx.Error != nil {
		return nil, tx.Error
	}
	var messages []Message
	err := db.conn.Model(&room).Association("Messages").Find(&messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
