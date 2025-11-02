package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// AssertionHelper provides custom assertion methods for testing
type AssertionHelper struct {
	t *testing.T
}

// NewAssertionHelper creates a new assertion helper
func NewAssertionHelper(t *testing.T) *AssertionHelper {
	return &AssertionHelper{t: t}
}

// AssertUserEqual asserts that two users are equal, ignoring timestamps
func (ah *AssertionHelper) AssertUserEqual(expected, actual *entities.User) {
	assert.Equal(ah.t, expected.ID, actual.ID, "User ID should match")
	assert.Equal(ah.t, expected.Email, actual.Email, "User email should match")
	assert.Equal(ah.t, expected.FirstName, actual.FirstName, "User first name should match")
	assert.Equal(ah.t, expected.LastName, actual.LastName, "User last name should match")
	assert.Equal(ah.t, expected.IsVerified, actual.IsVerified, "User verification status should match")
	assert.Equal(ah.t, expected.IsBanned, actual.IsBanned, "User ban status should match")
	// Don't compare timestamps as they may differ
}

// AssertProfileEqual asserts that two profiles are equal, ignoring timestamps
func (ah *AssertionHelper) AssertProfileEqual(expected, actual *entities.Profile) {
	assert.Equal(ah.t, expected.ID, actual.ID, "Profile ID should match")
	assert.Equal(ah.t, expected.UserID, actual.UserID, "Profile user ID should match")
	assert.Equal(ah.t, expected.Bio, actual.Bio, "Profile bio should match")
	assert.Equal(ah.t, expected.Age, actual.Age, "Profile age should match")
	assert.Equal(ah.t, expected.Gender, actual.Gender, "Profile gender should match")
	assert.Equal(ah.t, expected.Location, actual.Location, "Profile location should match")
	assert.ElementsMatch(ah.t, expected.InterestedIn, actual.InterestedIn, "Profile interested in should match")
	// Don't compare timestamps as they may differ
}

// AssertPhotoEqual asserts that two photos are equal, ignoring timestamps
func (ah *AssertionHelper) AssertPhotoEqual(expected, actual *entities.Photo) {
	assert.Equal(ah.t, expected.ID, actual.ID, "Photo ID should match")
	assert.Equal(ah.t, expected.UserID, actual.UserID, "Photo user ID should match")
	assert.Equal(ah.t, expected.URL, actual.URL, "Photo URL should match")
	assert.Equal(ah.t, expected.IsPrimary, actual.IsPrimary, "Photo primary status should match")
	assert.Equal(ah.t, expected.IsApproved, actual.IsApproved, "Photo approval status should match")
	// Don't compare timestamps as they may differ
}

// AssertMatchEqual asserts that two matches are equal, ignoring timestamps
func (ah *AssertionHelper) AssertMatchEqual(expected, actual *entities.Match) {
	assert.Equal(ah.t, expected.ID, actual.ID, "Match ID should match")
	assert.Equal(ah.t, expected.UserID, actual.UserID, "Match user ID should match")
	assert.Equal(ah.t, expected.MatchedUserID, actual.MatchedUserID, "Match matched user ID should match")
	assert.Equal(ah.t, expected.Status, actual.Status, "Match status should match")
	// Don't compare timestamps as they may differ
}

// AssertMessageEqual asserts that two messages are equal, ignoring timestamps
func (ah *AssertionHelper) AssertMessageEqual(expected, actual *entities.Message) {
	assert.Equal(ah.t, expected.ID, actual.ID, "Message ID should match")
	assert.Equal(ah.t, expected.SenderID, actual.SenderID, "Message sender ID should match")
	assert.Equal(ah.t, expected.ReceiverID, actual.ReceiverID, "Message receiver ID should match")
	assert.Equal(ah.t, expected.Content, actual.Content, "Message content should match")
	assert.Equal(ah.t, expected.IsRead, actual.IsRead, "Message read status should match")
	// Don't compare timestamps as they may differ
}

