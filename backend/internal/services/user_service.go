package services

import (
	"database/sql"
	"errors"
	"ai-doc-system/internal/models"
	"ai-doc-system/internal/utils"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Register(username, password string) (*models.User, error) {
	// Check if username already exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("username already exists")
	}
	
	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	
	// Create user
	var user models.User
	err = s.db.QueryRow(`
		INSERT INTO users (username, password_hash, role) 
		VALUES ($1, $2, 'user') 
		RETURNING id, username, role, created_at, updated_at`,
		username, hashedPassword).Scan(
		&user.ID, &user.Username, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (s *UserService) Login(username, password string) (*models.User, error) {
	var user models.User
	var hashedPassword string
	
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, role, avatar, profile, created_at, updated_at 
		FROM users WHERE username = $1`, username).Scan(
		&user.ID, &user.Username, &hashedPassword, &user.Role, 
		&user.Avatar, &user.Profile, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	
	if !utils.CheckPassword(password, hashedPassword) {
		return nil, errors.New("invalid password")
	}
	
	return &user, nil
}

func (s *UserService) GetUserByID(userID int) (*models.User, error) {
	var user models.User
	
	err := s.db.QueryRow(`
		SELECT id, username, role, avatar, profile, created_at, updated_at 
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.Role, &user.Avatar, 
		&user.Profile, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	
	return &user, nil
}

func (s *UserService) UpdateProfile(userID int, avatar, profile string) error {
	_, err := s.db.Exec(`
		UPDATE users SET avatar = $1, profile = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3`, avatar, profile, userID)
	return err
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT id, username, role, avatar, profile, created_at, updated_at 
		FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role, 
			&user.Avatar, &user.Profile, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return users, nil
}