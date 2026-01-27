package main

import (
	"log"
	"os"

	"github.com/paluras/product-recall-system/configs"
	"github.com/paluras/product-recall-system/internal/models"
	"github.com/paluras/product-recall-system/internal/notify"
)

func main() {
	conf := configs.ParseFlags()

	db, err := models.NewDB(conf.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Fatal("RESEND_API_KEY environment variable is required")
	}

	emailConfig := notify.EmailConfig{
		APIKey:    apiKey,
		FromEmail: "Latest Alert <alert@latest.produseretrase.eu>",
	}

	emailService, err := notify.NewEmailService(emailConfig, db)
	if err != nil {
		log.Fatal("Failed to create email service:", err)
	}

	items, err := db.GetUnnotifiedItems()
	if err != nil {
		log.Fatal(err)
	}

	if len(items) == 0 {
		log.Println("No new items to notify about")
		return
	}
	subscribers, err := db.GetSubscribersMail()
	if err != nil {
		log.Panic("Failed to fetch subscribers", err)
	}

	testRecipients := subscribers

	log.Printf("Subscribers: %v", testRecipients)

	err = emailService.SendBatchNotification(testRecipients, items)
	if err != nil {
		log.Printf("Failed to send notifications: %v", err)
		return
	}

	for _, item := range items {
		if err := db.MarkAsNotified(item.ID); err != nil {
			log.Printf("Failed to mark item %d as notified: %v", item.ID, err)
		}
	}

	log.Printf("Successfully sent notifications for %d items", len(items))
}
