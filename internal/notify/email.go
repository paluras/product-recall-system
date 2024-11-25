package notify

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/paluras/product-recall-system/internal/models"
	"github.com/paluras/product-recall-system/internal/templates/email"
)

const (
	defaultMaxSendRate  = 14                    // AWS SES default limit
	defaultBatchSize    = 50                    // Process 50 emails at a time
	defaultMaxRetries   = 3                     // Number of retries for failed sends
	defaultSendInterval = time.Millisecond * 71 // ~14 emails per second
)

type AWSConfig struct {
	Region    string
	FromEmail string
}

type EmailService struct {
	sesClient    *ses.Client
	config       AWSConfig
	db           *models.DB
	maxSendRate  int
	sendInterval time.Duration
	logger       *slog.Logger
}

func NewEmailService(awsConfig AWSConfig, db *models.DB) (*EmailService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsConfig.Region),
	)
	if err != nil {
		return nil, err
	}

	return &EmailService{
		sesClient:    ses.NewFromConfig(cfg),
		config:       awsConfig,
		db:           db,
		maxSendRate:  defaultMaxSendRate,
		sendInterval: defaultSendInterval,
		logger:       slog.Default(),
	}, nil
}

func (s *EmailService) SendBatchNotification(recipients []string, items []models.ScrapedItem) error {
	// Create channels for rate limiting and error handling
	semaphore := make(chan struct{}, s.maxSendRate)
	errorChan := make(chan error, len(recipients))
	var wg sync.WaitGroup

	// Process recipients in chunks
	for i := 0; i < len(recipients); i += defaultBatchSize {
		end := i + defaultBatchSize
		if end > len(recipients) {
			end = len(recipients)
		}

		chunk := recipients[i:end]
		s.logger.Info("processing email chunk",
			"start", i,
			"end", end,
			"size", len(chunk))

		// Process each recipient in the chunk
		for _, recipient := range chunk {
			wg.Add(1)
			go func(email string) {
				defer wg.Done()

				// Implement rate limiting
				semaphore <- struct{}{} // Acquire token
				defer func() {
					<-semaphore // Release token
					time.Sleep(s.sendInterval)
				}()

				// Create unsubscribe token
				token, err := s.db.CreateUnsubscribeToken(email)
				if err != nil {
					errorChan <- fmt.Errorf("failed to create token for %s: %w", email, err)
					return
				}

				// Prepare email content
				htmlBody, textBody, err := s.prepareEmailContent(items, token)
				if err != nil {
					errorChan <- fmt.Errorf("failed to prepare email for %s: %w", email, err)
					return
				}

				// Send email with retry
				if err := s.sendEmailWithRetry(email, htmlBody, textBody); err != nil {
					errorChan <- fmt.Errorf("failed to send email to %s: %w", email, err)
					return
				}

				s.logger.Info("email sent successfully", "recipient", email)
			}(recipient)
		}

		// Wait for current chunk to complete
		wg.Wait()

		// Check for any errors in the chunk
		select {
		case err := <-errorChan:
			s.logger.Error("error in batch", "error", err)
			// Continue processing despite errors
		default:
			// No errors, continue
		}
	}

	// Final error check
	select {
	case err := <-errorChan:
		return fmt.Errorf("batch notification failed: %w", err)
	default:
		return nil
	}
}

func (s *EmailService) sendEmailWithRetry(recipient, htmlBody, textBody string) error {
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(htmlBody),
				},
				Text: &types.Content{
					Data: aws.String(textBody),
				},
			},
			Subject: &types.Content{
				// You might want to make this configurable
				Data: aws.String("Confirmă abonarea la Alerte Retragere Produse"),
			},
		},
		Source: &s.config.FromEmail,
	}

	var lastErr error
	for i := 0; i < defaultMaxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(1<<uint(i)) * time.Second
			s.logger.Info("retrying send",
				"recipient", recipient,
				"attempt", i+1,
				"backoff", backoff)
			time.Sleep(backoff)
		}

		_, err := s.sesClient.SendEmail(context.Background(), input)
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetryableError(err) {
			return err
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryableError(err error) bool {
	return strings.Contains(err.Error(), "Throttling") ||
		strings.Contains(err.Error(), "Maximum sending rate exceeded") ||
		strings.Contains(err.Error(), "Network Error")
}

func (s *EmailService) prepareEmailContent(items []models.ScrapedItem, token string) (string, string, error) {
	data := struct {
		Items            []models.ScrapedItem
		UnsubscribeToken string
	}{
		Items:            items,
		UnsubscribeToken: token,
	}

	// Prepare HTML content
	htmlTmpl, err := template.New("email").Parse(email.NotificationHTMLTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse HTML template: %w", err)
	}
	var htmlBuffer bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuffer, data); err != nil {
		return "", "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	// Prepare text content
	textTmpl, err := template.New("email").Parse(email.NotificationTextTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse text template: %w", err)
	}
	var textBuffer bytes.Buffer
	if err := textTmpl.Execute(&textBuffer, data); err != nil {
		return "", "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return htmlBuffer.String(), textBuffer.String(), nil
}

func (s *EmailService) SendVerificationEmail(recipientEmail, token string) error {
	data := struct {
		Token string
	}{
		Token: token,
	}

	// Prepare HTML content
	htmlTmpl, err := template.New("verification_email").Parse(email.VerificationHTMLTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse verification HTML template: %w", err)
	}
	var htmlBuffer bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuffer, data); err != nil {
		return fmt.Errorf("failed to execute verification HTML template: %w", err)
	}

	// Prepare text content
	textTmpl, err := template.New("verification_email_text").Parse(email.VerificationTextTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse verification text template: %w", err)
	}
	var textBuffer bytes.Buffer
	if err := textTmpl.Execute(&textBuffer, data); err != nil {
		return fmt.Errorf("failed to execute verification text template: %w", err)
	}

	// Send email using the shared retry mechanism
	return s.sendEmailWithRetry(recipientEmail, htmlBuffer.String(), textBuffer.String())
}
