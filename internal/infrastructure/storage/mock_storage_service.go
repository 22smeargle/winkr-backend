package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// MockStorageService is a mock implementation of StorageService for testing
type MockStorageService struct {
	files map[string]*MockFile
	
	// Method call tracking
	CalledGetUploadURL      bool
	CalledGetDownloadURL     bool
	CalledUploadFile         bool
	CalledDeleteFile         bool
	CalledGetFile           bool
	CalledGetFileMetadata    bool
	CalledListFiles          bool
	CalledCopyFile           bool
	CalledMoveFile           bool
	
	// Mock responses
	UploadURL    string
	DownloadURL string
	UploadError  error
	DownloadError error
	DeleteError  error
	GetError     error
	ListError    error
	CopyError    error
	MoveError    error
}

// MockFile represents a mock file for testing
type MockFile struct {
	Key         string
	ContentType string
	Size        int64
	LastModified time.Time
	Data        []byte
	Metadata    map[string]string
}

// NewMockStorageService creates a new mock storage service
func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		files: make(map[string]*MockFile),
		UploadURL:    "https://mock-upload-url.com/upload",
		DownloadURL: "https://mock-download-url.com/download",
	}
}

// GetUploadURL generates a presigned URL for file upload
func (m *MockStorageService) GetUploadURL(ctx context.Context, key string, contentType string, expiresIn time.Duration) (string, error) {
	m.CalledGetUploadURL = true
	if m.UploadError != nil {
		return "", m.UploadError
	}
	return m.UploadURL + "?key=" + key + "&type=" + contentType, nil
}

// GetDownloadURL generates a presigned URL for file download
func (m *MockStorageService) GetDownloadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	m.CalledGetDownloadURL = true
	if m.DownloadError != nil {
		return "", m.DownloadError
	}
	return m.DownloadURL + "?key=" + key, nil
}

// UploadFile uploads a file to storage
func (m *MockStorageService) UploadFile(ctx context.Context, key string, file io.Reader, contentType string, size int64) error {
	m.CalledUploadFile = true
	if m.UploadError != nil {
		return m.UploadError
	}
	
	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	
	// Store mock file
	m.files[key] = &MockFile{
		Key:         key,
		ContentType: contentType,
		Size:        size,
		LastModified: time.Now(),
		Data:        data,
		Metadata:    make(map[string]string),
	}
	
	return nil
}

// UploadMultipartFile uploads a multipart file to storage
func (m *MockStorageService) UploadMultipartFile(ctx context.Context, key string, file *multipart.FileHeader) error {
	m.CalledUploadFile = true
	if m.UploadError != nil {
		return m.UploadError
	}
	
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	
	// Read file data
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	
	// Store mock file
	m.files[key] = &MockFile{
		Key:         key,
		ContentType: file.Header.Get("Content-Type"),
		Size:        file.Size,
		LastModified: time.Now(),
		Data:        data,
		Metadata:    make(map[string]string),
	}
	
	return nil
}

// DeleteFile deletes a file from storage
func (m *MockStorageService) DeleteFile(ctx context.Context, key string) error {
	m.CalledDeleteFile = true
	if m.DeleteError != nil {
		return m.DeleteError
	}
	
	delete(m.files, key)
	return nil
}

// GetFile gets a file from storage
func (m *MockStorageService) GetFile(ctx context.Context, key string) (io.ReadCloser, int64, string, error) {
	m.CalledGetFile = true
	if m.GetError != nil {
		return nil, 0, "", m.GetError
	}
	
	if file, exists := m.files[key]; exists {
		// Create a reader from the file data
		reader := io.NopCloser(bytes.NewReader(file.Data))
		return reader, file.Size, file.ContentType, nil
	}
	
	return nil, 0, "", fmt.Errorf("file not found: %s", key)
}

// GetFileMetadata gets file metadata from storage
func (m *MockStorageService) GetFileMetadata(ctx context.Context, key string) (*FileMetadata, error) {
	m.CalledGetFileMetadata = true
	if m.GetError != nil {
		return nil, m.GetError
	}
	
	if file, exists := m.files[key]; exists {
		return &FileMetadata{
			Key:          file.Key,
			Size:         file.Size,
			ContentType:  file.ContentType,
			LastModified: file.LastModified,
			Metadata:     file.Metadata,
		}, nil
	}
	
	return nil, fmt.Errorf("file not found: %s", key)
}

