package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// HTTPTestHelper provides utilities for HTTP testing
type HTTPTestHelper struct {
	t      *testing.T
	router *gin.Engine
	server *httptest.Server
	client *http.Client
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(t *testing.T, router *gin.Engine) *HTTPTestHelper {
	return &HTTPTestHelper{
		t:      t,
		router: router,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StartServer starts a test server
func (h *HTTPTestHelper) StartServer() {
	h.server = httptest.NewServer(h.router)
}

// StopServer stops the test server
func (h *HTTPTestHelper) StopServer() {
	if h.server != nil {
		h.server.Close()
	}
}

// GetServerURL returns the server URL
func (h *HTTPTestHelper) GetServerURL() string {
	if h.server != nil {
		return h.server.URL
	}
	return ""
}

// NewRequest creates a new HTTP request
func (h *HTTPTestHelper) NewRequest(method, path string, body interface{}, headers map[string]string) *http.Request {
	var reqBody io.Reader
	var contentType string

	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
			contentType = "text/plain"
		case []byte:
			reqBody = bytes.NewReader(v)
			contentType = "application/octet-stream"
		default:
			jsonBody, err := json.Marshal(body)
			require.NoError(h.t, err, "Failed to marshal request body")
			reqBody = bytes.NewReader(jsonBody)
			contentType = "application/json"
		}
	}

	req := httptest.NewRequest(method, path, reqBody)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req
}

// ExecuteRequest executes an HTTP request and returns the response
func (h *HTTPTestHelper) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	return w
}

// ExecuteRequestWithServer executes an HTTP request using the test server
func (h *HTTPTestHelper) ExecuteRequestWithServer(method, path string, body interface{}, headers map[string]string) *http.Response {
	require.NotNil(h.t, h.server, "Test server must be started before making requests")

	req := h.NewRequest(method, h.server.URL+path, body, headers)
	
	resp, err := h.client.Do(req)
	require.NoError(h.t, err, "Failed to execute request")
	
	return resp
}

// Get executes a GET request
func (h *HTTPTestHelper) Get(path string, headers map[string]string) *httptest.ResponseRecorder {
	req := h.NewRequest(http.MethodGet, path, nil, headers)
	return h.ExecuteRequest(req)
}

// Post executes a POST request
func (h *HTTPTestHelper) Post(path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	req := h.NewRequest(http.MethodPost, path, body, headers)
	return h.ExecuteRequest(req)
}

// Put executes a PUT request
func (h *HTTPTestHelper) Put(path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	req := h.NewRequest(http.MethodPut, path, body, headers)
	return h.ExecuteRequest(req)
}

// Patch executes a PATCH request
func (h *HTTPTestHelper) Patch(path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	req := h.NewRequest(http.MethodPatch, path, body, headers)
	return h.ExecuteRequest(req)
}

// Delete executes a DELETE request
func (h *HTTPTestHelper) Delete(path string, headers map[string]string) *httptest.ResponseRecorder {
	req := h.NewRequest(http.MethodDelete, path, nil, headers)
	return h.ExecuteRequest(req)
}

// GetWithServer executes a GET request using the test server
func (h *HTTPTestHelper) GetWithServer(path string, headers map[string]string) *http.Response {
	return h.ExecuteRequestWithServer(http.MethodGet, path, nil, headers)
}

// PostWithServer executes a POST request using the test server
func (h *HTTPTestHelper) PostWithServer(path string, body interface{}, headers map[string]string) *http.Response {
	return h.ExecuteRequestWithServer(http.MethodPost, path, body, headers)
}

// PutWithServer executes a PUT request using the test server
func (h *HTTPTestHelper) PutWithServer(path string, body interface{}, headers map[string]string) *http.Response {
	return h.ExecuteRequestWithServer(http.MethodPut, path, body, headers)
}

// PatchWithServer executes a PATCH request using the test server
func (h *HTTPTestHelper) PatchWithServer(path string, body interface{}, headers map[string]string) *http.Response {
	return h.ExecuteRequestWithServer(http.MethodPatch, path, body, headers)
}

