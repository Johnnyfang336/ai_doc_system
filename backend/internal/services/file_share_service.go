package services

import (
	"database/sql"
	"errors"
	"time"
	"ai-doc-system/internal/models"
	"github.com/google/uuid"
)

type FileShareService struct {
	db *sql.DB
}

func NewFileShareService(db *sql.DB) *FileShareService {
	return &FileShareService{db: db}
}

// Share file to friend
func (s *FileShareService) ShareFileToFriend(fileID, ownerID, friendID int) error {
	// Check file ownership
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM files WHERE id = $1 AND user_id = $2", fileID, ownerID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("file not found or permission denied")
	}
	
	// Check friendship
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM friendships 
		WHERE user_id = $1 AND friend_id = $2 AND status = 'accepted'`,
		ownerID, friendID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("can only share files with friends")
	}
	
	// Check if already shared
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM file_shares 
		WHERE file_id = $1 AND shared_with_user_id = $2 AND share_type = 'friend'`,
		fileID, friendID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("file already shared with this friend")
	}
	
	// Create share record
	_, err = s.db.Exec(`
		INSERT INTO file_shares (file_id, shared_by_user_id, shared_with_user_id, share_type, permissions) 
		VALUES ($1, $2, $3, 'friend', 'read')`,
		fileID, ownerID, friendID)
	
	return err
}

// Create public share link
func (s *FileShareService) CreatePublicShare(fileID, ownerID int, expiresAt *time.Time) (*models.FileShare, error) {
	// Check file ownership
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM files WHERE id = $1 AND user_id = $2", fileID, ownerID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("file not found or permission denied")
	}
	
	// Generate share link
	shareToken := uuid.New().String()
	
	// Create public share record
	var share models.FileShare
	err = s.db.QueryRow(`
		INSERT INTO file_shares (file_id, created_by, share_type, share_token, expires_at) 
		VALUES ($1, $2, 'public', $3, $4) 
		RETURNING id, file_id, created_by, share_type, share_token, expires_at, created_at`,
		fileID, ownerID, shareToken, expiresAt).Scan(
		&share.ID, &share.FileID, &share.CreatedBy,
		&share.ShareType, &share.ShareToken, &share.ExpiresAt, &share.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &share, nil
}

// Get files shared with me
func (s *FileShareService) GetSharedWithMeFiles(userID int) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.filename, f.file_size, f.mime_type, f.created_at,
		       u.username as shared_by, fs.created_at as shared_at
		FROM file_shares fs
		JOIN files f ON fs.file_id = f.id
		JOIN users u ON fs.shared_by_user_id = u.id
		WHERE fs.shared_with_user_id = $1
		ORDER BY fs.created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var files []map[string]interface{}
	for rows.Next() {
		var fileID int
		var filename, mimeType, sharedBy string
		var size int64
		var createdAt, sharedAt time.Time
		
		err := rows.Scan(&fileID, &filename, &size, &mimeType, &createdAt, &sharedBy, &sharedAt)
		if err != nil {
			return nil, err
		}
		
		file := map[string]interface{}{
			"id":            fileID,
			"filename":      filename,
			"size":          size,
			"mime_type":     mimeType,
			"created_at":    createdAt,
			"shared_by":     sharedBy,
			"shared_at":     sharedAt,
		}
		files = append(files, file)
	}
	
	return files, nil
}

// Get my shared files
func (s *FileShareService) GetMySharedFiles(userID int) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.filename, f.file_size, f.mime_type,
		       fs.share_type, fs.share_token, fs.expires_at, fs.created_at as shared_at,
		       u.username as shared_with
		FROM file_shares fs
		JOIN files f ON fs.file_id = f.id
		LEFT JOIN users u ON fs.shared_with_user_id = u.id
		WHERE fs.shared_by_user_id = $1
		ORDER BY fs.created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var files []map[string]interface{}
	for rows.Next() {
		var fileID int
		var filename, mimeType, shareType string
		var shareToken sql.NullString
		var sharedWith sql.NullString
		var size int64
		var expiresAt sql.NullTime
		var sharedAt time.Time
		
		err := rows.Scan(&fileID, &filename, &size, &mimeType, &shareType, 
			&shareToken, &expiresAt, &sharedAt, &sharedWith)
		if err != nil {
			return nil, err
		}
		
		file := map[string]interface{}{
			"id":            fileID,
			"filename":      filename,
			"size":          size,
			"mime_type":     mimeType,
			"share_type":    shareType,
			"shared_at":     sharedAt,
		}
		
		if shareToken.Valid {
			file["share_token"] = shareToken.String
		}
		if expiresAt.Valid {
			file["expires_at"] = expiresAt.Time
		}
		if sharedWith.Valid {
			file["shared_with"] = sharedWith.String
		}
		
		files = append(files, file)
	}
	
	return files, nil
}

// Get file by share token
func (s *FileShareService) GetFileByShareToken(shareToken string) (*models.File, error) {
	var file models.File
	err := s.db.QueryRow(`
		SELECT f.id, f.filename, f.file_path, f.file_size, f.mime_type, f.user_id, f.created_at, f.updated_at
		FROM files f
		JOIN file_shares fs ON f.id = fs.file_id
		WHERE fs.share_token = $1 AND fs.share_type = 'public'
		AND (fs.expires_at IS NULL OR fs.expires_at > CURRENT_TIMESTAMP)`,
		shareToken).Scan(
		&file.ID, &file.Filename, &file.FilePath,
		&file.FileSize, &file.MimeType, &file.UserID,
		&file.CreatedAt, &file.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &file, nil
}

// Remove share
func (s *FileShareService) RemoveShare(shareID, userID int) error {
	result, err := s.db.Exec(`
		DELETE FROM file_shares 
		WHERE id = $1 AND shared_by_user_id = $2`,
		shareID, userID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("share not found or permission denied")
	}
	
	return nil
}

// Check if user has file access permission
func (s *FileShareService) CheckFileAccess(fileID, userID int) (bool, error) {
	// Check if user is file owner
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM files WHERE id = $1 AND owner_id = $2", fileID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	
	// Check if user has share permission
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM file_shares 
		WHERE file_id = $1 AND shared_with_user_id = $2 AND share_type = 'friend'`,
		fileID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}