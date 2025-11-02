package photo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/internal/application/services"
)

// MockPhotoRepository is a mock implementation of the photo repository
type MockPhotoRepository struct {
	mock.Mock
}

func (m *MockPhotoRepository) Create(ctx context.Context, photo *entities.Photo) error {
	args := m.Called(ctx, photo)
	return args.Error(0)
}

func (m *MockPhotoRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Photo, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Photo), args.Error(1)
}

func (m *MockPhotoRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Photo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Photo), args.Error(1)
}

func (m *MockPhotoRepository) Update(ctx context.Context, photo *entities.Photo) error {
	args := m.Called(ctx, photo)
	return args.Error(0)
}

func (m *MockPhotoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPhotoRepository) GetPrimaryPhoto(ctx context.Context, userID uuid.UUID) (*entities.Photo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*entities.Photo), args.Error(1)
}

func (m *MockPhotoRepository) SetPrimaryPhoto(ctx context.Context, userID, photoID uuid.UUID) error {
	args := m.Called(ctx, userID, photoID)
	return args.Error(0)
}

func (m *MockPhotoRepository) CountUserPhotos(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

// MockStorageService is a mock implementation of the storage service
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) ValidateFile(contentType string, size int64) error {
	args := m.Called(contentType, size)
	return args.Error(0)
}

func (m *MockStorageService) GenerateUploadURL(key, contentType string) (string, error) {
	args := m.Called(key, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) GenerateDownloadURL(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) UploadFile(key string, data interface{}, contentType string, size int64) error {
	args := m.Called(key, data, contentType, size)
	return args.Error(0)
}

func (m *MockStorageService) DownloadFile(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockStorageService) DeleteFile(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockStorageService) FileExists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorageService) GetFileInfo(key string) (*storage.FileInfo, error) {
	args := m.Called(key)
	return args.Get(0).(*storage.FileInfo), args.Error(1)
}

func (m *MockStorageService) ListFiles(prefix string) ([]*storage.FileInfo, error) {
	args := m.Called(prefix)
	return args.Get(0).([]*storage.FileInfo), args.Error(1)
}

func (m *MockStorageService) CopyFile(sourceKey, destKey string) error {
	args := m.Called(sourceKey, destKey)
	return args.Error(0)
}

// MockImageProcessor is a mock implementation of the image processor
type MockImageProcessor struct {
	mock.Mock
}

func (m *MockImageProcessor) ValidateImage(data []byte, contentType string) error {
	args := m.Called(data, contentType)
	return args.Error(0)
}

func (m *MockImageProcessor) ResizeImage(data []byte, maxWidth, maxHeight int) ([]byte, error) {
	args := m.Called(data, maxWidth, maxHeight)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageProcessor) GenerateThumbnail(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageProcessor) OptimizeImage(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageProcessor) AddWatermark(data []byte, text string) ([]byte, error) {
	args := m.Called(data, text)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageProcessor) StripEXIF(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageProcessor) GetImageInfo(data []byte) (*services.ImageInfo, error) {
	args := m.Called(data)
	return args.Get(0).(*services.ImageInfo), args.Error(1)
}

func (m *MockImageProcessor) ConvertFormat(data []byte, format string) ([]byte, error) {
	args := m.Called(data, format)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageProcessor) ProcessImage(data []byte, contentType string, resize, generateThumbnail, optimize bool) (*services.ProcessedImageResult, error) {
	args := m.Called(data, contentType, resize, generateThumbnail, optimize)
	return args.Get(0).(*services.ProcessedImageResult), args.Error(1)
}

func TestNewUploadPhotoUseCase(t *testing.T) {
	mockRepo := &MockPhotoRepository{}
	mockStorage := &MockStorageService{}
	mockProcessor := &MockImageProcessor{}

	useCase := NewUploadPhotoUseCase(mockRepo, mockStorage, mockProcessor)
	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.photoRepo)
	assert.Equal(t, mockStorage, useCase.storageService)
	assert.Equal(t, mockProcessor, useCase.imageProcessor)
}