// DeleteWithServer executes a DELETE request using the test server
func (h *HTTPTestHelper) DeleteWithServer(path string, headers map[string]string) *http.Response {
	return h.ExecuteRequestWithServer(http.MethodDelete, path, nil, headers)
}

// CreateMultipartForm creates a multipart form for file uploads
func (h *HTTPTestHelper) CreateMultipartForm(fields map[string]string, files map[string][]byte) (bytes.Buffer, string) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add fields
	for key, value := range fields {
		err := writer.WriteField(key, value)
		require.NoError(h.t, err, "Failed to write field")
	}

	// Add files
	for fieldName, fileData := range files {
		part, err := writer.CreateFormFile(fieldName, "test.jpg")
		require.NoError(h.t, err, "Failed to create form file")
		
		_, err = part.Write(fileData)
		require.NoError(h.t, err, "Failed to write file data")
	}

	err := writer.Close()
	require.NoError(h.t, err, "Failed to close multipart writer")

	return requestBody, writer.FormDataContentType()
}

// PostMultipart executes a POST request with multipart form data
func (h *HTTPTestHelper) PostMultipart(path string, fields map[string]string, files map[string][]byte, headers map[string]string) *httptest.ResponseRecorder {
	body, contentType := h.CreateMultipartForm(fields, files)
	
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = contentType
	
	req := h.NewRequest(http.MethodPost, path, body.Bytes(), headers)
	return h.ExecuteRequest(req)
}

// PostMultipartWithServer executes a POST request with multipart form data using the test server
func (h *HTTPTestHelper) PostMultipartWithServer(path string, fields map[string]string, files map[string][]byte, headers map[string]string) *http.Response {
	body, contentType := h.CreateMultipartForm(fields, files)
	
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = contentType
	
	return h.ExecuteRequestWithServer(http.MethodPost, path, body.Bytes(), headers)
}

// WebSocketHelper provides utilities for WebSocket testing
type WebSocketHelper struct {
	t      *testing.T
	server *httptest.Server
	dialer *websocket.Dialer
}