// AssertSubscriptionEqual asserts that two subscriptions are equal, ignoring timestamps
func (ah *AssertionHelper) AssertSubscriptionEqual(expected, actual *entities.Subscription) {
	assert.Equal(ah.t, expected.ID, actual.ID, "Subscription ID should match")
	assert.Equal(ah.t, expected.UserID, actual.UserID, "Subscription user ID should match")
	assert.Equal(ah.t, expected.PlanType, actual.PlanType, "Subscription plan type should match")
	assert.Equal(ah.t, expected.Status, actual.Status, "Subscription status should match")
	assert.WithinDuration(ah.t, expected.StartDate, actual.StartDate, time.Second, "Subscription start date should match")
	assert.WithinDuration(ah.t, expected.EndDate, actual.EndDate, time.Second, "Subscription end date should match")
	// Don't compare timestamps as they may differ
}

// AssertVerificationEqual asserts that two verifications are equal, ignoring timestamps
func (ah *AssertionHelper) AssertVerificationEqual(expected, actual *entities.Verification) {
	assert.Equal(ah.t, expected.ID, actual.ID, "Verification ID should match")
	assert.Equal(ah.t, expected.UserID, actual.UserID, "Verification user ID should match")
	assert.Equal(ah.t, expected.Type, actual.Type, "Verification type should match")
	assert.Equal(ah.t, expected.Status, actual.Status, "Verification status should match")
	// Don't compare timestamps as they may differ
}

// AssertReportEqual asserts that two reports are equal, ignoring timestamps
func (ah *AssertionHelper) AssertReportEqual(expected, actual *entities.Report) {
	assert.Equal(ah.t, expected.ID, actual.ID, "Report ID should match")
	assert.Equal(ah.t, expected.ReporterID, actual.ReporterID, "Report reporter ID should match")
	assert.Equal(ah.t, expected.ReportedUserID, actual.ReportedUserID, "Report reported user ID should match")
	assert.Equal(ah.t, expected.Reason, actual.Reason, "Report reason should match")
	assert.Equal(ah.t, expected.Description, actual.Description, "Report description should match")
	assert.Equal(ah.t, expected.Status, actual.Status, "Report status should match")
	// Don't compare timestamps as they may differ
}

// AssertHTTPResponse asserts HTTP response properties
func (ah *AssertionHelper) AssertHTTPResponse(resp *http.Response, expectedStatusCode int, expectedContentType string) {
	assert.Equal(ah.t, expectedStatusCode, resp.StatusCode, "HTTP status code should match")
	if expectedContentType != "" {
		assert.Equal(ah.t, expectedContentType, resp.Header.Get("Content-Type"), "Content-Type should match")
	}
}

// AssertJSONResponse asserts that response contains valid JSON
func (ah *AssertionHelper) AssertJSONResponse(resp *http.Response) map[string]interface{} {
	require.Equal(ah.t, "application/json", resp.Header.Get("Content-Type"), "Response should be JSON")
	
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(ah.t, err, "Response should be valid JSON")
	
	return result
}

// AssertJSONError asserts that response contains a JSON error
func (ah *AssertionHelper) AssertJSONError(resp *http.Response, expectedMessage string) {
	data := ah.AssertJSONResponse(resp)
	
	assert.Contains(ah.t, data, "error", "Response should contain error field")
	if expectedMessage != "" {
		assert.Equal(ah.t, expectedMessage, data["error"], "Error message should match")
	}
}

// AssertJSONSuccess asserts that response contains a JSON success message
func (ah *AssertionHelper) AssertJSONSuccess(resp *http.Response, expectedMessage string) {
	data := ah.AssertJSONResponse(resp)
	
	assert.Contains(ah.t, data, "message", "Response should contain message field")
	if expectedMessage != "" {
		assert.Equal(ah.t, expectedMessage, data["message"], "Success message should match")
	}
}

// AssertJSONData asserts that response contains JSON data
func (ah *AssertionHelper) AssertJSONData(resp *http.Response, expectedData interface{}) {
	data := ah.AssertJSONResponse(resp)
	
	assert.Contains(ah.t, data, "data", "Response should contain data field")
	
	if expectedData != nil {
		actualData := data["data"]
		expectedJSON, _ := json.Marshal(expectedData)
		actualJSON, _ := json.Marshal(actualData)
		
		assert.JSONEq(ah.t, string(expectedJSON), string(actualJSON), "Response data should match expected data")
	}
}