func TestUploadPhotoUseCase_Execute(t *testing.T) {
	mockRepo := &MockPhotoRepository{}
	mockStorage := &MockStorageService{}
	mockProcessor := &MockImageProcessor{}

	useCase := NewUploadPhotoUseCase(mockRepo, mockStorage, mockProcessor)

	userID := uuid.New()
	imageData := []byte("test image data")
	contentType := "image/jpeg"

	tests := []struct {
		name          string
		setupMocks    func()
		expectError   bool
		errorMsg      string
		expectedPhoto *entities.Photo
	}{
		{
			name: "Successful upload",
			setupMocks: func() {
				mockProcessor.On("ValidateImage", imageData, contentType).Return(nil)
				mockProcessor.On("ProcessImage", imageData, contentType, true, true, true).Return(&services.ProcessedImageResult{
					ProcessedImage: []byte("processed image"),
					Thumbnail:      []byte("thumbnail"),
					Info: &services.ImageInfo{
						Width:  800,
						Height: 600,
						Format: "jpeg",
						Size:   1024,
					},
				}, nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("processed image"), contentType, int64(1024)).Return(nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("thumbnail"), "image/jpeg", mock.AnythingOfType("int64")).Return(nil)
				mockRepo.On("CountUserPhotos", mock.Anything, userID).Return(0, nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Photo")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Invalid image",
			setupMocks: func() {
				mockProcessor.On("ValidateImage", imageData, contentType).Return(errors.New("invalid image"))
			},
			expectError: true,
			errorMsg:    "invalid image",
		},
		{
			name: "Image processing failed",
			setupMocks: func() {
				mockProcessor.On("ValidateImage", imageData, contentType).Return(nil)
				mockProcessor.On("ProcessImage", imageData, contentType, true, true, true).Return(nil, errors.New("processing failed"))
			},
			expectError: true,
			errorMsg:    "processing failed",
		},
		{
			name: "Storage upload failed",
			setupMocks: func() {
				mockProcessor.On("ValidateImage", imageData, contentType).Return(nil)
				mockProcessor.On("ProcessImage", imageData, contentType, true, true, true).Return(&services.ProcessedImageResult{
					ProcessedImage: []byte("processed image"),
					Thumbnail:      []byte("thumbnail"),
					Info: &services.ImageInfo{
						Width:  800,
						Height: 600,
						Format: "jpeg",
						Size:   1024,
					},
				}, nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("processed image"), contentType, int64(1024)).Return(errors.New("upload failed"))
			},
			expectError: true,
			errorMsg:    "upload failed",
		},
		{
			name: "Database creation failed",
			setupMocks: func() {
				mockProcessor.On("ValidateImage", imageData, contentType).Return(nil)
				mockProcessor.On("ProcessImage", imageData, contentType, true, true, true).Return(&services.ProcessedImageResult{
					ProcessedImage: []byte("processed image"),
					Thumbnail:      []byte("thumbnail"),
					Info: &services.ImageInfo{
						Width:  800,
						Height: 600,
						Format: "jpeg",
						Size:   1024,
					},
				}, nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("processed image"), contentType, int64(1024)).Return(nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("thumbnail"), "image/jpeg", mock.AnythingOfType("int64")).Return(nil)
				mockRepo.On("CountUserPhotos", mock.Anything, userID).Return(0, nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Photo")).Return(errors.New("database error"))
			},
			expectError: true,
			errorMsg:    "database error",
		},
		{
			name: "First photo should be primary",
			setupMocks: func() {
				mockProcessor.On("ValidateImage", imageData, contentType).Return(nil)
				mockProcessor.On("ProcessImage", imageData, contentType, true, true, true).Return(&services.ProcessedImageResult{
					ProcessedImage: []byte("processed image"),
					Thumbnail:      []byte("thumbnail"),
					Info: &services.ImageInfo{
						Width:  800,
						Height: 600,
						Format: "jpeg",
						Size:   1024,
					},
				}, nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("processed image"), contentType, int64(1024)).Return(nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("thumbnail"), "image/jpeg", mock.AnythingOfType("int64")).Return(nil)
				mockRepo.On("CountUserPhotos", mock.Anything, userID).Return(0, nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Photo")).Return(nil).Run(func(args mock.Arguments) {
					photo := args.Get(1).(*entities.Photo)
					assert.True(t, photo.IsPrimary)
				})
			},
			expectError: false,
		},
		{
			name: "Non-first photo should not be primary",
			setupMocks: func() {
				mockProcessor.On("ValidateImage", imageData, contentType).Return(nil)
				mockProcessor.On("ProcessImage", imageData, contentType, true, true, true).Return(&services.ProcessedImageResult{
					ProcessedImage: []byte("processed image"),
					Thumbnail:      []byte("thumbnail"),
					Info: &services.ImageInfo{
						Width:  800,
						Height: 600,
						Format: "jpeg",
						Size:   1024,
					},
				}, nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("processed image"), contentType, int64(1024)).Return(nil)
				mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("thumbnail"), "image/jpeg", mock.AnythingOfType("int64")).Return(nil)
				mockRepo.On("CountUserPhotos", mock.Anything, userID).Return(2, nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Photo")).Return(nil).Run(func(args mock.Arguments) {
					photo := args.Get(1).(*entities.Photo)
					assert.False(t, photo.IsPrimary)
				})
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil
			mockStorage.ExpectedCalls = nil
			mockProcessor.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Execute use case
			photo, err := useCase.Execute(context.Background(), userID, imageData, contentType)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, photo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, photo)
				assert.Equal(t, userID, photo.UserID)
				assert.Equal(t, contentType, photo.ContentType)
				assert.Equal(t, 800, photo.Width)
				assert.Equal(t, 600, photo.Height)
				assert.Equal(t, int64(1024), photo.FileSize)
				assert.NotEmpty(t, photo.StorageKey)
				assert.NotEmpty(t, photo.ThumbnailKey)
				assert.False(t, photo.CreatedAt.IsZero())
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
			mockProcessor.AssertExpectations(t)
		})
	}
}

