package services

import (
	"database/sql"
	"errors"
	"ai-doc-system/internal/models"
)

type MessageService struct {
	db *sql.DB
}

func NewMessageService(db *sql.DB) *MessageService {
	return &MessageService{db: db}
}

// Send message
func (s *MessageService) SendMessage(fromUserID, toUserID int, content string) (*models.Message, error) {
	// Check if they are friends
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM friendships 
		WHERE user_id = $1 AND friend_id = $2 AND status = 'accepted'`,
		fromUserID, toUserID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("can only send messages to friends")
	}
	
	// Create message
	var message models.Message
	err = s.db.QueryRow(`
		INSERT INTO messages (sender_id, receiver_id, content) 
		VALUES ($1, $2, $3) 
		RETURNING id, sender_id, receiver_id, content, message_type, created_at`,
		fromUserID, toUserID, content).Scan(
		&message.ID, &message.SenderID, &message.ReceiverID, 
		&message.Content, &message.MessageType, &message.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &message, nil
}

// Get chat history with specific user
func (s *MessageService) GetChatHistory(userID, friendID int, limit, offset int) ([]models.Message, error) {
	rows, err := s.db.Query(`
		SELECT id, sender_id, receiver_id, content, message_type, created_at
		FROM messages 
		WHERE (sender_id = $1 AND receiver_id = $2) 
		   OR (sender_id = $2 AND receiver_id = $1)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		userID, friendID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(&message.ID, &message.SenderID, &message.ReceiverID,
			&message.Content, &message.MessageType, &message.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	
	return messages, nil
}

// Get user's all chat list
func (s *MessageService) GetChatList(userID int) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`
		WITH latest_messages AS (
			SELECT 
				CASE 
					WHEN sender_id = $1 THEN receiver_id 
					ELSE sender_id 
				END as friend_id,
				content,
				created_at,
				ROW_NUMBER() OVER (
					PARTITION BY CASE 
						WHEN sender_id = $1 THEN receiver_id 
						ELSE sender_id 
					END 
					ORDER BY created_at DESC
				) as rn
			FROM messages 
			WHERE sender_id = $1 OR receiver_id = $1
		)
		SELECT 
			u.id, u.username, u.avatar,
			lm.content, lm.created_at,
			COALESCE(unread.count, 0) as unread_count
		FROM latest_messages lm
		JOIN users u ON u.id = lm.friend_id
		LEFT JOIN (
			SELECT sender_id, COUNT(*) as count
			FROM messages 
			WHERE receiver_id = $1
			GROUP BY sender_id
		) unread ON unread.sender_id = u.id
		WHERE lm.rn = 1
		ORDER BY lm.created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var chatList []map[string]interface{}
	for rows.Next() {
		var friendID int
		var username, avatar, content string
		var createdAt string
		var unreadCount int
		
		err := rows.Scan(&friendID, &username, &avatar, &content, &createdAt, &unreadCount)
		if err != nil {
			return nil, err
		}
		
		chat := map[string]interface{}{
			"friend_id":     friendID,
			"username":      username,
			"avatar":        avatar,
			"last_message":  content,
			"last_time":     createdAt,
			"unread_count":  unreadCount,
		}
		chatList = append(chatList, chat)
	}
	
	return chatList, nil
}

// Mark messages as read
func (s *MessageService) MarkMessagesAsRead(userID, fromUserID int) error {
	_, err := s.db.Exec(`
		UPDATE messages 
		SET message_type = 'text' 
		WHERE receiver_id = $1 AND sender_id = $2`,
		userID, fromUserID)
	
	return err
}

// Get unread message count
func (s *MessageService) GetUnreadCount(userID int) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM messages 
		WHERE receiver_id = $1`,
		userID).Scan(&count)
	
	return count, err
}

// Delete message
func (s *MessageService) DeleteMessage(messageID, userID int) error {
	// Only allow sender to delete message
	result, err := s.db.Exec(`
		DELETE FROM messages 
		WHERE id = $1 AND sender_id = $2`,
		messageID, userID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("message not found or permission denied")
	}
	
	return nil
}