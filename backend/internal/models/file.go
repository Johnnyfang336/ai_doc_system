package models

import (
	"time"
)

type File struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Filename  string    `json:"filename" db:"filename"`
	FilePath  string    `json:"file_path" db:"file_path"`
	FileSize  int64     `json:"file_size" db:"file_size"`
	MimeType  string    `json:"mime_type" db:"mime_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type FileVersion struct {
	ID           int       `json:"id" db:"id"`
	FileID       int       `json:"file_id" db:"file_id"`
	VersionNumber int      `json:"version_number" db:"version_number"`
	FilePath     string    `json:"file_path" db:"file_path"`
	CreatedBy    int       `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type FileShare struct {
	ID        int       `json:"id" db:"id"`
	FileID    int       `json:"file_id" db:"file_id"`
	ShareType string    `json:"share_type" db:"share_type"` // friend, public
	ShareToken string   `json:"share_token" db:"share_token"`
	SharedWith *int     `json:"shared_with" db:"shared_with"` // User ID to share with, null for public sharing
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CollaborationSession struct {
	ID           int       `json:"id" db:"id"`
	FileID       int       `json:"file_id" db:"file_id"`
	Participants []int     `json:"participants" db:"participants"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}