// AssertPagination asserts pagination properties in response
func (ah *AssertionHelper) AssertPagination(resp *http.Response, expectedPage, expectedLimit, expectedTotal int) {
	data := ah.AssertJSONResponse(resp)
	
	assert.Contains(ah.t, data, "pagination", "Response should contain pagination")
	pagination := data["pagination"].(map[string]interface{})
	
	assert.Equal(ah.t, expectedPage, int(pagination["page"].(float64)), "Page should match")
	assert.Equal(ah.t, expectedLimit, int(pagination["limit"].(float64)), "Limit should match")
	assert.Equal(ah.t, expectedTotal, int(pagination["total"].(float64)), "Total should match")
}

// AssertUUID asserts that a string is a valid UUID
func (ah *AssertionHelper) AssertUUID(id string) {
	_, err := uuid.Parse(id)
	assert.NoError(ah.t, err, "ID should be a valid UUID")
}

// AssertEmail asserts that a string is a valid email format
func (ah *AssertionHelper) AssertEmail(email string) {
	assert.Contains(ah.t, email, "@", "Email should contain @ symbol")
	assert.Contains(ah.t, email, ".", "Email should contain domain")
	assert.Greater(ah.t, len(email), 5, "Email should have reasonable length")
}

// AssertPasswordHash asserts that a string is a valid password hash
func (ah *AssertionHelper) AssertPasswordHash(hash string) {
	assert.Greater(ah.t, len(hash), 20, "Password hash should be sufficiently long")
	assert.NotContains(ah.t, hash, "password", "Password hash should not contain plain text password")
	assert.NotContains(ah.t, hash, "123", "Password hash should not contain common patterns")
}

// AssertTimestamp asserts that a timestamp is within expected range
func (ah *AssertionHelper) AssertTimestamp(timestamp time.Time, expectedRange time.Duration) {
	now := time.Now()
	assert.WithinDuration(ah.t, now, timestamp, expectedRange, "Timestamp should be within expected range")
}

// AssertAge asserts that age is within reasonable range
func (ah *AssertionHelper) AssertAge(age int) {
	assert.GreaterOrEqual(ah.t, age, 18, "Age should be at least 18")
	assert.LessOrEqual(ah.t, age, 100, "Age should be at most 100")
}

// AssertGender asserts that gender is valid
func (ah *AssertionHelper) AssertGender(gender string) {
	validGenders := []string{"male", "female", "other"}
	assert.Contains(ah.t, validGenders, gender, "Gender should be valid")
}

// AssertRelationshipStatus asserts that relationship status is valid
func (ah *AssertionHelper) AssertRelationshipStatus(status string) {
	validStatuses := []string{"single", "in_a_relationship", "married", "divorced", "widowed"}
	assert.Contains(ah.t, validStatuses, status, "Relationship status should be valid")
}

// AssertVerificationStatus asserts that verification status is valid
func (ah *AssertionHelper) AssertVerificationStatus(status string) {
	validStatuses := []string{"pending", "approved", "rejected"}
	assert.Contains(ah.t, validStatuses, status, "Verification status should be valid")
}

// AssertMatchStatus asserts that match status is valid
func (ah *AssertionHelper) AssertMatchStatus(status string) {
	validStatuses := []string{"pending", "matched", "rejected"}
	assert.Contains(ah.t, validStatuses, status, "Match status should be valid")
}

// AssertSubscriptionStatus asserts that subscription status is valid
func (ah *AssertionHelper) AssertSubscriptionStatus(status string) {
	validStatuses := []string{"active", "cancelled", "expired", "trial"}
	assert.Contains(ah.t, validStatuses, status, "Subscription status should be valid")
}

// AssertReportStatus asserts that report status is valid
func (ah *AssertionHelper) AssertReportStatus(status string) {
	validStatuses := []string{"pending", "reviewing", "resolved", "dismissed"}
	assert.Contains(ah.t, validStatuses, status, "Report status should be valid")
}