// NewWebSocketHelper creates a new WebSocket test helper
func NewWebSocketHelper(t *testing.T, server *httptest.Server) *WebSocketHelper {
	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	return &WebSocketHelper{
		t:      t,
		server: server,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Connect establishes a WebSocket connection
func (w *WebSocketHelper) Connect(path string, headers http.Header) *websocket.Conn {
	wsURL := "ws" + strings.TrimPrefix(w.server.URL, "http") + path
	
	conn, _, err := w.dialer.Dial(wsURL, headers)
	require.NoError(w.t, err, "Failed to connect to WebSocket")
	
	return conn
}

// ConnectWithToken establishes a WebSocket connection with authentication token
func (w *WebSocketHelper) ConnectWithToken(path, token string) *websocket.Conn {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)
	
	return w.Connect(path, headers)
}

// SendMessage sends a message through WebSocket connection
func (w *WebSocketHelper) SendMessage(conn *websocket.Conn, message interface{}) {
	data, err := json.Marshal(message)
	require.NoError(w.t, err, "Failed to marshal message")
	
	err = conn.WriteMessage(websocket.TextMessage, data)
	require.NoError(w.t, err, "Failed to send message")
}

// ReadMessage reads a message from WebSocket connection
func (w *WebSocketHelper) ReadMessage(conn *websocket.Conn) []byte {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	messageType, data, err := conn.ReadMessage()
	require.NoError(w.t, err, "Failed to read message")
	require.Equal(w.t, websocket.TextMessage, messageType, "Expected text message")
	
	return data
}

// ReadMessageInto reads a message from WebSocket connection into the provided interface
func (w *WebSocketHelper) ReadMessageInto(conn *websocket.Conn, v interface{}) {
	data := w.ReadMessage(conn)
	err := json.Unmarshal(data, v)
	require.NoError(w.t, err, "Failed to unmarshal message")
}

// ExpectMessage expects to receive a specific message within timeout
func (w *WebSocketHelper) ExpectMessage(conn *websocket.Conn, expectedMessage interface{}) {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	data := w.ReadMessage(conn)
	
	expectedData, err := json.Marshal(expectedMessage)
	require.NoError(w.t, err, "Failed to marshal expected message")
	
	require.JSONEq(w.t, string(expectedData), string(data), "Received message does not match expected")
}

// ExpectMessageType expects to receive a message of specific type
func (w *WebSocketHelper) ExpectMessageType(conn *websocket.Conn, expectedType string) {
	var message map[string]interface{}
	w.ReadMessageInto(conn, &message)
	
	msgType, ok := message["type"].(string)
	require.True(w.t, ok, "Message should have type field")
	require.Equal(w.t, expectedType, msgType, "Message type should match expected")
}

// CloseConnection closes the WebSocket connection
func (w *WebSocketHelper) CloseConnection(conn *websocket.Conn) {
	err := conn.Close()
	require.NoError(w.t, err, "Failed to close WebSocket connection")
}

// AuthHelper provides utilities for authentication testing
type AuthHelper struct {
	t      *testing.T
	http   *HTTPTestHelper
	config *TestConfig
}

// NewAuthHelper creates a new authentication test helper
func NewAuthHelper(t *testing.T, httpHelper *HTTPTestHelper, config *TestConfig) *AuthHelper {
	return &AuthHelper{
		t:      t,
		http:   httpHelper,
		config: config,
	}
}

// RegisterUser registers a new user
func (a *AuthHelper) RegisterUser(email, password, firstName, lastName string) *http.Response {
	payload := map[string]interface{}{
		"email":     email,
		"password":  password,
		"firstName": firstName,
		"lastName":  lastName,
	}
	
	return a.http.PostWithServer("/auth/register", payload, nil)
}

// LoginUser logs in a user
func (a *AuthHelper) LoginUser(email, password string) *http.Response {
	payload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	
	return a.http.PostWithServer("/auth/login", payload, nil)
}

// GetAuthToken logs in a user and returns the auth token
func (a *AuthHelper) GetAuthToken(email, password string) string {
	resp := a.LoginUser(email, password)
	require.Equal(a.t, http.StatusOK, resp.StatusCode)
	
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(a.t, err, "Failed to decode login response")
	
	token, ok := result["token"].(string)
	require.True(a.t, ok, "Response should contain token")
	
	return token
}

// GetAuthHeaders returns authentication headers
func (a *AuthHelper) GetAuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// CreateAuthenticatedUser creates a new user and returns auth token
func (a *AuthHelper) CreateAuthenticatedUser(email, password, firstName, lastName string) string {
	// Register user
	resp := a.RegisterUser(email, password, firstName, lastName)
	require.Equal(a.t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
	
	// Login and get token
	return a.GetAuthToken(email, password)
}

// RefreshToken refreshes an authentication token
func (a *AuthHelper) RefreshToken(refreshToken string) *http.Response {
	payload := map[string]interface{}{
		"refreshToken": refreshToken,
	}
	
	return a.http.PostWithServer("/auth/refresh", payload, nil)
}

// LogoutUser logs out a user
func (a *AuthHelper) LogoutUser(token string) *http.Response {
	headers := a.GetAuthHeaders(token)
	return a.http.PostWithServer("/auth/logout", nil, headers)
}

// RequestPasswordReset requests a password reset
func (a *AuthHelper) RequestPasswordReset(email string) *http.Response {
	payload := map[string]interface{}{
		"email": email,
	}
	
	return a.http.PostWithServer("/auth/request-password-reset", payload, nil)
}

// ResetPassword resets a user's password
func (a *AuthHelper) ResetPassword(token, newPassword string) *http.Response {
	payload := map[string]interface{}{
		"token":    token,
		"password": newPassword,
	}
	
	return a.http.PostWithServer("/auth/reset-password", payload, nil)
}

// VerifyEmail verifies a user's email
func (a *AuthHelper) VerifyEmail(token string) *http.Response {
	payload := map[string]interface{}{
		"token": token,
	}
	
	return a.http.PostWithServer("/auth/verify-email", payload, nil)
}

// ChangePassword changes a user's password
func (a *AuthHelper) ChangePassword(token, currentPassword, newPassword string) *http.Response {
	payload := map[string]interface{}{
		"currentPassword": currentPassword,
		"newPassword":     newPassword,
	}
	
	headers := a.GetAuthHeaders(token)
	return a.http.PostWithServer("/auth/change-password", payload, headers)
}

// APIHelper provides utilities for API testing
type APIHelper struct {
	t      *testing.T
	http   *HTTPTestHelper
	auth   *AuthHelper
	assert *AssertionHelper
}

// NewAPIHelper creates a new API test helper
func NewAPIHelper(t *testing.T, httpHelper *HTTPTestHelper, authHelper *AuthHelper) *APIHelper {
	return &APIHelper{
		t:      t,
		http:   httpHelper,
		auth:   authHelper,
		assert: NewAssertionHelper(t),
	}
}

// CreateAuthenticatedRequest creates an authenticated HTTP request
func (a *APIHelper) CreateAuthenticatedRequest(method, path string, body interface{}, token string) *http.Response {
	headers := a.auth.GetAuthHeaders(token)
	return a.http.ExecuteRequestWithServer(method, path, body, headers)
}

// GetAuthenticated performs an authenticated GET request
func (a *APIHelper) GetAuthenticated(path, token string) *http.Response {
	return a.CreateAuthenticatedRequest(http.MethodGet, path, nil, token)
}

// PostAuthenticated performs an authenticated POST request
func (a *APIHelper) PostAuthenticated(path string, body interface{}, token string) *http.Response {
	return a.CreateAuthenticatedRequest(http.MethodPost, path, body, token)
}

// PutAuthenticated performs an authenticated PUT request
func (a *APIHelper) PutAuthenticated(path string, body interface{}, token string) *http.Response {
	return a.CreateAuthenticatedRequest(http.MethodPut, path, body, token)
}

// PatchAuthenticated performs an authenticated PATCH request
func (a *APIHelper) PatchAuthenticated(path string, body interface{}, token string) *http.Response {
	return a.CreateAuthenticatedRequest(http.MethodPatch, path, body, token)
}

// DeleteAuthenticated performs an authenticated DELETE request
func (a *APIHelper) DeleteAuthenticated(path, token string) *http.Response {
	return a.CreateAuthenticatedRequest(http.MethodDelete, path, nil, token)
}

// UploadFile uploads a file with authentication
func (a *APIHelper) UploadFile(path string, fields map[string]string, files map[string][]byte, token string) *http.Response {
	headers := a.auth.GetAuthHeaders(token)
	return a.http.PostMultipartWithServer(path, fields, files, headers)
}

// ExpectSuccess expects a successful response
func (a *APIHelper) ExpectSuccess(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusOK, "application/json")
}

// ExpectCreated expects a created response
func (a *APIHelper) ExpectCreated(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusCreated, "application/json")
}

