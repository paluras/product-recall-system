package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type Subscriber struct {
	ID               string
	Email            string
	CreatedAt        time.Time
	UnsubscribeToken string
}

type PendingSubscriber struct {
	Email     string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (db *DB) CreatePendingSubscriber(email string) (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(token)

	query := `
        INSERT INTO pending_subscribers (email, token, expires_at)
        VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 24 HOUR))
    `

	_, err := db.Exec(query, email, tokenStr)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (db *DB) ConfirmSubscriber(token string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var email string
	err = tx.QueryRow(`
        SELECT email FROM pending_subscribers
        WHERE token = ? AND expires_at > NOW()
    `, token).Scan(&email)
	if err == sql.ErrNoRows {
		return fmt.Errorf("invalid or expired token")
	}
	if err != nil {
		return fmt.Errorf("failed to get pending subscription: %w", err)
	}

	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM subscribers WHERE email = ?)`, email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check subscriber existence: %w", err)
	}
	if exists {
		_, err = tx.Exec(`DELETE FROM pending_subscribers WHERE token = ?`, token)
		if err != nil {
			return fmt.Errorf("failed to clean up pending subscription: %w", err)
		}
		return fmt.Errorf("email is already subscribed")
	}

	unsubscribeToken := generateUnsubscribeToken()
	_, err = tx.Exec(`
        INSERT INTO subscribers (email, unsubscribe_token)
        VALUES (?, ?)
    `, email, unsubscribeToken)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM pending_subscribers WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("failed to remove pending subscription: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (db *DB) DeleteExpiredPendingSubscribers() (int64, error) {
	result, err := db.Exec(`
        DELETE FROM pending_subscribers
        WHERE expires_at < NOW()
    `)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

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
	stmt := `
        SELECT EXISTS(
            SELECT 1 FROM subscribers WHERE email = ?
            UNION
            SELECT 1 FROM pending_subscribers WHERE email = ?
        )
    `
	err := db.QueryRow(stmt, email, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (db *DB) PendingEmailExists(email string) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM pending_subscribers WHERE email = ?)`
	err := db.QueryRow(stmt, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (db *DB) DeletePendingSubscriber(email string) error {
	_, err := db.Exec(`DELETE FROM pending_subscribers WHERE email = ?`, email)
	return err
}
