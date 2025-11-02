package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// DocumentService handles document processing and OCR
type DocumentService struct {
	aiService external.AIService
}

// NewDocumentService creates a new document service
func NewDocumentService(aiService external.AIService) *DocumentService {
	return &DocumentService{
		aiService: aiService,
	}
}

// ProcessDocument processes a document image and extracts information
func (ds *DocumentService) ProcessDocument(ctx context.Context, imageKey string) (*DocumentProcessingResult, error) {
	logger.Info("Processing document", "image_key", imageKey)

	// Use AI service to analyze the document
	analysis, err := ds.aiService.AnalyzeDocument(ctx, imageKey)
	if err != nil {
		logger.Error("Failed to analyze document with AI", err, "image_key", imageKey)
		return nil, fmt.Errorf("failed to analyze document: %w", err)
	}

	// Validate document type and extract fields
	result := &DocumentProcessingResult{
		DocumentType: analysis.DocumentType,
		IsValid:     analysis.IsValid,
		Confidence:   analysis.Confidence,
		Fields:       analysis.Fields,
		Details:      analysis.Details,
	}

	// Additional validation based on document type
	if analysis.IsValid {
		result.ValidationErrors = ds.validateDocumentFields(analysis.DocumentType, analysis.Fields)
		if len(result.ValidationErrors) > 0 {
			result.IsValid = false
			result.Details = fmt.Sprintf("Document validation failed: %s", strings.Join(result.ValidationErrors, ", "))
		}
	}

	logger.Info("Document processing completed", "image_key", imageKey, "type", result.DocumentType, "valid", result.IsValid, "confidence", result.Confidence)
	return result, nil
}

// ExtractDocumentData extracts structured data from document text
func (ds *DocumentService) ExtractDocumentData(documentType string, text string) (map[string]interface{}, error) {
	text = strings.ToLower(strings.TrimSpace(text))
	
	switch documentType {
	case "passport":
		return ds.extractPassportData(text), nil
	case "id_card":
		return ds.extractIDCardData(text), nil
	case "driver_license":
		return ds.extractDriverLicenseData(text), nil
	default:
		return map[string]interface{}{}, fmt.Errorf("unsupported document type: %s", documentType)
	}
}

// ValidateDocument validates document fields based on type
func (ds *DocumentService) ValidateDocument(documentType string, fields map[string]interface{}) []string {
	return ds.validateDocumentFields(documentType, fields)
}

// DocumentProcessingResult represents the result of document processing
type DocumentProcessingResult struct {
	DocumentType    string                 `json:"document_type"`
	IsValid        bool                    `json:"is_valid"`
	Confidence      float64                 `json:"confidence"`
	Fields          map[string]interface{}   `json:"fields"`
	ValidationErrors []string                 `json:"validation_errors,omitempty"`
	Details         string                  `json:"details"`
	ProcessedAt     time.Time               `json:"processed_at"`
}

// PassportData represents extracted passport information
type PassportData struct {
	PassportNumber    string `json:"passport_number"`
	FullName          string `json:"full_name"`
	DateOfBirth       string `json:"date_of_birth"`
	PlaceOfBirth      string `json:"place_of_birth"`
	Nationality       string `json:"nationality"`
	Gender            string `json:"gender"`
	IssueDate         string `json:"issue_date"`
	ExpiryDate       string `json:"expiry_date"`
	IssuingAuthority string `json:"issuing_authority"`
}

// IDCardData represents extracted ID card information
type IDCardData struct {
	IDNumber      string `json:"id_number"`
	FullName      string `json:"full_name"`
	DateOfBirth   string `json:"date_of_birth"`
	Address       string `json:"address"`
	Gender        string `json:"gender"`
	IssueDate     string `json:"issue_date"`
	ExpiryDate    string `json:"expiry_date"`
	IssuingState string `json:"issuing_state"`
}

// DriverLicenseData represents extracted driver license information
type DriverLicenseData struct {
	LicenseNumber    string `json:"license_number"`
	FullName         string `json:"full_name"`
	DateOfBirth      string `json:"date_of_birth"`
	Address          string `json:"address"`
	Gender           string `json:"gender"`
	IssueDate        string `json:"issue_date"`
	ExpiryDate       string `json:"expiry_date"`
	LicenseClass     string `json:"license_class"`
	Restrictions     string `json:"restrictions"`
	IssuingState    string `json:"issuing_state"`
}

