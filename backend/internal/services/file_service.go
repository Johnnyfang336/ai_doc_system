package services

import (
	"database/sql"
	"errors"
	"mime/multipart"
	"path/filepath"
	
	"ai-doc-system/internal/models"
	"ai-doc-system/internal/utils"
)

type FileService struct {
	db         *sql.DB
	uploadPath string
}

func NewFileService(db *sql.DB, uploadPath string) *FileService {
	return &FileService{
		db:         db,
		uploadPath: uploadPath,
	}
}

func (s *FileService) UploadFile(userID int, file *multipart.FileHeader) (*models.File, error) {
	// Check file size (10MB limit)
	if file.Size > 10*1024*1024 {
		return nil, errors.New("file size exceeds 10MB limit")
	}
	
	// Check user storage space (100MB limit)
	var totalSize int64
	err := s.db.QueryRow("SELECT COALESCE(SUM(file_size), 0) FROM files WHERE user_id = $1", userID).Scan(&totalSize)
	if err != nil {
		return nil, err
	}
	
	if totalSize+file.Size > 100*1024*1024 {
		return nil, errors.New("storage limit exceeded (100MB)")
	}
	
	// Save file to disk
	fileName, err := utils.SaveUploadedFile(file, s.uploadPath)
	if err != nil {
		return nil, err
	}
	
	filePath := filepath.Join(s.uploadPath, fileName)
	
	// Save file information to database
	var fileModel models.File
	err = s.db.QueryRow(`
		INSERT INTO files (filename, original_name, file_path, file_size, mime_type, user_id) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, filename, original_name, file_path, file_size, mime_type, user_id, created_at, updated_at`,
		fileName, file.Filename, filePath, file.Size, file.Header.Get("Content-Type"), userID).Scan(
		&fileModel.ID, &fileModel.Filename, &fileModel.OriginalName, &fileModel.FilePath,
		&fileModel.FileSize, &fileModel.MimeType, &fileModel.UserID,
		&fileModel.CreatedAt, &fileModel.UpdatedAt)
	
	if err != nil {
		// If database operation fails, delete uploaded file
		utils.DeleteFile(filePath)
		return nil, err
	}
	
	// Create initial version
	_, err = s.db.Exec(`
		INSERT INTO file_versions (file_id, version_number, file_path, created_by) 
		VALUES ($1, 1, $2, $3)`,
		fileModel.ID, filePath, userID)
	
	if err != nil {
		return nil, err
	}
	
	return &fileModel, nil
}

func (s *FileService) GetUserFiles(userID int) ([]models.File, error) {
	rows, err := s.db.Query(`
		SELECT id, filename, original_name, file_path, file_size, mime_type, user_id, created_at, updated_at 
		FROM files WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var files []models.File
	for rows.Next() {
		var file models.File
		err := rows.Scan(&file.ID, &file.Filename, &file.OriginalName, &file.FilePath,
			&file.FileSize, &file.MimeType, &file.UserID,
			&file.CreatedAt, &file.UpdatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	
	return files, nil
}

func (s *FileService) GetFileByID(fileID int) (*models.File, error) {
	var file models.File
	err := s.db.QueryRow(`
		SELECT id, filename, original_name, file_path, file_size, mime_type, user_id, created_at, updated_at 
		FROM files WHERE id = $1`, fileID).Scan(
		&file.ID, &file.Filename, &file.OriginalName, &file.FilePath,
		&file.FileSize, &file.MimeType, &file.UserID,
		&file.CreatedAt, &file.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &file, nil
}

func (s *FileService) DeleteFile(fileID, userID int) error {
	// Get file information
	file, err := s.GetFileByID(fileID)
	if err != nil {
		return err
	}
	
	// Check permissions
	if file.UserID != userID {
		return errors.New("permission denied")
	}
	
	// Delete database record
	_, err = s.db.Exec("DELETE FROM files WHERE id = $1", fileID)
	if err != nil {
		return err
	}
	
	// Delete physical file
	if utils.FileExists(file.FilePath) {
		utils.DeleteFile(file.FilePath)
	}
	
	return nil
}

func (s *FileService) RenameFile(fileID, userID int, newName string) error {
	// Check permissions
	var ownerID int
	err := s.db.QueryRow("SELECT user_id FROM files WHERE id = $1", fileID).Scan(&ownerID)
	if err != nil {
		return err
	}
	
	if ownerID != userID {
		return errors.New("permission denied")
	}
	
	// Update original_name (the displayed name)
	_, err = s.db.Exec(`
		UPDATE files SET original_name = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`, newName, fileID)
	
	return err
}

func (s *FileService) GetAllFiles() ([]models.File, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.filename, f.original_name, f.file_path, f.file_size, f.mime_type, 
		       f.user_id, f.created_at, f.updated_at, u.username
		FROM files f 
		JOIN users u ON f.user_id = u.id 
		ORDER BY f.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var files []models.File
	for rows.Next() {
		var file models.File
		var username string
		err := rows.Scan(&file.ID, &file.Filename, &file.OriginalName, &file.FilePath,
			&file.FileSize, &file.MimeType, &file.UserID,
			&file.CreatedAt, &file.UpdatedAt, &username)
		if err != nil {
			return nil, err
		}
		// Can add username information to file structure here
		files = append(files, file)
	}
	
	return files, nil
}

func (s *FileService) GetUserStorageUsage(userID int) (int64, error) {
	var totalSize int64
	err := s.db.QueryRow("SELECT COALESCE(SUM(file_size), 0) FROM files WHERE user_id = $1", userID).Scan(&totalSize)
	return totalSize, err
}