// AssertPlanType asserts that plan type is valid
func (ah *AssertionHelper) AssertPlanType(planType string) {
	validPlans := []string{"free", "premium", "premium_plus"}
	assert.Contains(ah.t, validPlans, planType, "Plan type should be valid")
}

// AssertVerificationType asserts that verification type is valid
func (ah *AssertionHelper) AssertVerificationType(verificationType string) {
	validTypes := []string{"document", "selfie", "video"}
	assert.Contains(ah.t, validTypes, verificationType, "Verification type should be valid")
}

// AssertReportReason asserts that report reason is valid
func (ah *AssertionHelper) AssertReportReason(reason string) {
	validReasons := []string{"inappropriate_content", "fake_profile", "spam", "harassment", "other"}
	assert.Contains(ah.t, validReasons, reason, "Report reason should be valid")
}

// AssertURL asserts that a string is a valid URL
func (ah *AssertionHelper) AssertURL(url string) {
	assert.Contains(ah.t, url, "http", "URL should contain http protocol")
	assert.Greater(ah.t, len(url), 10, "URL should have reasonable length")
}

// AssertImageURL asserts that a string is a valid image URL
func (ah *AssertionHelper) AssertImageURL(url string) {
	ah.AssertURL(url)
	
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	hasImageExtension := false
	for _, ext := range imageExtensions {
		if len(url) > len(ext) && url[len(url)-len(ext):] == ext {
			hasImageExtension = true
			break
		}
	}
	assert.True(ah.t, hasImageExtension, "URL should point to an image file")
}

// AssertBio asserts that bio is valid
func (ah *AssertionHelper) AssertBio(bio string) {
	assert.LessOrEqual(ah.t, len(bio), 500, "Bio should not exceed 500 characters")
	assert.GreaterOrEqual(ah.t, len(bio), 0, "Bio length should be non-negative")
}

// AssertLocation asserts that location is valid
func (ah *AssertionHelper) AssertLocation(location string) {
	assert.Greater(ah.t, len(location), 0, "Location should not be empty")
	assert.LessOrEqual(ah.t, len(location), 100, "Location should not exceed 100 characters")
}

// AssertMessageContent asserts that message content is valid
func (ah *AssertionHelper) AssertMessageContent(content string) {
	assert.Greater(ah.t, len(content), 0, "Message content should not be empty")
	assert.LessOrEqual(ah.t, len(content), 1000, "Message content should not exceed 1000 characters")
}

// AssertSliceNotEmpty asserts that a slice is not empty
func (ah *AssertionHelper) AssertSliceNotEmpty(slice interface{}) {
	value := reflect.ValueOf(slice)
	assert.Greater(ah.t, value.Len(), 0, "Slice should not be empty")
}

// AssertSliceLength asserts that a slice has expected length
func (ah *AssertionHelper) AssertSliceLength(slice interface{}, expectedLength int) {
	value := reflect.ValueOf(slice)
	assert.Equal(ah.t, expectedLength, value.Len(), "Slice should have expected length")
}

// AssertMapContainsKey asserts that a map contains a specific key
func (ah *AssertionHelper) AssertMapContainsKey(m map[string]interface{}, key string) {
	_, exists := m[key]
	assert.True(ah.t, exists, "Map should contain key: %s", key)
}

// AssertMapNotContainsKey asserts that a map does not contain a specific key
func (ah *AssertionHelper) AssertMapNotContainsKey(m map[string]interface{}, key string) {
	_, exists := m[key]
	assert.False(ah.t, exists, "Map should not contain key: %s", key)
}

// AssertMapValue asserts that a map has a specific value for a key
func (ah *AssertionHelper) AssertMapValue(m map[string]interface{}, key, expectedValue string) {
	ah.AssertMapContainsKey(m, key)
	assert.Equal(ah.t, expectedValue, m[key], "Map value should match for key: %s", key)
}

// AssertContextCanceled asserts that context is canceled
func (ah *AssertionHelper) AssertContextCanceled(ctx context.Context) {
	select {
	case <-ctx.Done():
		assert.Equal(ah.t, context.Canceled, ctx.Err(), "Context should be canceled")
	default:
		assert.Fail(ah.t, "Context should be canceled")
	}
}

