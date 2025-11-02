package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// EmailService interface defines email operations
type EmailService interface {
	SendVerificationEmail(ctx context.Context, to, code string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
	SendWelcomeEmail(ctx context.Context, to, firstName string) error
}

// SMTPEmailService implements EmailService using SMTP
type SMTPEmailService struct {
	config *config.EmailConfig
	from    string
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService(cfg *config.EmailConfig) *SMTPEmailService {
	return &SMTPEmailService{
		config: cfg,
		from:    fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromEmail),
	}
}

// SendVerificationEmail sends an email verification code
func (es *SMTPEmailService) SendVerificationEmail(ctx context.Context, to, code string) error {
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to Winkr!</h2>
			<p>Thank you for signing up. Please use the verification code below to verify your email address:</p>
			<div style="background-color: #f0f0f0; padding: 20px; border-radius: 5px; text-align: center; margin: 20px 0;">
				<h1 style="font-size: 32px; letter-spacing: 5px; color: #333;">%s</h1>
			</div>
			<p>This code will expire in 15 minutes.</p>
			<p>If you didn't request this verification, please ignore this email.</p>
			<p>Best regards,<br>The Winkr Team</p>
		</body>
		</html>
	`, code)

	return es.sendEmail(to, subject, body)
}

// SendPasswordResetEmail sends a password reset email
func (es *SMTPEmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	subject := "Reset Your Password"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>We received a request to reset your password for your Winkr account.</p>
			<p>Click the link below to reset your password:</p>
			<div style="margin: 20px 0;">
				<a href="%s/reset-password?token=%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">
					Reset Password
				</a>
			</div>
			<p>Or copy and paste this link in your browser:</p>
			<p style="background-color: #f0f0f0; padding: 10px; border-radius: 4px; word-break: break-all;">
				%s/reset-password?token=%s
			</p>
			<p>This link will expire in 1 hour.</p>
			<p>If you didn't request this password reset, please ignore this email.</p>
			<p>Best regards,<br>The Winkr Team</p>
		</body>
		</html>
	`, es.config.FrontendURL, token, es.config.FrontendURL, token)

	return es.sendEmail(to, subject, body)
}

// SendWelcomeEmail sends a welcome email to new users
func (es *SMTPEmailService) SendWelcomeEmail(ctx context.Context, to, firstName string) error {
	subject := "Welcome to Winkr!"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to Winkr, %s!</h2>
			<p>Thank you for joining our community. We're excited to have you on board!</p>
			<p>Here are a few things you can do to get started:</p>
			<ul>
				<li>Complete your profile with photos and bio</li>
				<li>Set your preferences to find matches</li>
				<li>Start swiping and connecting with others</li>
			</ul>
			<div style="margin: 20px 0;">
				<a href="%s/profile" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">
					Complete Your Profile
				</a>
			</div>
			<p>If you have any questions, feel free to contact our support team.</p>
			<p>Best regards,<br>The Winkr Team</p>
		</body>
		</html>
	`, firstName, es.config.FrontendURL)

	return es.sendEmail(to, subject, body)
}

// sendEmail sends an email using SMTP
func (es *SMTPEmailService) sendEmail(to, subject, body string) error {
	// Create message
	message := fmt.Sprintf("From: %s\r\n", es.from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-version: 1.0;\r\n"
	message += "Content-Type: text/html; charset=\"UTF-8\";\r\n"
	message += "\r\n"
	message += body

	// Connect to SMTP server
	auth := smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.Host)
	addr := fmt.Sprintf("%s:%d", es.config.Host, es.config.Port)

	err := smtp.SendMail(addr, auth, es.config.FromEmail, []string{to}, []byte(message))
	if err != nil {
		logger.Error("Failed to send email", err, "to", to, "subject", subject)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("Email sent successfully", "to", to, "subject", subject)
	return nil
}

// MockEmailService implements EmailService for testing
type MockEmailService struct {
	SentEmails []MockEmail
}

// MockEmail represents a sent email for testing
type MockEmail struct {
	To      string
	Subject string
	Body    string
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]MockEmail, 0),
	}
}

// SendVerificationEmail sends a verification email (mock)
func (mes *MockEmailService) SendVerificationEmail(ctx context.Context, to, code string) error {
	email := MockEmail{
		To:      to,
		Subject: "Verify Your Email Address",
		Body:    fmt.Sprintf("Verification code: %s", code),
	}
	mes.SentEmails = append(mes.SentEmails, email)
	return nil
}

// SendPasswordResetEmail sends a password reset email (mock)
func (mes *MockEmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	email := MockEmail{
		To:      to,
		Subject: "Reset Your Password",
		Body:    fmt.Sprintf("Reset token: %s", token),
	}
	mes.SentEmails = append(mes.SentEmails, email)
	return nil
}

// SendWelcomeEmail sends a welcome email (mock)
func (mes *MockEmailService) SendWelcomeEmail(ctx context.Context, to, firstName string) error {
	email := MockEmail{
		To:      to,
		Subject: "Welcome to Winkr!",
		Body:    fmt.Sprintf("Welcome %s!", firstName),
	}
	mes.SentEmails = append(mes.SentEmails, email)
	return nil
}

// GetLastSentEmail returns the last sent email (for testing)
func (mes *MockEmailService) GetLastSentEmail() *MockEmail {
	if len(mes.SentEmails) == 0 {
		return nil
	}
	return &mes.SentEmails[len(mes.SentEmails)-1]
}

// Clear clears all sent emails (for testing)
func (mes *MockEmailService) Clear() {
	mes.SentEmails = make([]MockEmail, 0)
}

// FindEmailByRecipient finds emails sent to a specific recipient (for testing)
func (mes *MockEmailService) FindEmailByRecipient(to string) []MockEmail {
	var emails []MockEmail
	for _, email := range mes.SentEmails {
		if strings.EqualFold(email.To, to) {
			emails = append(emails, email)
		}
	}
	return emails
}