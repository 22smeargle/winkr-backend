package assertions

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// AssertUserEqual asserts that two users are equal, ignoring certain fields
func AssertUserEqual(t *testing.T, expected, actual *entities.User, ignoreFields ...string) {
	if expected == nil && actual == nil {
		return
	}
	require.NotNil(t, expected, "Expected user should not be nil")
	require.NotNil(t, actual, "Actual user should not be nil")

	// Create copies to avoid modifying originals
	expectedCopy := *expected
	actualCopy := *actual

	// Ignore specified fields
	for _, field := range ignoreFields {
		switch field {
		case "CreatedAt":
			expectedCopy.CreatedAt = time.Time{}
			actualCopy.CreatedAt = time.Time{}
		case "UpdatedAt":
			expectedCopy.UpdatedAt = time.Time{}
			actualCopy.UpdatedAt = time.Time{}
		case "LastActiveAt":
			expectedCopy.LastActiveAt = time.Time{}
			actualCopy.LastActiveAt = time.Time{}
		case "Password":
			expectedCopy.Password = ""
			actualCopy.Password = ""
		}
	}

	assert.Equal(t, expectedCopy, actualCopy, "Users should be equal")
}

// AssertPhotoEqual asserts that two photos are equal, ignoring certain fields
func AssertPhotoEqual(t *testing.T, expected, actual *entities.Photo, ignoreFields ...string) {
	if expected == nil && actual == nil {
		return
	}
	require.NotNil(t, expected, "Expected photo should not be nil")
	require.NotNil(t, actual, "Actual photo should not be nil")

	// Create copies to avoid modifying originals
	expectedCopy := *expected
	actualCopy := *actual

	// Ignore specified fields
	for _, field := range ignoreFields {
		switch field {
		case "CreatedAt":
			expectedCopy.CreatedAt = time.Time{}
			actualCopy.CreatedAt = time.Time{}
		case "UpdatedAt":
			expectedCopy.UpdatedAt = time.Time{}
			actualCopy.UpdatedAt = time.Time{}
		}
	}

	assert.Equal(t, expectedCopy, actualCopy, "Photos should be equal")
}

// AssertMatchEqual asserts that two matches are equal, ignoring certain fields
func AssertMatchEqual(t *testing.T, expected, actual *entities.Match, ignoreFields ...string) {
	if expected == nil && actual == nil {
		return
	}
	require.NotNil(t, expected, "Expected match should not be nil")
	require.NotNil(t, actual, "Actual match should not be nil")

	// Create copies to avoid modifying originals
	expectedCopy := *expected
	actualCopy := *actual

	// Ignore specified fields
	for _, field := range ignoreFields {
		switch field {
		case "CreatedAt":
			expectedCopy.CreatedAt = time.Time{}
			actualCopy.CreatedAt = time.Time{}
		case "UpdatedAt":
			expectedCopy.UpdatedAt = time.Time{}
			actualCopy.UpdatedAt = time.Time{}
		case "MatchedAt":
			expectedCopy.MatchedAt = time.Time{}
			actualCopy.MatchedAt = time.Time{}
		}
	}

	assert.Equal(t, expectedCopy, actualCopy, "Matches should be equal")
}

// AssertMessageEqual asserts that two messages are equal, ignoring certain fields
func AssertMessageEqual(t *testing.T, expected, actual *entities.Message, ignoreFields ...string) {
	if expected == nil && actual == nil {
		return
	}
	require.NotNil(t, expected, "Expected message should not be nil")
	require.NotNil(t, actual, "Actual message should not be nil")

	// Create copies to avoid modifying originals
	expectedCopy := *expected
	actualCopy := *actual

	// Ignore specified fields
	for _, field := range ignoreFields {
		switch field {
		case "CreatedAt":
			expectedCopy.CreatedAt = time.Time{}
			actualCopy.CreatedAt = time.Time{}
		case "UpdatedAt":
			expectedCopy.UpdatedAt = time.Time{}
			actualCopy.UpdatedAt = time.Time{}
		}
	}

	assert.Equal(t, expectedCopy, actualCopy, "Messages should be equal")
}