// AssertContextDeadline asserts that context has deadline exceeded
func (ah *AssertionHelper) AssertContextDeadline(ctx context.Context) {
	select {
	case <-ctx.Done():
		assert.Equal(ah.t, context.DeadlineExceeded, ctx.Err(), "Context should have deadline exceeded")
	default:
		assert.Fail(ah.t, "Context should have deadline exceeded")
	}
}

// AssertNoError asserts that error is nil
func (ah *AssertionHelper) AssertNoError(err error) {
	assert.NoError(ah.t, err, "Error should be nil")
}

// AssertError asserts that error is not nil and contains expected message
func (ah *AssertionHelper) AssertError(err error, expectedMessage string) {
	assert.Error(ah.t, err, "Error should not be nil")
	if expectedMessage != "" {
		assert.Contains(ah.t, err.Error(), expectedMessage, "Error message should contain expected text")
	}
}

// AssertErrorType asserts that error is of expected type
func (ah *AssertionHelper) AssertErrorType(err error, expectedType error) {
	assert.Error(ah.t, err, "Error should not be nil")
	assert.IsType(ah.t, expectedType, err, "Error should be of expected type")
}

// AssertEventually asserts that a condition eventually becomes true within timeout
func (ah *AssertionHelper) AssertEventually(condition func() bool, timeout time.Duration, message string) {
	start := time.Now()
	for time.Since(start) < timeout {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Fail(ah.t, fmt.Sprintf("Condition did not become true within %v: %s", timeout, message))
}

// AssertEventuallyEqual asserts that a value eventually equals expected value within timeout
func (ah *AssertionHelper) AssertEventuallyEqual(getValue func() interface{}, expected interface{}, timeout time.Duration) {
	ah.AssertEventually(func() bool {
		return reflect.DeepEqual(getValue(), expected)
	}, timeout, fmt.Sprintf("Value did not equal expected %v", expected))
}

// AssertBufferContains asserts that buffer contains expected string
func (ah *AssertionHelper) AssertBufferContains(buffer *bytes.Buffer, expected string) {
	assert.Contains(ah.t, buffer.String(), expected, "Buffer should contain expected string")
}

// AssertBufferEmpty asserts that buffer is empty
func (ah *AssertionHelper) AssertBufferEmpty(buffer *bytes.Buffer) {
	assert.Empty(ah.t, buffer.String(), "Buffer should be empty")
}

// AssertBufferNotEmpty asserts that buffer is not empty
func (ah *AssertionHelper) AssertBufferNotEmpty(buffer *bytes.Buffer) {
	assert.NotEmpty(ah.t, buffer.String(), "Buffer should not be empty")
}

// AssertJSONEqual asserts that two JSON strings are equal
func (ah *AssertionHelper) AssertJSONEqual(expected, actual string) {
	assert.JSONEq(ah.t, expected, actual, "JSON should be equal")
}

// AssertJSONContains asserts that JSON contains expected key/value
func (ah *AssertionHelper) AssertJSONContains(jsonStr, key, expectedValue string) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	require.NoError(ah.t, err, "JSON should be valid")
	
	ah.AssertMapValue(data, key, expectedValue)
}

// AssertTimeInRange asserts that time is within range
func (ah *AssertionHelper) AssertTimeInRange(t, start, end time.Time) {
	assert.True(ah.t, t.After(start) || t.Equal(start), "Time should be after or equal to start")
	assert.True(ah.t, t.Before(end) || t.Equal(end), "Time should be before or equal to end")
}

// AssertDurationInRange asserts that duration is within range
func (ah *AssertionHelper) AssertDurationInRange(d, min, max time.Duration) {
	assert.GreaterOrEqual(ah.t, d, min, "Duration should be at least minimum")
	assert.LessOrEqual(ah.t, d, max, "Duration should be at most maximum")
}

// AssertPositive asserts that value is positive
func (ah *AssertionHelper) AssertPositive(value int) {
	assert.Greater(ah.t, value, 0, "Value should be positive")
}