// Helper methods for document extraction

func (ds *DocumentService) extractPassportData(text string) map[string]interface{} {
	data := make(map[string]interface{})
	
	// Extract passport number
	if passportNum := ds.extractPassportNumber(text); passportNum != "" {
		data["passport_number"] = passportNum
	}
	
	// Extract full name
	if name := ds.extractFullName(text); name != "" {
		data["full_name"] = name
	}
	
	// Extract date of birth
	if dob := ds.extractDateOfBirth(text); dob != "" {
		data["date_of_birth"] = dob
	}
	
	// Extract place of birth
	if pob := ds.extractPlaceOfBirth(text); pob != "" {
		data["place_of_birth"] = pob
	}
	
	// Extract nationality
	if nationality := ds.extractNationality(text); nationality != "" {
		data["nationality"] = nationality
	}
	
	// Extract gender
	if gender := ds.extractGender(text); gender != "" {
		data["gender"] = gender
	}
	
	// Extract dates
	if issueDate := ds.extractIssueDate(text); issueDate != "" {
		data["issue_date"] = issueDate
	}
	
	if expiryDate := ds.extractExpiryDate(text); expiryDate != "" {
		data["expiry_date"] = expiryDate
	}
	
	// Extract issuing authority
	if authority := ds.extractIssuingAuthority(text); authority != "" {
		data["issuing_authority"] = authority
	}
	
	return data
}

func (ds *DocumentService) extractIDCardData(text string) map[string]interface{} {
	data := make(map[string]interface{})
	
	// Extract ID number
	if idNum := ds.extractIDNumber(text); idNum != "" {
		data["id_number"] = idNum
	}
	
	// Extract full name
	if name := ds.extractFullName(text); name != "" {
		data["full_name"] = name
	}
	
	// Extract date of birth
	if dob := ds.extractDateOfBirth(text); dob != "" {
		data["date_of_birth"] = dob
	}
	
	// Extract address
	if address := ds.extractAddress(text); address != "" {
		data["address"] = address
	}
	
	// Extract gender
	if gender := ds.extractGender(text); gender != "" {
		data["gender"] = gender
	}
	
	// Extract dates
	if issueDate := ds.extractIssueDate(text); issueDate != "" {
		data["issue_date"] = issueDate
	}
	
	if expiryDate := ds.extractExpiryDate(text); expiryDate != "" {
		data["expiry_date"] = expiryDate
	}
	
	// Extract issuing state
	if state := ds.extractIssuingState(text); state != "" {
		data["issuing_state"] = state
	}
	
	return data
}

func (ds *DocumentService) extractDriverLicenseData(text string) map[string]interface{} {
	data := make(map[string]interface{})
	
	// Extract license number
	if licenseNum := ds.extractLicenseNumber(text); licenseNum != "" {
		data["license_number"] = licenseNum
	}
	
	// Extract full name
	if name := ds.extractFullName(text); name != "" {
		data["full_name"] = name
	}
	
	// Extract date of birth
	if dob := ds.extractDateOfBirth(text); dob != "" {
		data["date_of_birth"] = dob
	}
	
	// Extract address
	if address := ds.extractAddress(text); address != "" {
		data["address"] = address
	}
	
	// Extract gender
	if gender := ds.extractGender(text); gender != "" {
		data["gender"] = gender
	}
	
	// Extract dates
	if issueDate := ds.extractIssueDate(text); issueDate != "" {
		data["issue_date"] = issueDate
	}
	
	if expiryDate := ds.extractExpiryDate(text); expiryDate != "" {
		data["expiry_date"] = expiryDate
	}
	
	// Extract license class
	if class := ds.extractLicenseClass(text); class != "" {
		data["license_class"] = class
	}
	
	// Extract restrictions
	if restrictions := ds.extractRestrictions(text); restrictions != "" {
		data["restrictions"] = restrictions
	}
	
	// Extract issuing state
	if state := ds.extractIssuingState(text); state != "" {
		data["issuing_state"] = state
	}
	
	return data
}

// Validation methods