// ListFiles lists files in storage with optional prefix
func (m *MockStorageService) ListFiles(ctx context.Context, prefix string, limit int) ([]*FileMetadata, error) {
	m.CalledListFiles = true
	if m.ListError != nil {
		return nil, m.ListError
	}
	
	var files []*FileMetadata
	count := 0
	for _, file := range m.files {
		if len(prefix) == 0 || (len(file.Key) >= len(prefix) && file.Key[:len(prefix)] == prefix) {
			files = append(files, &FileMetadata{
				Key:          file.Key,
				Size:         file.Size,
				ContentType:  file.ContentType,
				LastModified: file.LastModified,
				Metadata:     file.Metadata,
			})
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	
	return files, nil
}

// CopyFile copies a file within storage
func (m *MockStorageService) CopyFile(ctx context.Context, sourceKey, destKey string) error {
	m.CalledCopyFile = true
	if m.CopyError != nil {
		return m.CopyError
	}
	
	if file, exists := m.files[sourceKey]; exists {
		// Create a copy
		copy := &MockFile{
			Key:         destKey,
			ContentType: file.ContentType,
			Size:        file.Size,
			LastModified: time.Now(),
			Data:        make([]byte, len(file.Data)),
			Metadata:    make(map[string]string),
		}
		copy(copy.Data, file.Data)
		
		// Copy metadata
		for k, v := range file.Metadata {
			copy.Metadata[k] = v
		}
		
		m.files[destKey] = copy
		return nil
	}
	
	return fmt.Errorf("source file not found: %s", sourceKey)
}

// MoveFile moves a file within storage
func (m *MockStorageService) MoveFile(ctx context.Context, sourceKey, destKey string) error {
	m.CalledMoveFile = true
	if m.MoveError != nil {
		return m.MoveError
	}
	
	// First copy the file
	err := m.CopyFile(ctx, sourceKey, destKey)
	if err != nil {
		return err
	}
	
	// Then delete the source
	return m.DeleteFile(ctx, sourceKey)
}

// Reset resets the mock storage service state
func (m *MockStorageService) Reset() {
	m.files = make(map[string]*MockFile)
	
	m.CalledGetUploadURL = false
	m.CalledGetDownloadURL = false
	m.CalledUploadFile = false
	m.CalledDeleteFile = false
	m.CalledGetFile = false
	m.CalledGetFileMetadata = false
	m.CalledListFiles = false
	m.CalledCopyFile = false
	m.CalledMoveFile = false
	
	m.UploadError = nil
	m.DownloadError = nil
	m.DeleteError = nil
	m.GetError = nil
	m.ListError = nil
	m.CopyError = nil
	m.MoveError = nil
}

// SetUploadURL sets the mock upload URL
func (m *MockStorageService) SetUploadURL(url string) {
	m.UploadURL = url
}

// SetDownloadURL sets the mock download URL
func (m *MockStorageService) SetDownloadURL(url string) {
	m.DownloadURL = url
}

// SetUploadError sets the mock upload error
func (m *MockStorageService) SetUploadError(err error) {
	m.UploadError = err
}

// SetDownloadError sets the mock download error
func (m *MockStorageService) SetDownloadError(err error) {
	m.DownloadError = err
}

// SetDeleteError sets the mock delete error
func (m *MockStorageService) SetDeleteError(err error) {
	m.DeleteError = err
}

// SetGetError sets the mock get error
func (m *MockStorageService) SetGetError(err error) {
	m.GetError = err
}

// AddFile adds a file to the mock storage
func (m *MockStorageService) AddFile(key string, data []byte, contentType string) {
	m.files[key] = &MockFile{
		Key:         key,
		ContentType: contentType,
		Size:        int64(len(data)),
		LastModified: time.Now(),
		Data:        data,
		Metadata:    make(map[string]string),
	}
}

// GetFileCount returns the number of files in mock storage
func (m *MockStorageService) GetFileCount() int {
	return len(m.files)
}

// FileExists checks if a file exists in mock storage
func (m *MockStorageService) FileExists(key string) bool {
	_, exists := m.files[key]
	return exists
}

// VerifyMethodCalls verifies that specific methods were called
func (m *MockStorageService) VerifyMethodCalls(expectedGetUploadURL, expectedGetDownloadURL, expectedUploadFile, expectedDeleteFile, expectedGetFile, expectedGetFileMetadata, expectedListFiles, expectedCopyFile, expectedMoveFile bool) error {
	if expectedGetUploadURL && !m.CalledGetUploadURL {
		return fmt.Errorf("expected GetUploadURL to be called")
	}
	if !expectedGetUploadURL && m.CalledGetUploadURL {
		return fmt.Errorf("expected GetUploadURL NOT to be called")
	}
	
	if expectedGetDownloadURL && !m.CalledGetDownloadURL {
		return fmt.Errorf("expected GetDownloadURL to be called")
	}
	if !expectedGetDownloadURL && m.CalledGetDownloadURL {
		return fmt.Errorf("expected GetDownloadURL NOT to be called")
	}
	
	if expectedUploadFile && !m.CalledUploadFile {
		return fmt.Errorf("expected UploadFile to be called")
	}
	if !expectedUploadFile && m.CalledUploadFile {
		return fmt.Errorf("expected UploadFile NOT to be called")
	}
	
	if expectedDeleteFile && !m.CalledDeleteFile {
		return fmt.Errorf("expected DeleteFile to be called")
	}
	if !expectedDeleteFile && m.CalledDeleteFile {
		return fmt.Errorf("expected DeleteFile NOT to be called")
	}
	
	if expectedGetFile && !m.CalledGetFile {
		return fmt.Errorf("expected GetFile to be called")
	}
	if !expectedGetFile && m.CalledGetFile {
		return fmt.Errorf("expected GetFile NOT to be called")
	}
	
	if expectedGetFileMetadata && !m.CalledGetFileMetadata {
		return fmt.Errorf("expected GetFileMetadata to be called")
	}
	if !expectedGetFileMetadata && m.CalledGetFileMetadata {
		return fmt.Errorf("expected GetFileMetadata NOT to be called")
	}
	
	if expectedListFiles && !m.CalledListFiles {
		return fmt.Errorf("expected ListFiles to be called")
	}
	if !expectedListFiles && m.CalledListFiles {
		return fmt.Errorf("expected ListFiles NOT to be called")
	}
	
	if expectedCopyFile && !m.CalledCopyFile {
		return fmt.Errorf("expected CopyFile to be called")
	}
	if !expectedCopyFile && m.CalledCopyFile {
		return fmt.Errorf("expected CopyFile NOT to be called")
	}
	
	if expectedMoveFile && !m.CalledMoveFile {
		return fmt.Errorf("expected MoveFile to be called")
	}
	if !expectedMoveFile && m.CalledMoveFile {
		return fmt.Errorf("expected MoveFile NOT to be called")
	}
	
	return nil
}