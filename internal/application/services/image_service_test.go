package services

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/22smeargle/winkr-backend/pkg/config"
)

// createTestImage creates a test image for testing purposes
func createTestImage(width, height int, format string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Fill with some color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, image.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
		return buf.Bytes(), err
	case "png":
		err := png.Encode(&buf, img)
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func TestNewImageProcessor(t *testing.T) {
	cfg := &config.StorageConfig{
		MaxWidth:     2048,
		MaxHeight:    2048,
		ThumbnailSize: 200,
		Quality:      85,
		MaxFileSize:  10 * 1024 * 1024, // 10MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
	}

	processor := NewImageProcessor(cfg)
	assert.NotNil(t, processor)
	assert.Equal(t, cfg, processor.config)
}

func TestImageProcessor_ValidateImage(t *testing.T) {
	cfg := &config.StorageConfig{
		MaxWidth:     2048,
		MaxHeight:    2048,
		MaxFileSize:  10 * 1024 * 1024, // 10MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
	}

	processor := NewImageProcessor(cfg)

	tests := []struct {
		name        string
		imageData   []byte
		contentType string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid JPEG",
			imageData:   func() []byte { data, _ := createTestImage(800, 600, "jpeg"); return data }(),
			contentType: "image/jpeg",
			expectError: false,
		},
		{
			name:        "Valid PNG",
			imageData:   func() []byte { data, _ := createTestImage(800, 600, "png"); return data }(),
			contentType: "image/png",
			expectError: false,
		},
		{
			name:        "Invalid content type",
			imageData:   func() []byte { data, _ := createTestImage(800, 600, "jpeg"); return data }(),
			contentType: "application/pdf",
			expectError: true,
			errorMsg:    "content type application.pdf is not allowed",
		},
		{
			name:        "File too large",
			imageData:   make([]byte, 20*1024*1024), // 20MB
			contentType: "image/jpeg",
			expectError: true,
			errorMsg:    "file size exceeds maximum allowed size",
		},
		{
			name:        "Invalid image data",
			imageData:   []byte("not an image"),
			contentType: "image/jpeg",
			expectError: true,
			errorMsg:    "failed to decode image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateImage(tt.imageData, tt.contentType)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImageProcessor_ResizeImage(t *testing.T) {
	cfg := &config.StorageConfig{
		MaxWidth:     2048,
		MaxHeight:    2048,
		ThumbnailSize: 200,
		Quality:      85,
	}

	processor := NewImageProcessor(cfg)

	tests := []struct {
		name           string
		originalWidth  int
		originalHeight int
		maxWidth       int
		maxHeight      int
		expectResize   bool
	}{
		{
			name:           "Image within limits",
			originalWidth:  800,
			originalHeight: 600,
			maxWidth:       2048,
			maxHeight:      2048,
			expectResize:   false,
		},
		{
			name:           "Image too wide",
			originalWidth:  3000,
			originalHeight: 2000,
			maxWidth:       2048,
			maxHeight:      2048,
			expectResize:   true,
		},
		{
			name:           "Image too tall",
			originalWidth:  1500,
			originalHeight: 3000,
			maxWidth:       2048,
			maxHeight:      2048,
			expectResize:   true,
		},
		{
			name:           "Image needs both width and height reduction",
			originalWidth:  3000,
			originalHeight: 4000,
			maxWidth:       2048,
			maxHeight:      2048,
			expectResize:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image
			imageData, err := createTestImage(tt.originalWidth, tt.originalHeight, "jpeg")
			require.NoError(t, err)

			// Resize image
			resizedData, err := processor.ResizeImage(imageData, tt.maxWidth, tt.maxHeight)
			assert.NoError(t, err)
			assert.NotNil(t, resizedData)

			// Decode the resized image to check dimensions
			resizedImg, _, err := image.Decode(bytes.NewReader(resizedData))
			assert.NoError(t, err)

			bounds := resizedImg.Bounds()
			newWidth := bounds.Dx()
			newHeight := bounds.Dy()

			if tt.expectResize {
				// Should be resized to fit within limits
				assert.LessOrEqual(t, newWidth, tt.maxWidth)
				assert.LessOrEqual(t, newHeight, tt.maxHeight)
			} else {
				// Should maintain original size
				assert.Equal(t, tt.originalWidth, newWidth)
				assert.Equal(t, tt.originalHeight, newHeight)
			}
		})
	}
}

func TestImageProcessor_GenerateThumbnail(t *testing.T) {
	cfg := &config.StorageConfig{
		ThumbnailSize: 200,
		Quality:       85,
	}

	processor := NewImageProcessor(cfg)

	// Create test image
	imageData, err := createTestImage(800, 600, "jpeg")
	require.NoError(t, err)

	// Generate thumbnail
	thumbnailData, err := processor.GenerateThumbnail(imageData)
	assert.NoError(t, err)
	assert.NotNil(t, thumbnailData)

	// Decode the thumbnail to check dimensions
	thumbnailImg, _, err := image.Decode(bytes.NewReader(thumbnailData))
	assert.NoError(t, err)

	bounds := thumbnailImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Thumbnail should be square and within size limits
	assert.Equal(t, cfg.ThumbnailSize, width)
	assert.Equal(t, cfg.ThumbnailSize, height)
}

func TestImageProcessor_OptimizeImage(t *testing.T) {
	cfg := &config.StorageConfig{
		Quality: 85,
	}

	processor := NewImageProcessor(cfg)

	// Create test image
	imageData, err := createTestImage(800, 600, "jpeg")
	require.NoError(t, err)

	// Optimize image
	optimizedData, err := processor.OptimizeImage(imageData)
	assert.NoError(t, err)
	assert.NotNil(t, optimizedData)

	// Optimized image should be smaller or equal in size
	assert.LessOrEqual(t, len(optimizedData), len(imageData))
}

func TestImageProcessor_AddWatermark(t *testing.T) {
	cfg := &config.StorageConfig{
		Quality: 85,
	}

	processor := NewImageProcessor(cfg)

	// Create test image
	imageData, err := createTestImage(800, 600, "jpeg")
	require.NoError(t, err)

	// Add watermark
	watermarkedData, err := processor.AddWatermark(imageData, "© Test App")
	assert.NoError(t, err)
	assert.NotNil(t, watermarkedData)

	// Watermarked image should be different from original
	assert.NotEqual(t, imageData, watermarkedData)
}

func TestImageProcessor_StripEXIF(t *testing.T) {
	cfg := &config.StorageConfig{
		Quality: 85,
	}

	processor := NewImageProcessor(cfg)

	// Create test image
	imageData, err := createTestImage(800, 600, "jpeg")
	require.NoError(t, err)

	// Strip EXIF data
	strippedData, err := processor.StripEXIF(imageData)
	assert.NoError(t, err)
	assert.NotNil(t, strippedData)

	// Stripped image should still be a valid image
	strippedImg, _, err := image.Decode(bytes.NewReader(strippedData))
	assert.NoError(t, err)
	assert.NotNil(t, strippedImg)
}

func TestImageProcessor_GetImageInfo(t *testing.T) {
	cfg := &config.StorageConfig{}

	processor := NewImageProcessor(cfg)

	tests := []struct {
		name          string
		width         int
		height        int
		format        string
		expectedWidth int
		expectedHeight int
	}{
		{
			name:           "JPEG image",
			width:          800,
			height:         600,
			format:         "jpeg",
			expectedWidth:  800,
			expectedHeight: 600,
		},
		{
			name:           "PNG image",
			width:          1024,
			height:         768,
			format:         "png",
			expectedWidth:  1024,
			expectedHeight: 768,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image
			imageData, err := createTestImage(tt.width, tt.height, tt.format)
			require.NoError(t, err)

			// Get image info
			info, err := processor.GetImageInfo(imageData)
			assert.NoError(t, err)
			assert.NotNil(t, info)
			assert.Equal(t, tt.expectedWidth, info.Width)
			assert.Equal(t, tt.expectedHeight, info.Height)
			assert.Equal(t, tt.format, info.Format)
			assert.Greater(t, info.Size, int64(0))
		})
	}
}