func TestUploadPhotoUseCase_Execute_WithVerificationPhoto(t *testing.T) {
	mockRepo := &MockPhotoRepository{}
	mockStorage := &MockStorageService{}
	mockProcessor := &MockImageProcessor{}

	useCase := NewUploadPhotoUseCase(mockRepo, mockStorage, mockProcessor)

	userID := uuid.New()
	imageData := []byte("test image data")
	contentType := "image/jpeg"
	isVerification := true

	// Setup mocks
	mockProcessor.On("ValidateImage", imageData, contentType).Return(nil)
	mockProcessor.On("ProcessImage", imageData, contentType, true, true, true).Return(&services.ProcessedImageResult{
		ProcessedImage: []byte("processed image"),
		Thumbnail:      []byte("thumbnail"),
		Info: &services.ImageInfo{
			Width:  800,
			Height: 600,
			Format: "jpeg",
			Size:   1024,
		},
	}, nil)
	mockProcessor.On("AddWatermark", []byte("processed image"), "© Verified").Return([]byte("watermarked image"), nil)
	mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("watermarked image"), contentType, int64(1024)).Return(nil)
	mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("thumbnail"), "image/jpeg", mock.AnythingOfType("int64")).Return(nil)
	mockRepo.On("CountUserPhotos", mock.Anything, userID).Return(0, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Photo")).Return(nil).Run(func(args mock.Arguments) {
		photo := args.Get(1).(*entities.Photo)
		assert.True(t, photo.IsVerification)
	})

	// Execute use case
	photo, err := useCase.Execute(context.Background(), userID, imageData, contentType, isVerification)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, photo)
	assert.True(t, photo.IsVerification)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockProcessor.AssertExpectations(t)
}

func TestUploadPhotoUseCase_Execute_WithEphemeralPhoto(t *testing.T) {
	mockRepo := &MockPhotoRepository{}
	mockStorage := &MockStorageService{}
	mockProcessor := &MockImageProcessor{}

	useCase := NewUploadPhotoUseCase(mockRepo, mockStorage, mockProcessor)

	userID := uuid.New()
	imageData := []byte("test image data")
	contentType := "image/jpeg"
	isEphemeral := true
	viewCount := 5

	// Setup mocks
	mockProcessor.On("ValidateImage", imageData, contentType).Return(nil)
	mockProcessor.On("ProcessImage", imageData, contentType, true, true, true).Return(&services.ProcessedImageResult{
		ProcessedImage: []byte("processed image"),
		Thumbnail:      []byte("thumbnail"),
		Info: &services.ImageInfo{
			Width:  800,
			Height: 600,
			Format: "jpeg",
			Size:   1024,
		},
	}, nil)
	mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("processed image"), contentType, int64(1024)).Return(nil)
	mockStorage.On("UploadFile", mock.AnythingOfType("string"), []byte("thumbnail"), "image/jpeg", mock.AnythingOfType("int64")).Return(nil)
	mockRepo.On("CountUserPhotos", mock.Anything, userID).Return(0, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Photo")).Return(nil).Run(func(args mock.Arguments) {
		photo := args.Get(1).(*entities.Photo)
		assert.True(t, photo.IsEphemeral)
		assert.Equal(t, viewCount, photo.MaxViewCount)
		assert.Equal(t, 0, photo.ViewCount)
		assert.False(t, photo.ExpiresAt.IsZero())
	})

	// Execute use case
	photo, err := useCase.Execute(context.Background(), userID, imageData, contentType, false, isEphemeral, &viewCount)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, photo)
	assert.True(t, photo.IsEphemeral)
	assert.Equal(t, viewCount, photo.MaxViewCount)
	assert.Equal(t, 0, photo.ViewCount)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockProcessor.AssertExpectations(t)
}