// ExpectBadRequest expects a bad request response
func (a *APIHelper) ExpectBadRequest(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusBadRequest, "application/json")
}

// ExpectUnauthorized expects an unauthorized response
func (a *APIHelper) ExpectUnauthorized(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusUnauthorized, "application/json")
}

// ExpectForbidden expects a forbidden response
func (a *APIHelper) ExpectForbidden(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusForbidden, "application/json")
}

// ExpectNotFound expects a not found response
func (a *APIHelper) ExpectNotFound(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusNotFound, "application/json")
}

// ExpectConflict expects a conflict response
func (a *APIHelper) ExpectConflict(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusConflict, "application/json")
}

// ExpectTooManyRequests expects a too many requests response
func (a *APIHelper) ExpectTooManyRequests(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusTooManyRequests, "application/json")
}

// ExpectInternalServerError expects an internal server error response
func (a *APIHelper) ExpectInternalServerError(resp *http.Response) {
	a.assert.AssertHTTPResponse(resp, http.StatusInternalServerError, "application/json")
}

// ExtractData extracts data from JSON response
func (a *APIHelper) ExtractData(resp *http.Response, target interface{}) {
	data := a.assert.AssertJSONResponse(resp)
	a.assert.AssertMapContainsKey(data, "data")
	
	jsonData, err := json.Marshal(data["data"])
	require.NoError(a.t, err, "Failed to marshal response data")
	
	err = json.Unmarshal(jsonData, target)
	require.NoError(a.t, err, "Failed to unmarshal response data")
}

