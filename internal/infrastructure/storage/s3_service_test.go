package storage

import (
	"context"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/22smeargle/winkr-backend/pkg/config"
)

// MockS3Client is a mock implementation of the S3 client
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

func (m *MockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.HeadObjectOutput), args.Error(1)
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

func (m *MockS3Client) CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.CopyObjectOutput), args.Error(1)
}

func (m *MockS3Client) PresignClient() *s3.PresignClient {
	args := m.Called()
	return args.Get(0).(*s3.PresignClient)
}

// MockPresignClient is a mock implementation of the S3 presign client
type MockPresignClient struct {
	mock.Mock
}

func (m *MockPresignClient) PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*aws.PresignedURLRequest, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*aws.PresignedURLRequest), args.Error(1)
}

func (m *MockPresignClient) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*aws.PresignedURLRequest, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*aws.PresignedURLRequest), args.Error(1)
}

func TestNewS3Storage(t *testing.T) {
	cfg := &config.StorageConfig{
		Provider:      "aws",
		Bucket:        "test-bucket",
		Region:        "us-east-1",
		AccessKey:     "test-access-key",
		SecretKey:     "test-secret-key",
		Endpoint:      "",
		UseSSL:        true,
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		AllowedTypes:  []string{"image/jpeg", "image/png", "image/webp"},
		UploadExpiry:  15 * time.Minute,
		DownloadExpiry: 30 * time.Minute,
	}

	storage, err := NewS3Storage(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, cfg, storage.config)
}

func TestS3Storage_ValidateFile(t *testing.T) {
	cfg := &config.StorageConfig{
		MaxFileSize:  10 * 1024 * 1024, // 10MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
	}

	storage := &S3Storage{
		config: cfg,
	}

	tests := []struct {
		name        string
		contentType string
		size        int64
		expectError bool
	}{
		{
			name:        "Valid JPEG",
			contentType: "image/jpeg",
			size:        1024,
			expectError: false,
		},
		{
			name:        "Valid PNG",
			contentType: "image/png",
			size:        2048,
			expectError: false,
		},
		{
			name:        "Valid WebP",
			contentType: "image/webp",
			size:        3072,
			expectError: false,
		},
		{
			name:        "Invalid content type",
			contentType: "application/pdf",
			size:        1024,
			expectError: true,
		},
		{
			name:        "File too large",
			contentType: "image/jpeg",
			size:        20 * 1024 * 1024, // 20MB
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.ValidateFile(tt.contentType, tt.size)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestS3Storage_GenerateUploadURL(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket:        "test-bucket",
		Region:        "us-east-1",
		UploadExpiry:  15 * time.Minute,
		AllowedTypes:  []string{"image/jpeg", "image/png", "image/webp"},
	}

	mockS3Client := &MockS3Client{}
	mockPresignClient := &MockPresignClient{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
		presigner: mockPresignClient,
	}

	// Setup mock expectations
	expectedURL := "https://test-bucket.s3.amazonaws.com/test-key?X-Amz-SignedHeaders=host"
	mockPresignClient.On("PresignPutObject", mock.Anything, mock.Anything, mock.Anything).Return(&aws.PresignedURLRequest{
		URL: &url.URL{String: expectedURL},
	}, nil)

	mockS3Client.On("PresignClient").Return(mockPresignClient)

	url, err := storage.GenerateUploadURL("test-key", "image/jpeg")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)

	mockS3Client.AssertExpectations(t)
	mockPresignClient.AssertExpectations(t)
}

func TestS3Storage_GenerateDownloadURL(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket:         "test-bucket",
		Region:         "us-east-1",
		DownloadExpiry: 30 * time.Minute,
	}

	mockS3Client := &MockS3Client{}
	mockPresignClient := &MockPresignClient{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
		presigner: mockPresignClient,
	}

	// Setup mock expectations
	expectedURL := "https://test-bucket.s3.amazonaws.com/test-key?X-Amz-SignedHeaders=host"
	mockPresignClient.On("PresignGetObject", mock.Anything, mock.Anything, mock.Anything).Return(&aws.PresignedURLRequest{
		URL: &url.URL{String: expectedURL},
	}, nil)

	mockS3Client.On("PresignClient").Return(mockPresignClient)

	url, err := storage.GenerateDownloadURL("test-key")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)

	mockS3Client.AssertExpectations(t)
	mockPresignClient.AssertExpectations(t)
}

