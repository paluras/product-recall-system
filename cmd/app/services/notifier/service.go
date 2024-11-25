package notifier

import (
	"context"
	"log/slog"
	"time"

	"github.com/paluras/product-recall-system/internal/models"
	"github.com/paluras/product-recall-system/internal/notify"
)

type Service struct {
	db        *models.DB
	emailSvc  *notify.EmailService
	logger    *slog.Logger
	interval  time.Duration
	batchSize int
}

type ServiceConfig struct {
	DB        *models.DB
	EmailSvc  *notify.EmailService
	Logger    *slog.Logger
	Interval  time.Duration
	BatchSize int
}

func NewService(config ServiceConfig) *Service {
	return &Service{
		db:        config.DB,
		emailSvc:  config.EmailSvc,
		logger:    config.Logger,
		interval:  config.Interval,
		batchSize: config.BatchSize,
	}
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	if err := s.sendNotifications(); err != nil {
		s.logger.Error("initial notification failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.sendNotifications(); err != nil {
				s.logger.Error("notification failed", "error", err)
			}
		}
	}
}

func (s *Service) sendNotifications() error {
	items, err := s.db.GetUnnotifiedItems()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		s.logger.Info("no new items to notify about")
		return nil
	}

	subscribers, err := s.db.GetSubscribersMail()
	if err != nil {
		return err
	}

	if len(subscribers) == 0 {
		s.logger.Info("no subscribers to notify")
		return nil
	}

	for i := 0; i < len(items); i += s.batchSize {
		end := i + s.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]

		if err := s.emailSvc.SendBatchNotification(subscribers, batch); err != nil {
			s.logger.Error("failed to send batch notification",
				"error", err,
				"batch_start", i,
				"batch_end", end)
			continue
		}

		for _, item := range batch {
			if err := s.db.MarkAsNotified(item.ID); err != nil {
				s.logger.Error("failed to mark item as notified",
					"item_id", item.ID,
					"error", err)
			}
		}
	}

	s.logger.Info("notifications sent successfully",
		"items_count", len(items),
		"subscribers_count", len(subscribers))

	return nil
}

func (s *Service) SendVerificationEmail(email string) error {
	// Generate verification token
	token, err := s.db.CreatePendingSubscriber(email)
	if err != nil {
		return err
	}

	// Send verification email
	if err := s.emailSvc.SendVerificationEmail(email, token); err != nil {
		return err
	}

	s.logger.Info("verification email sent",
		"email", email)

	return nil
}

func (s *Service) CleanupExpiredVerifications() error {
	// Add this to your periodic tasks if needed
	affected, err := s.db.DeleteExpiredPendingSubscribers()
	if err != nil {
		return err
	}

	if affected > 0 {
		s.logger.Info("cleaned up expired verifications",
			"count", affected)
	}

	return nil
}