// ExtractMessage extracts message from JSON response
func (a *APIHelper) ExtractMessage(resp *http.Response) string {
	data := a.assert.AssertJSONResponse(resp)
	a.assert.AssertMapContainsKey(data, "message")
	
	return data["message"].(string)
}

// ExtractError extracts error from JSON response
func (a *APIHelper) ExtractError(resp *http.Response) string {
	data := a.assert.AssertJSONResponse(resp)
	a.assert.AssertMapContainsKey(data, "error")
	
	return data["error"].(string)
}

// ExtractPagination extracts pagination info from JSON response
func (a *APIHelper) ExtractPagination(resp *http.Response) map[string]interface{} {
	data := a.assert.AssertJSONResponse(resp)
	a.assert.AssertMapContainsKey(data, "pagination")
	
	return data["pagination"].(map[string]interface{})
}

// BuildURL builds a URL with query parameters
func (a *APIHelper) BuildURL(baseURL string, params map[string]string) string {
	if len(params) == 0 {
		return baseURL
	}
	
	values := url.Values{}
	for key, value := range params {
		values.Add(key, value)
	}
	
	return baseURL + "?" + values.Encode()
}

// TestContext provides a complete testing context
type TestContext struct {
	T           *testing.T
	Config      *TestConfig
	HTTP        *HTTPTestHelper
	WebSocket   *WebSocketHelper
	Auth        *AuthHelper
	API         *APIHelper
	Assert      *AssertionHelper
	DataManager *TestDataManager
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T, router *gin.Engine, config *TestConfig, dataManager *TestDataManager) *TestContext {
	httpHelper := NewHTTPTestHelper(t, router)
	httpHelper.StartServer()
	
	authHelper := NewAuthHelper(t, httpHelper, config)
	wsHelper := NewWebSocketHelper(t, httpHelper.server)
	apiHelper := NewAPIHelper(t, httpHelper, authHelper)
	assertHelper := NewAssertionHelper(t)
	
	return &TestContext{
		T:           t,
		Config:      config,
		HTTP:        httpHelper,
		WebSocket:   wsHelper,
		Auth:        authHelper,
		API:         apiHelper,
		Assert:      assertHelper,
		DataManager: dataManager,
	}
}

// Cleanup cleans up the test context
func (tc *TestContext) Cleanup() {
	if tc.HTTP != nil {
		tc.HTTP.StopServer()
	}
}

// CreateTestUser creates a test user and returns auth token
func (tc *TestContext) CreateTestUser(email, password, firstName, lastName string) string {
	return tc.Auth.CreateAuthenticatedUser(email, password, firstName, lastName)
}

// CreateDefaultTestUser creates a default test user
func (tc *TestContext) CreateDefaultTestUser() string {
	email := fmt.Sprintf("testuser_%d@example.com", time.Now().Unix())
	return tc.CreateTestUser(email, "password123", "Test", "User")
}

// WithTimeout executes a function with timeout
func (tc *TestContext) WithTimeout(timeout time.Duration, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return fn(ctx)
}

// Eventually asserts that a condition eventually becomes true
func (tc *TestContext) Eventually(condition func() bool, timeout time.Duration, message string) {
	tc.Assert.AssertEventually(condition, timeout, message)
}