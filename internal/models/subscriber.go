package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

type Subscriber struct {
	ID               string
	Email            string
	CreatedAt        time.Time
	UnsubscribeToken string
}

// refactor into a helper file
func generateUnsubscribeToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (db *DB) CreateUnsubscribeToken(email string) (string, error) {
	token := generateUnsubscribeToken()
	query := `UPDATE subscribers SET unsubscribe_token = ? WHERE email = ?`
	_, err := db.Exec(query, token, email)
	return token, err
}

func (db *DB) UnsubscribeWithToken(token string) error {
	query := `DELETE FROM subscribers WHERE unsubscribe_token = ?`
	result, err := db.Exec(query, token)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("invalid unsubscribe token")
	}
	return nil
}

func (db *DB) AddSubscriber(email string) error {
	query := `INSERT INTO subscribers (email) VALUES (?)`
	_, err := db.Exec(query, email)
	return err
}

func (db *DB) GetSubscribersMail() ([]string, error) {
	query := `SELECT email FROM subscribers`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscribers []string

	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		subscribers = append(subscribers, s)
	}

	return subscribers, nil
}

func (db *DB) EmailExists(email string) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM subscribers WHERE email = ?)`
	err := db.QueryRow(stmt, email).Scan(&exists)
	if err != nil {
		log.Printf("EmailExists error: %v", err)
		return false, err
	}
	return exists, nil
}