func TestS3Storage_UploadFile(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	mockS3Client := &MockS3Client{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
	}

	// Setup mock expectations
	mockS3Client.On("PutObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.PutObjectOutput{
		ETag: aws.String("test-etag"),
	}, nil)

	content := strings.NewReader("test content")
	err := storage.UploadFile("test-key", content, "image/jpeg", 12)
	assert.NoError(t, err)

	mockS3Client.AssertExpectations(t)
}

func TestS3Storage_DownloadFile(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	mockS3Client := &MockS3Client{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
	}

	// Setup mock expectations
	testContent := "test content"
	mockS3Client.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader(testContent)),
		ContentLength: aws.Int64(int64(len(testContent))),
		ContentType:   aws.String("image/jpeg"),
	}, nil)

	reader, err := storage.DownloadFile("test-key")
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	// Read the content to verify
	content, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, testContent, string(content))

	// Close the reader
	reader.Close()

	mockS3Client.AssertExpectations(t)
}

func TestS3Storage_DeleteFile(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	mockS3Client := &MockS3Client{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
	}

	// Setup mock expectations
	mockS3Client.On("DeleteObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.DeleteObjectOutput{}, nil)

	err := storage.DeleteFile("test-key")
	assert.NoError(t, err)

	mockS3Client.AssertExpectations(t)
}

func TestS3Storage_FileExists(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	mockS3Client := &MockS3Client{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
	}

	tests := []struct {
		name        string
		exists      bool
		expectError bool
	}{
		{
			name:        "File exists",
			exists:      true,
			expectError: false,
		},
		{
			name:        "File does not exist",
			exists:      false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exists {
				// Setup mock expectations for existing file
				mockS3Client.On("HeadObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.HeadObjectOutput{
					ContentLength: aws.Int64(1024),
					ContentType:   aws.String("image/jpeg"),
				}, nil)
			} else {
				// Setup mock expectations for non-existing file
				mockS3Client.On("HeadObject", mock.Anything, mock.Anything, mock.Anything).Return(nil, &types.NotFound{
					Message: aws.String("Not Found"),
				})
			}

			exists, err := storage.FileExists("test-key")
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.exists, exists)
			}

			mockS3Client.AssertExpectations(t)
		})
	}
}

func TestS3Storage_GetFileInfo(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	mockS3Client := &MockS3Client{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
	}

	// Setup mock expectations
	mockS3Client.On("HeadObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.HeadObjectOutput{
		ContentLength: aws.Int64(1024),
		ContentType:   aws.String("image/jpeg"),
		LastModified:  aws.Time(time.Now()),
		ETag:          aws.String("test-etag"),
	}, nil)

	info, err := storage.GetFileInfo("test-key")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, int64(1024), info.Size)
	assert.Equal(t, "image/jpeg", info.ContentType)
	assert.Equal(t, "test-etag", info.ETag)

	mockS3Client.AssertExpectations(t)
}

func TestS3Storage_ListFiles(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	mockS3Client := &MockS3Client{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
	}

	// Setup mock expectations
	mockS3Client.On("ListObjectsV2", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{
			{
				Key:          aws.String("test-key-1"),
				Size:         aws.Int64(1024),
				LastModified: aws.Time(time.Now()),
				ETag:         aws.String("test-etag-1"),
			},
			{
				Key:          aws.String("test-key-2"),
				Size:         aws.Int64(2048),
				LastModified: aws.Time(time.Now()),
				ETag:         aws.String("test-etag-2"),
			},
		},
	}, nil)

	files, err := storage.ListFiles("test-prefix")
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, "test-key-1", files[0].Key)
	assert.Equal(t, "test-key-2", files[1].Key)

	mockS3Client.AssertExpectations(t)
}

func TestS3Storage_CopyFile(t *testing.T) {
	cfg := &config.StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	mockS3Client := &MockS3Client{}
	
	storage := &S3Storage{
		config: cfg,
		client: mockS3Client,
	}

	// Setup mock expectations
	mockS3Client.On("CopyObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.CopyObjectOutput{}, nil)

	err := storage.CopyFile("source-key", "dest-key")
	assert.NoError(t, err)

	mockS3Client.AssertExpectations(t)
}