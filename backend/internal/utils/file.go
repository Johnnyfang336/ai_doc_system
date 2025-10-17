package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/google/uuid"
)

func GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Unix()
	uid := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%d_%s%s", name, timestamp, uid, ext)
}

func SaveUploadedFile(file *multipart.FileHeader, uploadPath string) (string, error) {
	// Create upload directory
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", err
	}
	
	// Generate unique filename
	fileName := GenerateFileName(file.Filename)
	filePath := filepath.Join(uploadPath, fileName)
	
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	
	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	
	// Copy file content
	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}
	
	return fileName, nil
}

func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}