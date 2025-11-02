package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"golang.org/x/image/webp"
)

// ImageProcessingService defines interface for image processing operations
type ImageProcessingService interface {
	// ValidateImage validates an image file
	ValidateImage(ctx context.Context, file io.Reader) (*ImageValidationResult, error)
	
	// ProcessImage processes an image for storage
	ProcessImage(ctx context.Context, file io.Reader, options *ProcessOptions) (*ProcessResult, error)
	
	// GenerateThumbnail generates a thumbnail from an image
	GenerateThumbnail(ctx context.Context, file io.Reader, width, height int) ([]byte, error)
	
	// ResizeImage resizes an image to specified dimensions
	ResizeImage(ctx context.Context, file io.Reader, width, height int) ([]byte, error)
	
	// OptimizeImage optimizes an image for web
	OptimizeImage(ctx context.Context, file io.Reader, quality int) ([]byte, error)
	
	// AddWatermark adds a watermark to an image
	AddWatermark(ctx context.Context, file io.Reader, watermarkText string) ([]byte, error)
	
	// StripEXIF removes EXIF data from an image
	StripEXIF(ctx context.Context, file io.Reader) ([]byte, error)
	
	// DetectContent detects inappropriate content (basic implementation)
	DetectContent(ctx context.Context, file io.Reader) (*ContentDetectionResult, error)
	
	// GetImageInfo extracts image information
	GetImageInfo(ctx context.Context, file io.Reader) (*ImageInfo, error)
}

// ImageValidationResult represents result of image validation
type ImageValidationResult struct {
	IsValid   bool     `json:"is_valid"`
	Format    string   `json:"format"`
	Size      int64    `json:"size"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	Errors    []string `json:"errors"`
	Warnings  []string `json:"warnings"`
}

// ProcessOptions represents options for image processing
type ProcessOptions struct {
	ResizeWidth     int    `json:"resize_width"`
	ResizeHeight    int    `json:"resize_height"`
	Quality         int    `json:"quality"`
	GenerateThumb   bool   `json:"generate_thumb"`
	ThumbWidth      int    `json:"thumb_width"`
	ThumbHeight     int    `json:"thumb_height"`
	StripEXIF       bool   `json:"strip_exif"`
	AddWatermark    bool   `json:"add_watermark"`
	WatermarkText   string `json:"watermark_text"`
	Optimize        bool   `json:"optimize"`
}

// ProcessResult represents result of image processing
type ProcessResult struct {
	OriginalKey     string `json:"original_key"`
	ProcessedKey    string `json:"processed_key"`
	ThumbnailKey    string `json:"thumbnail_key"`
	OriginalSize    int64  `json:"original_size"`
	ProcessedSize   int64  `json:"processed_size"`
	ThumbnailSize   int64  `json:"thumbnail_size"`
	Format          string `json:"format"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	ProcessingTime   int64   `json:"processing_time_ms"`
}

// ContentDetectionResult represents result of content detection
type ContentDetectionResult struct {
	IsAppropriate bool     `json:"is_appropriate"`
	Confidence   float64  `json:"confidence"`
	Labels       []string  `json:"labels"`
	Warnings     []string  `json:"warnings"`
}