func TestImageProcessor_ConvertFormat(t *testing.T) {
	cfg := &config.StorageConfig{
		Quality: 85,
	}

	processor := NewImageProcessor(cfg)

	// Create test image in JPEG format
	imageData, err := createTestImage(800, 600, "jpeg")
	require.NoError(t, err)

	// Convert to PNG
	convertedData, err := processor.ConvertFormat(imageData, "png")
	assert.NoError(t, err)
	assert.NotNil(t, convertedData)

	// Verify the converted image is valid PNG
	convertedImg, format, err := image.Decode(bytes.NewReader(convertedData))
	assert.NoError(t, err)
	assert.Equal(t, "png", format)
	assert.NotNil(t, convertedImg)
}

func TestImageProcessor_ProcessImage(t *testing.T) {
	cfg := &config.StorageConfig{
		MaxWidth:      2048,
		MaxHeight:     2048,
		ThumbnailSize: 200,
		Quality:       85,
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		AllowedTypes:  []string{"image/jpeg", "image/png", "image/webp"},
	}

	processor := NewImageProcessor(cfg)

	// Create test image
	imageData, err := createTestImage(3000, 2000, "jpeg") // Large image that needs resizing
	require.NoError(t, err)

	// Process image
	result, err := processor.ProcessImage(imageData, "image/jpeg", true, true, true)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.ProcessedImage)
	assert.NotNil(t, result.Thumbnail)
	assert.NotNil(t, result.Info)

	// Check that the image was processed correctly
	assert.LessOrEqual(t, result.Info.Width, cfg.MaxWidth)
	assert.LessOrEqual(t, result.Info.Height, cfg.MaxHeight)
	assert.Equal(t, cfg.ThumbnailSize, result.ThumbnailInfo.Width)
	assert.Equal(t, cfg.ThumbnailSize, result.ThumbnailInfo.Height)
}