// AssertNegative asserts that value is negative
func (ah *AssertionHelper) AssertNegative(value int) {
	assert.Less(ah.t, value, 0, "Value should be negative")
}

// AssertNonNegative asserts that value is non-negative
func (ah *AssertionHelper) AssertNonNegative(value int) {
	assert.GreaterOrEqual(ah.t, value, 0, "Value should be non-negative")
}

// AssertNonPositive asserts that value is non-positive
func (ah *AssertionHelper) AssertNonPositive(value int) {
	assert.LessOrEqual(ah.t, value, 0, "Value should be non-positive")
}

// AssertZero asserts that value is zero
func (ah *AssertionHelper) AssertZero(value int) {
	assert.Equal(ah.t, 0, value, "Value should be zero")
}

// AssertNonZero asserts that value is not zero
func (ah *AssertionHelper) AssertNonZero(value int) {
	assert.NotEqual(ah.t, 0, value, "Value should not be zero")
}

// AssertEmpty asserts that value is empty
func (ah *AssertionHelper) AssertEmpty(value interface{}) {
	assert.Empty(ah.t, value, "Value should be empty")
}

// AssertNotEmpty asserts that value is not empty
func (ah *AssertionHelper) AssertNotEmpty(value interface{}) {
	assert.NotEmpty(ah.t, value, "Value should not be empty")
}

// AssertNil asserts that value is nil
func (ah *AssertionHelper) AssertNil(value interface{}) {
	assert.Nil(ah.t, value, "Value should be nil")
}

// AssertNotNil asserts that value is not nil
func (ah *AssertionHelper) AssertNotNil(value interface{}) {
	assert.NotNil(ah.t, value, "Value should not be nil")
}

// AssertTrue asserts that value is true
func (ah *AssertionHelper) AssertTrue(value bool) {
	assert.True(ah.t, value, "Value should be true")
}

// AssertFalse asserts that value is false
func (ah *AssertionHelper) AssertFalse(value bool) {
	assert.False(ah.t, value, "Value should be false")
}

// AssertEqual asserts that two values are equal
func (ah *AssertionHelper) AssertEqual(expected, actual interface{}) {
	assert.Equal(ah.t, expected, actual, "Values should be equal")
}

// AssertNotEqual asserts that two values are not equal
func (ah *AssertionHelper) AssertNotEqual(expected, actual interface{}) {
	assert.NotEqual(ah.t, expected, actual, "Values should not be equal")
}

// AssertSame asserts that two values are the same (reference equality)
func (ah *AssertionHelper) AssertSame(expected, actual interface{}) {
	assert.Same(ah.t, expected, actual, "Values should be the same")
}

// AssertNotSame asserts that two values are not the same
func (ah *AssertionHelper) AssertNotSame(expected, actual interface{}) {
	assert.NotSame(ah.t, expected, actual, "Values should not be the same")
}

// AssertType asserts that value is of expected type
func (ah *AssertionHelper) AssertType(value interface{}, expectedType interface{}) {
	assert.IsType(ah.t, expectedType, value, "Value should be of expected type")
}

// AssertImplements asserts that value implements expected interface
func (ah *AssertionHelper) AssertImplements(value interface{}, interfacePtr interface{}) {
	assert.Implements(ah.t, interfacePtr, value, "Value should implement interface")
}

// AssertPanics asserts that function panics
func (ah *AssertionHelper) AssertPanics(f func()) {
	assert.Panics(ah.t, f, "Function should panic")
}

// AssertNotPanics asserts that function does not panic
func (ah *AssertionHelper) AssertNotPanics(f func()) {
	assert.NotPanics(ah.t, f, "Function should not panic")
}

// AssertPanicsWithValue asserts that function panics with specific value
func (ah *AssertionHelper) AssertPanicsWithValue(expected interface{}, f func()) {
	assert.PanicsWithValue(ah.t, expected, f, "Function should panic with expected value")
}

// AssertPanicsWithError asserts that function panics with specific error
func (ah *AssertionHelper) AssertPanicsWithError(expectedError string, f func()) {
	assert.PanicsWithError(ah.t, expectedError, f, "Function should panic with expected error")
}