// ImageInfo represents basic image information
type ImageInfo struct {
	Format     string `json:"format"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Size       int64  `json:"size"`
	HasAlpha   bool   `json:"has_alpha"`
	ColorSpace  string `json:"color_space"`
}

// ImageProcessor implements ImageProcessingService
type ImageProcessor struct {
	config *config.StorageConfig
}

// NewImageProcessor creates a new image processing service
func NewImageProcessor(cfg *config.StorageConfig) ImageProcessingService {
	return &ImageProcessor{
		config: cfg,
	}
}

// ValidateImage validates an image file
func (p *ImageProcessor) ValidateImage(ctx context.Context, file io.Reader) (*ImageValidationResult, error) {
	result := &ImageValidationResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	// Read file to buffer for multiple operations
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	result.Size = size

	// Check file size
	if size > p.config.MaxFileSize {
		result.Errors = append(result.Errors, fmt.Sprintf("file size %d exceeds maximum allowed size %d", size, p.config.MaxFileSize))
	}

	// Decode image to validate format and get dimensions
	img, format, err := image.Decode(buf)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid image format: %v", err))
		result.IsValid = false
		return result, nil
	}

	result.Format = format
	bounds := img.Bounds()
	result.Width = bounds.Dx()
	result.Height = bounds.Dy()

	// Check if format is allowed
	if !p.isFormatAllowed(format) {
		result.Errors = append(result.Errors, fmt.Sprintf("format %s is not allowed", format))
	}

	// Check minimum dimensions
	if result.Width < 100 || result.Height < 100 {
		result.Warnings = append(result.Warnings, "image dimensions are very small (minimum 100x100 recommended)")
	}

	// Check maximum dimensions
	if result.Width > 4096 || result.Height > 4096 {
		result.Warnings = append(result.Warnings, "image dimensions are very large (maximum 4096x4096 recommended)")
	}

	// Check aspect ratio
	aspectRatio := float64(result.Width) / float64(result.Height)
	if aspectRatio < 0.3 || aspectRatio > 3.0 {
		result.Warnings = append(result.Warnings, "image aspect ratio is unusual (recommended between 0.3 and 3.0)")
	}

	result.IsValid = len(result.Errors) == 0

	logger.Info("Image validation completed", map[string]interface{}{
		"format":   format,
		"size":     size,
		"width":    result.Width,
		"height":   result.Height,
		"is_valid": result.IsValid,
		"errors":   len(result.Errors),
	})

	return result, nil
}

// ProcessImage processes an image for storage
func (p *ImageProcessor) ProcessImage(ctx context.Context, file io.Reader, options *ProcessOptions) (*ProcessResult, error) {
	startTime := time.Now()

	// Set default options if not provided
	if options == nil {
		options = &ProcessOptions{
			Quality:       85,
			GenerateThumb:  true,
			ThumbWidth:     200,
			ThumbHeight:    200,
			StripEXIF:      true,
			Optimize:       true,
		}
	}

	// Read file to buffer
	buf := new(bytes.Buffer)
	originalSize, err := io.Copy(buf, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Decode image
	img, format, err := image.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	result := &ProcessResult{
		OriginalSize:  originalSize,
		Format:        format,
		Width:         img.Bounds().Dx(),
		Height:        img.Bounds().Dy(),
		OriginalKey:   uuid.New().String(),
	}

	// Process main image
	processedImg := img
	if options.ResizeWidth > 0 && options.ResizeHeight > 0 {
		processedImg = p.resizeImage(processedImg, options.ResizeWidth, options.ResizeHeight)
		result.Width = options.ResizeWidth
		result.Height = options.ResizeHeight
	}

	if options.StripEXIF {
		// EXIF is already stripped during re-encoding
	}

	if options.AddWatermark && options.WatermarkText != "" {
		processedImg = p.addWatermark(processedImg, options.WatermarkText)
	}

	// Encode processed image
	processedBuf := new(bytes.Buffer)
	if err := p.encodeImage(processedBuf, processedImg, format, options.Quality); err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %w", err)
	}

	result.ProcessedSize = int64(processedBuf.Len())
	result.ProcessedKey = uuid.New().String()

	// Generate thumbnail if requested
	if options.GenerateThumb {
		thumbBuf, err := p.GenerateThumbnail(ctx, bytes.NewReader(buf.Bytes()), options.ThumbWidth, options.ThumbHeight)
		if err != nil {
			logger.Error("Failed to generate thumbnail", err)
		} else {
			result.ThumbnailSize = int64(len(thumbBuf))
			result.ThumbnailKey = uuid.New().String()
		}
	}

	result.ProcessingTime = time.Since(startTime).Milliseconds()

	logger.Info("Image processing completed", map[string]interface{}{
		"original_size":   result.OriginalSize,
		"processed_size":  result.ProcessedSize,
		"thumbnail_size":  result.ThumbnailSize,
		"processing_time": result.ProcessingTime,
	})

	return result, nil
}

// GenerateThumbnail generates a thumbnail from an image
func (p *ImageProcessor) GenerateThumbnail(ctx context.Context, file io.Reader, width, height int) ([]byte, error) {
	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Generate thumbnail using smart cropping
	thumbnail := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	// Encode thumbnail
	buf := new(bytes.Buffer)
	if err := p.encodeImage(buf, thumbnail, format, 85); err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// ResizeImage resizes an image to specified dimensions
func (p *ImageProcessor) ResizeImage(ctx context.Context, file io.Reader, width, height int) ([]byte, error) {
	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image maintaining aspect ratio
	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	// Encode resized image
	buf := new(bytes.Buffer)
	if err := p.encodeImage(buf, resized, format, 85); err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return buf.Bytes(), nil
}

// OptimizeImage optimizes an image for web
func (p *ImageProcessor) OptimizeImage(ctx context.Context, file io.Reader, quality int) ([]byte, error) {
	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Encode with optimization
	buf := new(bytes.Buffer)
	if err := p.encodeImage(buf, img, format, quality); err != nil {
		return nil, fmt.Errorf("failed to encode optimized image: %w", err)
	}

	return buf.Bytes(), nil
}

// AddWatermark adds a watermark to an image
func (p *ImageProcessor) AddWatermark(ctx context.Context, file io.Reader, watermarkText string) ([]byte, error) {
	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Add watermark
	watermarked := p.addWatermark(img, watermarkText)

	// Encode watermarked image
	buf := new(bytes.Buffer)
	if err := p.encodeImage(buf, watermarked, format, 85); err != nil {
		return nil, fmt.Errorf("failed to encode watermarked image: %w", err)
	}

	return buf.Bytes(), nil
}

// StripEXIF removes EXIF data from an image
func (p *ImageProcessor) StripEXIF(ctx context.Context, file io.Reader) ([]byte, error) {
	// Decode image (this automatically strips EXIF)
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Re-encode image without EXIF
	buf := new(bytes.Buffer)
	if err := p.encodeImage(buf, img, format, 85); err != nil {
		return nil, fmt.Errorf("failed to encode image without EXIF: %w", err)
	}

	return buf.Bytes(), nil
}

// DetectContent detects inappropriate content (basic implementation)
func (p *ImageProcessor) DetectContent(ctx context.Context, file io.Reader) (*ContentDetectionResult, error) {
	result := &ContentDetectionResult{
		IsAppropriate: true,
		Confidence:     0.95,
		Labels:        []string{},
		Warnings:      []string{},
	}

	// Decode image to get basic properties
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Basic content analysis (this is a simplified implementation)
	// In production, you would integrate with AWS Rekognition or similar service
	
	// Check for very small images (could be inappropriate)
	if width < 50 || height < 50 {
		result.Warnings = append(result.Warnings, "very small image detected")
		result.Confidence = math.Max(result.Confidence-0.1, 0.5)
	}

	// Check for unusual aspect ratios
	aspectRatio := float64(width) / float64(height)
	if aspectRatio < 0.2 || aspectRatio > 5.0 {
		result.Warnings = append(result.Warnings, "unusual aspect ratio detected")
		result.Confidence = math.Max(result.Confidence-0.05, 0.5)
	}

	logger.Info("Content detection completed", map[string]interface{}{
		"confidence":     result.Confidence,
		"is_appropriate": result.IsAppropriate,
		"warnings":       len(result.Warnings),
	})

	return result, nil
}

// GetImageInfo extracts image information
func (p *ImageProcessor) GetImageInfo(ctx context.Context, file io.Reader) (*ImageInfo, error) {
	// Get file size
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Decode image to get format and dimensions
	img, format, err := image.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	
	// Check if image has alpha channel
	hasAlpha := false
	switch img.(type) {
	case *image.NRGBA, *image.RGBA:
		hasAlpha = true
	}

	return &ImageInfo{
		Format:    format,
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		Size:      size,
		HasAlpha:  hasAlpha,
		ColorSpace: "sRGB", // Default assumption
	}, nil
}

// Helper methods

// isFormatAllowed checks if image format is allowed
func (p *ImageProcessor) isFormatAllowed(format string) bool {
	for _, allowedType := range p.config.AllowedTypes {
		if strings.Contains(allowedType, format) {
			return true
		}
	}
	return false
}

// resizeImage resizes an image maintaining aspect ratio
func (p *ImageProcessor) resizeImage(img image.Image, width, height int) image.Image {
	return imaging.Resize(img, width, height, imaging.Lanczos)
}

// addWatermark adds a text watermark to an image
func (p *ImageProcessor) addWatermark(img image.Image, text string) image.Image {
	// This is a simplified watermark implementation
	// In production, you might want to use a more sophisticated approach
	bounds := img.Bounds()
	
	// Create a new image with watermark
	watermarked := imaging.Clone(img)
	
	// For now, just return the original image
	// TODO: Implement proper watermarking
	return watermarked
}

// encodeImage encodes an image to the specified format
func (p *ImageProcessor) encodeImage(w io.Writer, img image.Image, format string, quality int) error {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	case "png":
		return png.Encode(w, img)
	case "webp":
		return webp.Encode(w, img, &webp.Options{Quality: float32(quality)})
	default:
		// Default to JPEG
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	}
}