func (ds *DocumentService) validateDocumentFields(documentType string, fields map[string]interface{}) []string {
	var errors []string
	
	switch documentType {
	case "passport":
		errors = ds.validatePassportFields(fields)
	case "id_card":
		errors = ds.validateIDCardFields(fields)
	case "driver_license":
		errors = ds.validateDriverLicenseFields(fields)
	default:
		errors = append(errors, "unknown document type")
	}
	
	return errors
}

func (ds *DocumentService) validatePassportFields(fields map[string]interface{}) []string {
	var errors []string
	
	// Check required fields
	requiredFields := []string{"passport_number", "full_name", "date_of_birth"}
	for _, field := range requiredFields {
		if _, exists := fields[field]; !exists {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}
	
	// Validate passport number format
	if passportNum, exists := fields["passport_number"]; exists {
		if num, ok := passportNum.(string); ok && !ds.isValidPassportNumber(num) {
			errors = append(errors, "invalid passport number format")
		}
	}
	
	// Validate expiry date
	if expiryDate, exists := fields["expiry_date"]; exists {
		if date, ok := expiryDate.(string); ok && !ds.isValidExpiryDate(date) {
			errors = append(errors, "passport has expired or invalid expiry date")
		}
	}
	
	return errors
}

func (ds *DocumentService) validateIDCardFields(fields map[string]interface{}) []string {
	var errors []string
	
	// Check required fields
	requiredFields := []string{"id_number", "full_name", "date_of_birth"}
	for _, field := range requiredFields {
		if _, exists := fields[field]; !exists {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}
	
	// Validate ID number format
	if idNum, exists := fields["id_number"]; exists {
		if num, ok := idNum.(string); ok && !ds.isValidIDNumber(num) {
			errors = append(errors, "invalid ID number format")
		}
	}
	
	// Validate expiry date
	if expiryDate, exists := fields["expiry_date"]; exists {
		if date, ok := expiryDate.(string); ok && !ds.isValidExpiryDate(date) {
			errors = append(errors, "ID card has expired or invalid expiry date")
		}
	}
	
	return errors
}

func (ds *DocumentService) validateDriverLicenseFields(fields map[string]interface{}) []string {
	var errors []string
	
	// Check required fields
	requiredFields := []string{"license_number", "full_name", "date_of_birth"}
	for _, field := range requiredFields {
		if _, exists := fields[field]; !exists {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}
	
	// Validate license number format
	if licenseNum, exists := fields["license_number"]; exists {
		if num, ok := licenseNum.(string); ok && !ds.isValidLicenseNumber(num) {
			errors = append(errors, "invalid license number format")
		}
	}
	
	// Validate expiry date
	if expiryDate, exists := fields["expiry_date"]; exists {
		if date, ok := expiryDate.(string); ok && !ds.isValidExpiryDate(date) {
			errors = append(errors, "driver license has expired or invalid expiry date")
		}
	}
	
	return errors
}

// Text extraction methods using regex patterns

func (ds *DocumentService) extractPassportNumber(text string) string {
	// Passport number patterns (various formats)
	patterns := []string{
		`[A-Z]{1,2}\d{7,8}`,           // Standard format: 2 letters + 7-8 digits
		`\d{9}`,                           // Numeric format: 9 digits
		`[A-Z]\d{8}`,                    // Alternative: 1 letter + 8 digits
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindString(text); match != "" {
			return match
		}
	}
	return ""
}

func (ds *DocumentService) extractIDNumber(text string) string {
	// ID number patterns
	patterns := []string{
		`\b\d{8,12}\b`,                 // 8-12 digits
		`\b[A-Z]{2}\d{6,7}\b`,          // 2 letters + 6-7 digits
		`\b\d{3}-\d{2}-\d{4}\b`,        // Format: 123-45-6789
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindString(text); match != "" {
			return match
		}
	}
	return ""
}

func (ds *DocumentService) extractLicenseNumber(text string) string {
	// License number patterns
	patterns := []string{
		`\b[A-Z]{1,2}\d{6,8}\b`,          // State code + digits
		`\b\d{8,12}\b`,                    // Pure digits
		`\b[A-Z]{2}\d{6}\b`,              // Specific format
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindString(text); match != "" {
			return match
		}
	}
	return ""
}

func (ds *DocumentService) extractFullName(text string) string {
	// Name extraction patterns
	patterns := []string{
		`[A-Z][a-z]+ [A-Z][a-z]+`,           // First Last
		`[A-Z][a-z]+ [A-Z][a-z]+ [A-Z][a-z]+`, // First Middle Last
		`Name[:\s]*([A-Z][a-z]+ [A-Z][a-z]+)`, // Name: field
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractDateOfBirth(text string) string {
	// Date of birth patterns
	patterns := []string{
		`DOB[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
		`Date of Birth[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
		`Born[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
		`(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractPlaceOfBirth(text string) string {
	// Place of birth patterns
	patterns := []string{
		`Place of Birth[:\s]*([A-Za-z\s,]+)`,
		`POB[:\s]*([A-Za-z\s,]+)`,
		`Born in[:\s]*([A-Za-z\s,]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractNationality(text string) string {
	// Nationality patterns
	patterns := []string{
		`Nationality[:\s]*([A-Za-z\s,]+)`,
		`Citizen of[:\s]*([A-Za-z\s,]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractGender(text string) string {
	// Gender patterns
	patterns := []string{
		`Gender[:\s]*([Mm]ale|[Ff]emale)`,
		`Sex[:\s]*([Mm]ale|[Ff]emale)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			gender := strings.ToLower(match[1])
			if strings.HasPrefix(gender, "m") {
				return "male"
			}
			return "female"
		}
	}
	return ""
}

func (ds *DocumentService) extractIssueDate(text string) string {
	// Issue date patterns
	patterns := []string{
		`Issue Date[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
		`Issued[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
		`Date of Issue[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractExpiryDate(text string) string {
	// Expiry date patterns
	patterns := []string{
		`Expiry Date[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
		`Expires[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
		`Expiration[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
		`Exp[:\s]*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractIssuingAuthority(text string) string {
	// Authority patterns
	patterns := []string{
		`Authority[:\s]*([A-Za-z\s,]+)`,
		`Issuing Authority[:\s]*([A-Za-z\s,]+)`,
		`Issued by[:\s]*([A-Za-z\s,]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractAddress(text string) string {
	// Address patterns
	patterns := []string{
		`Address[:\s]*([A-Za-z0-9\s,.-]+)`,
		`Residence[:\s]*([A-Za-z0-9\s,.-]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractIssuingState(text string) string {
	// State patterns
	patterns := []string{
		`State[:\s]*([A-Za-z\s,]+)`,
		`Issuing State[:\s]*([A-Za-z\s,]+)`,
		`([A-Z]{2})`, // Two-letter state code
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractLicenseClass(text string) string {
	// License class patterns
	patterns := []string{
		`Class[:\s]*([A-Z0-9]+)`,
		`License Class[:\s]*([A-Z0-9]+)`,
		`Type[:\s]*([A-Z0-9]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ds *DocumentService) extractRestrictions(text string) string {
	// Restriction patterns
	patterns := []string{
		`Restrictions[:\s]*([A-Za-z0-9\s,]+)`,
		`Endorsements[:\s]*([A-Za-z0-9\s,]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

// Validation helper methods

func (ds *DocumentService) isValidPassportNumber(number string) bool {
	// Basic validation for passport number format
	if len(number) < 6 || len(number) > 12 {
		return false
	}
	
	// Check if alphanumeric
	for _, char := range number {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	
	return true
}

func (ds *DocumentService) isValidIDNumber(number string) bool {
	// Basic validation for ID number
	if len(number) < 6 || len(number) > 15 {
		return false
	}
	
	// Check if numeric
	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}

func (ds *DocumentService) isValidLicenseNumber(number string) bool {
	// Basic validation for license number
	if len(number) < 6 || len(number) > 15 {
		return false
	}
	
	// Check if alphanumeric
	for _, char := range number {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	
	return true
}

func (ds *DocumentService) isValidExpiryDate(dateStr string) bool {
	// Parse date in various formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
	}
	
	var expiryDate time.Time
	var err error
	
	for _, format := range formats {
		expiryDate, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		return false
	}
	
	// Check if date is in the future or too far in the past
	now := time.Now()
	minDate := now.AddDate(-100, 0, 0) // 100 years ago
	maxDate := now.AddDate(10, 0, 0)  // 10 years in future
	
	return expiryDate.After(minDate) && expiryDate.Before(maxDate)
}