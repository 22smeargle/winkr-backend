package external

import (
	"context"
	"fmt"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// MockAIService is a mock implementation of AIService for testing
type MockAIService struct {
	CalledAnalyzeFace    bool
	CalledCompareFaces    bool
	CalledDetectNSFW      bool
	CalledDetectLiveness  bool
	CalledAnalyzeDocument bool
	
	// Mock responses
	AnalyzeFaceResult    *entities.FaceAnalysisResult
	CompareFacesResult   *entities.FaceComparisonResult
	DetectNSFWResult     *entities.ContentModerationResult
	DetectLivenessResult *entities.LivenessDetectionResult
	AnalyzeDocumentResult *entities.DocumentAnalysisResult
	
	// Mock errors
	AnalyzeFaceError    error
	CompareFacesError   error
	DetectNSFWError     error
	DetectLivenessError error
	AnalyzeDocumentError error
}

// NewMockAIService creates a new mock AI service
func NewMockAIService() *MockAIService {
	return &MockAIService{
		// Set default mock responses
		AnalyzeFaceResult: &entities.FaceAnalysisResult{
			HasFace:      true,
			Confidence:    0.95,
			BoundingBox:   &entities.BoundingBox{Width: 100, Height: 100, Left: 50, Top: 50},
			FaceID:       "mock-face-id",
			Emotions:     map[string]float64{"happy": 0.8, "neutral": 0.2},
			AgeRange:     &entities.AgeRange{Low: 25, High: 35},
			Gender:       "male",
			Quality:      0.9,
		},
		CompareFacesResult: &entities.FaceComparisonResult{
			Similarity:    0.92,
			Confidence:     0.88,
			IsMatch:       true,
			MatchedFaceID: "mock-face-id",
		},
		DetectNSFWResult: &entities.ContentModerationResult{
			IsNSFW:      false,
			Confidence:    0.1,
			Categories:    map[string]float64{"explicit": 0.05, "suggestive": 0.05},
		},
		DetectLivenessResult: &entities.LivenessDetectionResult{
			IsLive:       true,
			Confidence:    0.93,
			Score:         0.89,
			Challenges:    []string{"blink", "smile"},
		},
		AnalyzeDocumentResult: &entities.DocumentAnalysisResult{
			DocumentType:  "passport",
			Confidence:    0.87,
			IsValid:       true,
			ExtractedData: map[string]interface{}{
				"document_number": "P123456789",
				"name":           "John Doe",
				"nationality":    "US",
				"date_of_birth":  "1990-01-01",
				"expiry_date":    "2030-01-01",
			},
		},
	}
}

// AnalyzeFace analyzes a face in an image
func (m *MockAIService) AnalyzeFace(ctx context.Context, imageURL string) (*entities.FaceAnalysisResult, error) {
	m.CalledAnalyzeFace = true
	if m.AnalyzeFaceError != nil {
		return nil, m.AnalyzeFaceError
	}
	return m.AnalyzeFaceResult, nil
}

// CompareFaces compares two faces
func (m *MockAIService) CompareFaces(ctx context.Context, sourceImageURL, targetImageURL string) (*entities.FaceComparisonResult, error) {
	m.CalledCompareFaces = true
	if m.CompareFacesError != nil {
		return nil, m.CompareFacesError
	}
	return m.CompareFacesResult, nil
}

// DetectNSFW detects NSFW content in an image
func (m *MockAIService) DetectNSFW(ctx context.Context, imageURL string) (*entities.ContentModerationResult, error) {
	m.CalledDetectNSFW = true
	if m.DetectNSFWError != nil {
		return nil, m.DetectNSFWError
	}
	return m.DetectNSFWResult, nil
}

// DetectLiveness detects if a face is live (anti-spoofing)
func (m *MockAIService) DetectLiveness(ctx context.Context, imageURL string) (*entities.LivenessDetectionResult, error) {
	m.CalledDetectLiveness = true
	if m.DetectLivenessError != nil {
		return nil, m.DetectLivenessError
	}
	return m.DetectLivenessResult, nil
}

// AnalyzeDocument analyzes a document image
func (m *MockAIService) AnalyzeDocument(ctx context.Context, imageURL string) (*entities.DocumentAnalysisResult, error) {
	m.CalledAnalyzeDocument = true
	if m.AnalyzeDocumentError != nil {
		return nil, m.AnalyzeDocumentError
	}
	return m.AnalyzeDocumentResult, nil
}

// Reset resets the mock service state
func (m *MockAIService) Reset() {
	m.CalledAnalyzeFace = false
	m.CalledCompareFaces = false
	m.CalledDetectNSFW = false
	m.CalledDetectLiveness = false
	m.CalledAnalyzeDocument = false
	m.AnalyzeFaceError = nil
	m.CompareFacesError = nil
	m.DetectNSFWError = nil
	m.DetectLivenessError = nil
	m.AnalyzeDocumentError = nil
}

// SetAnalyzeFaceResult sets the mock result for AnalyzeFace
func (m *MockAIService) SetAnalyzeFaceResult(result *entities.FaceAnalysisResult, err error) {
	m.AnalyzeFaceResult = result
	m.AnalyzeFaceError = err
}

// SetCompareFacesResult sets the mock result for CompareFaces
func (m *MockAIService) SetCompareFacesResult(result *entities.FaceComparisonResult, err error) {
	m.CompareFacesResult = result
	m.CompareFacesError = err
}

// SetDetectNSFWResult sets the mock result for DetectNSFW
func (m *MockAIService) SetDetectNSFWResult(result *entities.ContentModerationResult, err error) {
	m.DetectNSFWResult = result
	m.DetectNSFWError = err
}

// SetDetectLivenessResult sets the mock result for DetectLiveness
func (m *MockAIService) SetDetectLivenessResult(result *entities.LivenessDetectionResult, err error) {
	m.DetectLivenessResult = result
	m.DetectLivenessError = err
}

// SetAnalyzeDocumentResult sets the mock result for AnalyzeDocument
func (m *MockAIService) SetAnalyzeDocumentResult(result *entities.DocumentAnalysisResult, err error) {
	m.AnalyzeDocumentResult = result
	m.AnalyzeDocumentError = err
}

// VerifyMethodCalls verifies that specific methods were called
func (m *MockAIService) VerifyMethodCalls(expectedAnalyzeFace, expectedCompareFaces, expectedDetectNSFW, expectedDetectLiveness, expectedAnalyzeDocument bool) error {
	if expectedAnalyzeFace && !m.CalledAnalyzeFace {
		return fmt.Errorf("expected AnalyzeFace to be called")
	}
	if !expectedAnalyzeFace && m.CalledAnalyzeFace {
		return fmt.Errorf("expected AnalyzeFace NOT to be called")
	}
	
	if expectedCompareFaces && !m.CalledCompareFaces {
		return fmt.Errorf("expected CompareFaces to be called")
	}
	if !expectedCompareFaces && m.CalledCompareFaces {
		return fmt.Errorf("expected CompareFaces NOT to be called")
	}
	
	if expectedDetectNSFW && !m.CalledDetectNSFW {
		return fmt.Errorf("expected DetectNSFW to be called")
	}
	if !expectedDetectNSFW && m.CalledDetectNSFW {
		return fmt.Errorf("expected DetectNSFW NOT to be called")
	}
	
	if expectedDetectLiveness && !m.CalledDetectLiveness {
		return fmt.Errorf("expected DetectLiveness to be called")
	}
	if !expectedDetectLiveness && m.CalledDetectLiveness {
		return fmt.Errorf("expected DetectLiveness NOT to be called")
	}
	
	if expectedAnalyzeDocument && !m.CalledAnalyzeDocument {
		return fmt.Errorf("expected AnalyzeDocument to be called")
	}
	if !expectedAnalyzeDocument && m.CalledAnalyzeDocument {
		return fmt.Errorf("expected AnalyzeDocument NOT to be called")
	}
	
	return nil
}