// TestImageProcessor_Integration tests the full workflow
func TestImageProcessor_Integration(t *testing.T) {
	cfg := &config.StorageConfig{
		MaxWidth:      2048,
		MaxHeight:     2048,
		ThumbnailSize: 200,
		Quality:       85,
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		AllowedTypes:  []string{"image/jpeg", "image/png", "image/webp"},
	}

	processor := NewImageProcessor(cfg)

	// Create test image
	imageData, err := createTestImage(3000, 2000, "jpeg")
	require.NoError(t, err)

	// Validate image
	err = processor.ValidateImage(imageData, "image/jpeg")
	assert.NoError(t, err)

	// Get image info
	info, err := processor.GetImageInfo(imageData)
	assert.NoError(t, err)
	assert.Equal(t, 3000, info.Width)
	assert.Equal(t, 2000, info.Height)

	// Resize image
	resizedData, err := processor.ResizeImage(imageData, cfg.MaxWidth, cfg.MaxHeight)
	assert.NoError(t, err)

	// Verify resized image
	resizedInfo, err := processor.GetImageInfo(resizedData)
	assert.NoError(t, err)
	assert.LessOrEqual(t, resizedInfo.Width, cfg.MaxWidth)
	assert.LessOrEqual(t, resizedInfo.Height, cfg.MaxHeight)

	// Generate thumbnail
	thumbnailData, err := processor.GenerateThumbnail(resizedData)
	assert.NoError(t, err)

	// Verify thumbnail
	thumbnailInfo, err := processor.GetImageInfo(thumbnailData)
	assert.NoError(t, err)
	assert.Equal(t, cfg.ThumbnailSize, thumbnailInfo.Width)
	assert.Equal(t, cfg.ThumbnailSize, thumbnailInfo.Height)

	// Optimize image
	optimizedData, err := processor.OptimizeImage(resizedData)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(optimizedData), len(resizedData))

	// Add watermark
	watermarkedData, err := processor.AddWatermark(optimizedData, "© Test App")
	assert.NoError(t, err)
	assert.NotEqual(t, optimizedData, watermarkedData)

	// Strip EXIF
	strippedData, err := processor.StripEXIF(watermarkedData)
	assert.NoError(t, err)
	assert.NotNil(t, strippedData)
}

// BenchmarkImageProcessor_ResizeImage benchmarks the resize operation
func BenchmarkImageProcessor_ResizeImage(b *testing.B) {
	cfg := &config.StorageConfig{
		MaxWidth:  2048,
		MaxHeight: 2048,
		Quality:   85,
	}

	processor := NewImageProcessor(cfg)

	// Create test image
	imageData, err := createTestImage(3000, 2000, "jpeg")
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ResizeImage(imageData, cfg.MaxWidth, cfg.MaxHeight)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkImageProcessor_GenerateThumbnail benchmarks the thumbnail generation
func BenchmarkImageProcessor_GenerateThumbnail(b *testing.B) {
	cfg := &config.StorageConfig{
		ThumbnailSize: 200,
		Quality:       85,
	}

	processor := NewImageProcessor(cfg)

	// Create test image
	imageData, err := createTestImage(800, 600, "jpeg")
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.GenerateThumbnail(imageData)
		if err != nil {
			b.Fatal(err)
		}
	}
}