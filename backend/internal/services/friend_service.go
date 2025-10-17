package services

import (
	"database/sql"
	"errors"
	"ai-doc-system/internal/models"
)

type FriendService struct {
	db *sql.DB
}

func NewFriendService(db *sql.DB) *FriendService {
	return &FriendService{db: db}
}

// Send friend request
func (s *FriendService) SendFriendRequest(fromUserID, toUserID int) error {
	if fromUserID == toUserID {
		return errors.New("cannot send friend request to yourself")
	}
	
	// Check if already friends
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM friendships 
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)`,
		fromUserID, toUserID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("friendship already exists or request pending")
	}
	
	// Create friend request
	_, err = s.db.Exec(`
		INSERT INTO friendships (user_id, friend_id, status) 
		VALUES ($1, $2, 'pending')`,
		fromUserID, toUserID)
	
	return err
}

// Accept friend request
func (s *FriendService) AcceptFriendRequest(userID, friendID int) error {
	// Update request status to accepted
	result, err := s.db.Exec(`
		UPDATE friendships SET status = 'accepted', updated_at = CURRENT_TIMESTAMP 
		WHERE user_id = $1 AND friend_id = $2 AND status = 'pending'`,
		friendID, userID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("friend request not found")
	}
	
	// Create reverse relationship
	_, err = s.db.Exec(`
		INSERT INTO friendships (user_id, friend_id, status) 
		VALUES ($1, $2, 'accepted')`,
		userID, friendID)
	
	return err
}

// Reject friend request
func (s *FriendService) RejectFriendRequest(userID, friendID int) error {
	result, err := s.db.Exec(`
		DELETE FROM friendships 
		WHERE user_id = $1 AND friend_id = $2 AND status = 'pending'`,
		friendID, userID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("friend request not found")
	}
	
	return nil
}

// Remove friend
func (s *FriendService) RemoveFriend(userID, friendID int) error {
	// Delete bidirectional relationship
	_, err := s.db.Exec(`
		DELETE FROM friendships 
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)`,
		userID, friendID)
	
	return err
}

// Get friends list
func (s *FriendService) GetFriends(userID int) ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.username, u.role, u.avatar, u.profile, u.created_at, u.updated_at
		FROM users u
		JOIN friendships f ON u.id = f.friend_id
		WHERE f.user_id = $1 AND f.status = 'accepted'
		ORDER BY u.username`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var friends []models.User
	for rows.Next() {
		var friend models.User
		err := rows.Scan(&friend.ID, &friend.Username, &friend.Role,
			&friend.Avatar, &friend.Profile, &friend.CreatedAt, &friend.UpdatedAt)
		if err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}
	
	return friends, nil
}

// Get pending friend requests
func (s *FriendService) GetPendingRequests(userID int) ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.username, u.role, u.avatar, u.profile, u.created_at, u.updated_at
		FROM users u
		JOIN friendships f ON u.id = f.user_id
		WHERE f.friend_id = $1 AND f.status = 'pending'
		ORDER BY f.created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var requests []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role,
			&user.Avatar, &user.Profile, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		requests = append(requests, user)
	}
	
	return requests, nil
}

// Search users (for adding friends)
func (s *FriendService) SearchUsers(currentUserID int, keyword string) ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.username, u.role, u.avatar, u.profile, u.created_at, u.updated_at
		FROM users u
		WHERE u.id != $1 AND u.username ILIKE $2
		AND NOT EXISTS (
			SELECT 1 FROM friendships f 
			WHERE (f.user_id = $1 AND f.friend_id = u.id) 
			   OR (f.user_id = u.id AND f.friend_id = $1)
		)
		ORDER BY u.username
		LIMIT 20`,
		currentUserID, "%"+keyword+"%")
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
	
	return users, nil
}

// Create friend group
func (s *FriendService) CreateFriendGroup(userID int, name string) (*models.FriendGroup, error) {
	var group models.FriendGroup
	err := s.db.QueryRow(`
		INSERT INTO friend_groups (user_id, group_name) 
		VALUES ($1, $2) 
		RETURNING id, user_id, group_name, created_at`,
		userID, name).Scan(
		&group.ID, &group.UserID, &group.GroupName, &group.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &group, nil
}

// Get user's friend groups
func (s *FriendService) GetFriendGroups(userID int) ([]models.FriendGroup, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, group_name, created_at 
		FROM friend_groups 
		WHERE user_id = $1 
		ORDER BY group_name`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var groups []models.FriendGroup
	for rows.Next() {
		var group models.FriendGroup
		err := rows.Scan(&group.ID, &group.UserID, &group.GroupName,
			&group.CreatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	
	return groups, nil
}

// Add friend to group
func (s *FriendService) AddFriendToGroup(userID, friendID, groupID int) error {
	// Check if friendship exists
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM friendships 
		WHERE user_id = $1 AND friend_id = $2 AND status = 'accepted'`,
		userID, friendID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("friendship not found")
	}
	
	// Update friend's group
	_, err = s.db.Exec(`
		UPDATE friendships SET group_id = $1 
		WHERE user_id = $2 AND friend_id = $3`,
		groupID, userID, friendID)
	
	return err
}