// AssertUUIDEqual asserts that two UUIDs are equal
func AssertUUIDEqual(t *testing.T, expected, actual uuid.UUID) {
	assert.Equal(t, expected, actual, "UUIDs should be equal")
}

// AssertUUIDNotNil asserts that a UUID is not nil
func AssertUUIDNotNil(t *testing.T, id uuid.UUID) {
	assert.NotEqual(t, uuid.Nil, id, "UUID should not be nil")
}

// AssertTimeClose asserts that two times are close to each other within a tolerance
func AssertTimeClose(t *testing.T, expected, actual time.Time, tolerance time.Duration) {
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	assert.LessOrEqual(t, diff, tolerance, "Times should be close within %v", tolerance)
}

// AssertTimeRecent asserts that a time is recent within a tolerance
func AssertTimeRecent(t *testing.T, actual time.Time, tolerance time.Duration) {
	now := time.Now()
	diff := now.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	assert.LessOrEqual(t, diff, tolerance, "Time should be recent within %v", tolerance)
}

// AssertSliceEqual asserts that two slices are equal, ignoring order
func AssertSliceEqual[T comparable](t *testing.T, expected, actual []T) {
	require.Equal(t, len(expected), len(actual), "Slices should have the same length")
	
	expectedMap := make(map[T]struct{})
	for _, item := range expected {
		expectedMap[item] = struct{}{}
	}
	
	for _, item := range actual {
		_, exists := expectedMap[item]
		assert.True(t, exists, "Item %v should exist in expected slice", item)
	}
}

// AssertSliceContains asserts that a slice contains an item
func AssertSliceContains[T comparable](t *testing.T, slice []T, item T) {
	for _, sliceItem := range slice {
		if sliceItem == item {
			return
		}
	}
	assert.Fail(t, fmt.Sprintf("Slice should contain item %v", item))
}

// AssertSliceNotContains asserts that a slice does not contain an item
func AssertSliceNotContains[T comparable](t *testing.T, slice []T, item T) {
	for _, sliceItem := range slice {
		if sliceItem == item {
			assert.Fail(t, fmt.Sprintf("Slice should not contain item %v", item))
			return
		}
	}
}

// AssertMapEqual asserts that two maps are equal
func AssertMapEqual[K comparable, V comparable](t *testing.T, expected, actual map[K]V) {
	require.Equal(t, len(expected), len(actual), "Maps should have the same length")
	
	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		assert.True(t, exists, "Key %v should exist in actual map", key)
		assert.Equal(t, expectedValue, actualValue, "Values for key %v should be equal", key)
	}
}

// AssertMapContains asserts that a map contains a key
func AssertMapContains[K comparable, V any](t *testing.T, m map[K]V, key K) {
	_, exists := m[key]
	assert.True(t, exists, "Map should contain key %v", key)
}

// AssertMapNotContains asserts that a map does not contain a key
func AssertMapNotContains[K comparable, V any](t *testing.T, m map[K]V, key K) {
	_, exists := m[key]
	assert.False(t, exists, "Map should not contain key %v", key)
}

// AssertJSONEqual asserts that two JSON strings are equal
func AssertJSONEqual(t *testing.T, expected, actual string) {
	var expectedJSON, actualJSON interface{}
	
	err := json.Unmarshal([]byte(expected), &expectedJSON)
	require.NoError(t, err, "Expected JSON should be valid")
	
	err = json.Unmarshal([]byte(actual), &actualJSON)
	require.NoError(t, err, "Actual JSON should be valid")
	
	assert.Equal(t, expectedJSON, actualJSON, "JSON should be equal")
}

// AssertJSONContains asserts that a JSON string contains a field with expected value
func AssertJSONContains(t *testing.T, jsonStr, field string, expectedValue interface{}) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	require.NoError(t, err, "JSON should be valid")
	
	actualValue, exists := data[field]
	assert.True(t, exists, "JSON should contain field %s", field)
	assert.Equal(t, expectedValue, actualValue, "Field %s should have expected value", field)
}

// AssertPointerEqual asserts that two pointers point to equal values
func AssertPointerEqual[T comparable](t *testing.T, expected, actual *T) {
	if expected == nil && actual == nil {
		return
	}
	require.NotNil(t, expected, "Expected pointer should not be nil")
	require.NotNil(t, actual, "Actual pointer should not be nil")
	assert.Equal(t, *expected, *actual, "Pointer values should be equal")
}

