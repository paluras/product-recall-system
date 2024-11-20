package main

import (
	"fmt"
	"log"

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

	awsConfig := notify.AWSConfig{
		Region:    "eu-central-1",
		FromEmail: "Latest Alert <alert@latest.produseretrase.eu>",
	}

	emailService, err := notify.NewEmailService(awsConfig, db)
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

	fmt.Printf("Subscriber %v", testRecipients)

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
