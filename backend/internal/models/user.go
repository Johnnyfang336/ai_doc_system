package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NullString custom type for proper JSON serialization handling
type NullString struct {
	sql.NullString
}

// MarshalJSON implements JSON serialization
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON implements JSON deserialization
func (ns *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ns.String = s
	ns.Valid = true
	return nil
}

// Value implements driver.Valuer interface
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

type User struct {
	ID        int        `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	Password  string     `json:"-" db:"password_hash"`
	Role      string     `json:"role" db:"role"`
	Avatar    NullString `json:"avatar" db:"avatar"`
	Profile   NullString `json:"profile" db:"profile"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type Friendship struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	FriendID  int       `json:"friend_id" db:"friend_id"`
	Status    string    `json:"status" db:"status"` // pending, accepted, blocked
	GroupID   *int      `json:"group_id" db:"group_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type FriendGroup struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	GroupName string    `json:"group_name" db:"group_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Message struct {
	ID         int       `json:"id" db:"id"`
	SenderID   int       `json:"sender_id" db:"sender_id"`
	ReceiverID int       `json:"receiver_id" db:"receiver_id"`
	Content    string    `json:"content" db:"content"`
	MessageType string   `json:"message_type" db:"message_type"` // text, file, system
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}