// AssertPointerNotNil asserts that a pointer is not nil
func AssertPointerNotNil[T any](t *testing.T, ptr *T) {
	assert.NotNil(t, ptr, "Pointer should not be nil")
}

// AssertPointerNil asserts that a pointer is nil
func AssertPointerNil[T any](t *testing.T, ptr *T) {
	assert.Nil(t, ptr, "Pointer should be nil")
}

// AssertError asserts that an error is not nil and has the expected message
func AssertError(t *testing.T, err error, expectedMessage string) {
	require.Error(t, err, "Error should not be nil")
	if expectedMessage != "" {
		assert.Contains(t, err.Error(), expectedMessage, "Error message should contain expected text")
	}
}

// AssertNoError asserts that an error is nil
func AssertNoError(t *testing.T, err error) {
	assert.NoError(t, err, "Error should be nil")
}

// AssertErrorType asserts that an error is of the expected type
func AssertErrorType[T error](t *testing.T, err error) {
	require.Error(t, err, "Error should not be nil")
	
	var expectedErr T
	assert.IsType(t, expectedErr, err, "Error should be of expected type")
}

// AssertFunctionPanics asserts that a function panics
func AssertFunctionPanics(t *testing.T, fn func()) {
	assert.Panics(t, fn, "Function should panic")
}

// AssertFunctionNotPanics asserts that a function does not panic
func AssertFunctionNotPanics(t *testing.T, fn func()) {
	assert.NotPanics(t, fn, "Function should not panic")
}

// AssertChannelClosed asserts that a channel is closed
func AssertChannelClosed[T any](t *testing.T, ch <-chan T) {
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "Channel should be closed")
	default:
		// Channel might not be closed yet, but we can't wait indefinitely
		assert.Fail(t, "Channel should be closed")
	}
}

// AssertChannelOpen asserts that a channel is open
func AssertChannelOpen[T any](t *testing.T, ch <-chan T) {
	select {
	case _, ok := <-ch:
		if !ok {
			assert.Fail(t, "Channel should be open")
		}
	default:
		// Channel is open (no immediate receive)
	}
}

// AssertStructFieldsEqual asserts that specific fields of two structs are equal
func AssertStructFieldsEqual(t *testing.T, expected, actual interface{}, fields ...string) {
	expectedValue := reflect.ValueOf(expected)
	actualValue := reflect.ValueOf(actual)

	require.Equal(t, expectedValue.Kind(), reflect.Struct, "Expected must be a struct")
	require.Equal(t, actualValue.Kind(), reflect.Struct, "Actual must be a struct")
	require.Equal(t, expectedValue.Type(), actualValue.Type(), "Structs must be of the same type")

	for _, field := range fields {
		expectedField := expectedValue.FieldByName(field)
		actualField := actualValue.FieldByName(field)

		assert.True(t, expectedField.IsValid(), "Expected struct should have field %s", field)
		assert.True(t, actualField.IsValid(), "Actual struct should have field %s", field)

		assert.Equal(t, expectedField.Interface(), actualField.Interface(), 
			fmt.Sprintf("Field %s should be equal", field))
	}
}

// AssertStructFieldsNotEqual asserts that specific fields of two structs are not equal
func AssertStructFieldsNotEqual(t *testing.T, expected, actual interface{}, fields ...string) {
	expectedValue := reflect.ValueOf(expected)
	actualValue := reflect.ValueOf(actual)

	require.Equal(t, expectedValue.Kind(), reflect.Struct, "Expected must be a struct")
	require.Equal(t, actualValue.Kind(), reflect.Struct, "Actual must be a struct")
	require.Equal(t, expectedValue.Type(), actualValue.Type(), "Structs must be of the same type")

	for _, field := range fields {
		expectedField := expectedValue.FieldByName(field)
		actualField := actualValue.FieldByName(field)

		assert.True(t, expectedField.IsValid(), "Expected struct should have field %s", field)
		assert.True(t, actualField.IsValid(), "Actual struct should have field %s", field)

		assert.NotEqual(t, expectedField.Interface(), actualField.Interface(), 
			fmt.Sprintf("Field %s should not be equal", field))
	}
}

