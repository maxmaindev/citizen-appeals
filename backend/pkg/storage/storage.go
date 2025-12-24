package storage

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"citizen-appeals/config"
)

var (
	ErrFileTooLarge     = errors.New("file size exceeds maximum allowed")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrTooManyFiles     = errors.New("too many files")
	ErrFileNotFound     = errors.New("file not found")
	ErrStorageNotReady  = errors.New("storage not ready")
)

// AllowedMimeTypes for photos
var AllowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/png":       true,
	"image/webp":      true,
	"image/gif":       true,
}

// Storage interface for file storage operations
type Storage interface {
	Save(file multipart.File, header *multipart.FileHeader, appealID int64, isResultPhoto bool) (string, string, int64, error)
	Get(filePath string) (io.ReadCloser, error)
	Delete(filePath string) error
	GetURL(filePath string) string
}

// LocalStorage implements Storage interface for local file system
type LocalStorage struct {
	basePath string
	baseURL  string
	maxSize  int64
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(cfg *config.Config) (*LocalStorage, error) {
	uploadPath := cfg.Upload.UploadPath
	if uploadPath == "" {
		uploadPath = "./uploads"
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	baseURL := "/uploads"
	if cfg.Env == "production" {
		// In production, you might want to use a CDN URL
		baseURL = "/api/uploads"
	}

	return &LocalStorage{
		basePath: uploadPath,
		baseURL:  baseURL,
		maxSize:  cfg.Upload.MaxSize,
	}, nil
}

// Save saves a file to local storage
func (s *LocalStorage) Save(file multipart.File, header *multipart.FileHeader, appealID int64, isResultPhoto bool) (string, string, int64, error) {
	// Validate file size
	if header.Size > s.maxSize {
		return "", "", 0, ErrFileTooLarge
	}

	// Determine MIME type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		// Try to detect from file extension
		ext := filepath.Ext(header.Filename)
		detectedType := mime.TypeByExtension(ext)
		if detectedType != "" {
			mimeType = detectedType
		} else {
			// Try to detect from file content (read first 512 bytes)
			buffer := make([]byte, 512)
			n, _ := file.Read(buffer)
			file.Seek(0, 0) // Reset file pointer
			mimeType = http.DetectContentType(buffer[:n])
		}
	}

	// Validate MIME type
	if !AllowedMimeTypes[mimeType] {
		return "", "", 0, ErrInvalidFileType
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%d%s", appealID, timestamp, ext)
	
	// Create subdirectory for appeal
	subDir := fmt.Sprintf("appeal_%d", appealID)
	if isResultPhoto {
		subDir = filepath.Join(subDir, "results")
	}
	
	fullDir := filepath.Join(s.basePath, subDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", "", 0, fmt.Errorf("failed to create directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(fullDir, filename)
	relativePath := filepath.Join(subDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return "", "", 0, fmt.Errorf("failed to save file: %w", err)
	}

	return relativePath, mimeType, written, nil
}

// Get retrieves a file from local storage
func (s *LocalStorage) Get(filePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, filePath)
	
	// Security: prevent directory traversal
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.basePath)) {
		return nil, ErrFileNotFound
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete removes a file from local storage
func (s *LocalStorage) Delete(filePath string) error {
	fullPath := filepath.Join(s.basePath, filePath)
	
	// Security: prevent directory traversal
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.basePath)) {
		return ErrFileNotFound
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL returns the URL to access the file
func (s *LocalStorage) GetURL(filePath string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, filepath.ToSlash(filePath))
}

// ValidateFile validates file before upload
func ValidateFile(header *multipart.FileHeader, maxSize int64, maxCount, currentCount int) error {
	// Check file size
	if header.Size > maxSize {
		return fmt.Errorf("%w: file size %d exceeds maximum %d", ErrFileTooLarge, header.Size, maxSize)
	}

	// Determine MIME type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		// Try to detect from file extension
		ext := filepath.Ext(header.Filename)
		detectedType := mime.TypeByExtension(ext)
		if detectedType != "" {
			mimeType = detectedType
		}
		// Note: We can't use http.DetectContentType here because we don't have file content yet
		// The actual validation will happen in Save() method
	}

	// Check MIME type (allow if empty/octet-stream, will be validated in Save)
	if mimeType != "" && mimeType != "application/octet-stream" && !AllowedMimeTypes[mimeType] {
		return fmt.Errorf("%w: %s is not allowed", ErrInvalidFileType, mimeType)
	}

	// Check file count
	if currentCount >= maxCount {
		return fmt.Errorf("%w: maximum %d files allowed", ErrTooManyFiles, maxCount)
	}

	return nil
}

