package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SMSService interface defines SMS operations
type SMSService interface {
	SendVerificationSMS(ctx context.Context, to, code string) error
}

// TwilioSMSService implements SMSService using Twilio
type TwilioSMSService struct {
	config *config.SMSConfig
	client *http.Client
}

// NewTwilioSMSService creates a new Twilio SMS service
func NewTwilioSMSService(cfg *config.SMSConfig) *TwilioSMSService {
	return &TwilioSMSService{
		config: cfg,
		client: &http.Client{},
	}
}

// SendVerificationSMS sends a verification SMS using Twilio
func (ts *TwilioSMSService) SendVerificationSMS(ctx context.Context, to, code string) error {
	// Format phone number if needed
	if !strings.HasPrefix(to, "+") {
		to = "+" + to
	}

	// Prepare message
	message := fmt.Sprintf("Your Winkr verification code is: %s. This code will expire in 15 minutes.", code)

	// Build request data
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", ts.Config.FromNumber)
	data.Set("Body", message)

	// Create request
	url := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", ts.Config.AccountSID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("Failed to create Twilio request", err)
		return fmt.Errorf("failed to create SMS request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(ts.Config.AccountSID, ts.Config.AuthToken)

	// Send request
	resp, err := ts.client.Do(req)
	if err != nil {
		logger.Error("Failed to send Twilio SMS", err)
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		var errorResp TwilioError
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			logger.Error("Failed to decode Twilio error response", err)
			return fmt.Errorf("SMS sending failed with status: %d", resp.StatusCode)
		}
		logger.Error("Twilio SMS error", nil, "code", errorResp.Code, "message", errorResp.Message)
		return fmt.Errorf("SMS sending failed: %s", errorResp.Message)
	}

	logger.Info("SMS sent successfully", "to", to)
	return nil
}

// TwilioError represents a Twilio API error response
type TwilioError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

// MockSMSService implements SMSService for testing
type MockSMSService struct {
	SentSMSs []MockSMS
}

// MockSMS represents a sent SMS for testing
type MockSMS struct {
	To   string
	Body string
}

// NewMockSMSService creates a new mock SMS service
func NewMockSMSService() *MockSMSService {
	return &MockSMSService{
		SentSMSs: make([]MockSMS, 0),
	}
}

// SendVerificationSMS sends a verification SMS (mock)
func (mss *MockSMSService) SendVerificationSMS(ctx context.Context, to, code string) error {
	sms := MockSMS{
		To:   to,
		Body:  fmt.Sprintf("Your Winkr verification code is: %s", code),
	}
	mss.SentSMSs = append(mss.SentSMSs, sms)
	return nil
}

// GetLastSentSMS returns the last sent SMS (for testing)
func (mss *MockSMSService) GetLastSentSMS() *MockSMS {
	if len(mss.SentSMSs) == 0 {
		return nil
	}
	return &mss.SentSMSs[len(mss.SentSMSs)-1]
}

// Clear clears all sent SMSs (for testing)
func (mss *MockSMSService) Clear() {
	mss.SentSMSs = make([]MockSMS, 0)
}

// FindSMSByRecipient finds SMSs sent to a specific recipient (for testing)
func (mss *MockSMSService) FindSMSByRecipient(to string) []MockSMS {
	var smsList []MockSMS
	for _, sms := range mss.SentSMSs {
		if strings.EqualFold(sms.To, to) {
			smsList = append(smsList, sms)
		}
	}
	return smsList
}