// AssertStructHasField asserts that a struct has a specific field
func AssertStructHasField(t *testing.T, structValue interface{}, field string) {
	value := reflect.ValueOf(structValue)
	require.Equal(t, value.Kind(), reflect.Struct, "Value must be a struct")

	fieldValue := value.FieldByName(field)
	assert.True(t, fieldValue.IsValid(), "Struct should have field %s", field)
}

// AssertStructNotHasField asserts that a struct does not have a specific field
func AssertStructNotHasField(t *testing.T, structValue interface{}, field string) {
	value := reflect.ValueOf(structValue)
	require.Equal(t, value.Kind(), reflect.Struct, "Value must be a struct")

	fieldValue := value.FieldByName(field)
	assert.False(t, fieldValue.IsValid(), "Struct should not have field %s", field)
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}) {
	assert.NotNil(t, value, "Value should not be nil")
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, value interface{}) {
	assert.Nil(t, value, "Value should be nil")
}

// AssertEmpty asserts that a value is empty (zero value)
func AssertEmpty(t *testing.T, value interface{}) {
	assert.Empty(t, value, "Value should be empty")
}

// AssertNotEmpty asserts that a value is not empty
func AssertNotEmpty(t *testing.T, value interface{}) {
	assert.NotEmpty(t, value, "Value should not be empty")
}

// AssertZero asserts that a numeric value is zero
func AssertZero[T ~int | ~int32 | ~int64 | ~float32 | ~float64](t *testing.T, value T) {
	assert.Equal(t, T(0), value, "Value should be zero")
}

// AssertNotZero asserts that a numeric value is not zero
func AssertNotZero[T ~int | ~int32 | ~int64 | ~float32 | ~float64](t *testing.T, value T) {
	assert.NotEqual(t, T(0), value, "Value should not be zero")
}

// AssertPositive asserts that a numeric value is positive
func AssertPositive[T ~int | ~int32 | ~int64 | ~float32 | ~float64](t *testing.T, value T) {
	assert.Greater(t, value, T(0), "Value should be positive")
}

// AssertNegative asserts that a numeric value is negative
func AssertNegative[T ~int | ~int32 | ~int64 | ~float32 | ~float64](t *testing.T, value T) {
	assert.Less(t, value, T(0), "Value should be negative")
}

// AssertGreater asserts that a value is greater than another
func AssertGreater[T ~int | ~int32 | ~int64 | ~float32 | ~float64](t *testing.T, greater, less T) {
	assert.Greater(t, greater, less, "Value should be greater")
}

// AssertLess asserts that a value is less than another
func AssertLess[T ~int | ~int32 | ~int64 | ~float32 | ~float64](t *testing.T, less, greater T) {
	assert.Less(t, less, greater, "Value should be less")
}

// AssertBetween asserts that a value is between two values (inclusive)
func AssertBetween[T ~int | ~int32 | ~int64 | ~float32 | ~float64](t *testing.T, value, min, max T) {
	assert.GreaterOrEqual(t, value, min, "Value should be greater than or equal to min")
	assert.LessOrEqual(t, value, max, "Value should be less than or equal to max")
}

// AssertLength asserts that a slice, map, or string has the expected length
func AssertLength(t *testing.T, value interface{}, expectedLength int) {
	length := reflect.ValueOf(value).Len()
	assert.Equal(t, expectedLength, length, "Value should have expected length")
}

// AssertContains asserts that a slice, map, or string contains a value
func AssertContains(t *testing.T, container, element interface{}) {
	assert.Contains(t, container, element, "Container should contain element")
}

// AssertNotContains asserts that a slice, map, or string does not contain a value
func AssertNotContains(t *testing.T, container, element interface{}) {
	assert.NotContains(t, container, element, "Container should not contain element")
}

// AssertRegexp asserts that a string matches a regular expression
func AssertRegexp(t *testing.T, pattern, string string) {
	assert.Regexp(t, pattern, string, "String should match pattern")
}

// AssertNotRegexp asserts that a string does not match a regular expression
func AssertNotRegexp(t *testing.T, pattern, string string) {
	assert.NotRegexp(t, pattern, string, "String should not match pattern")
}