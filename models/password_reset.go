package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/Rahul4469/cloud-memory/rand"
)

const (
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID     int
	UserID int
	//Token is only set when a PasswordReset is being created.
	Token     string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct {
	DB *sql.DB
	//BytesPerToken is used to determine how many bytes to use when generating
	//each password token. Is this value is not set or is less than the
	// MinBytesPerToken const  it will be ignored and MinBytesPerToken will be used.
	BytesPerToken int
	//Duration is the amount of time that a PasswordRest is valid for.
	Duration time.Duration
}

// select user ID using email from users db -> create token bytes, set expiration
// save userID, tokenHash, expiration into password_rest db -> return *PasswordReset fields to be used in controller
func (service *PasswordResetService) Create(email string) (*PasswordReset, error) {
	//verify we have a valid email addres for a user, and get the user's ID
	email = strings.ToLower(email)
	var userID int
	row := service.DB.QueryRow(`
	SELECT id FROM users WHERE email = $1;`, email)
	err := row.Scan(&userID)
	if err != nil {
		//Consider returning a specific error when the user does not exist.
		return nil, fmt.Errorf("create: %w", err)
	}
	//Build the PasswordReset
	bytesPerToken := service.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create password: %w", err)
	}

	duration := service.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}
	pwReset := PasswordReset{
		UserID:    userID,
		Token:     token,
		TokenHash: service.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	//Insert the PasswordRest into the DB
	row = service.DB.QueryRow(`
		INSERT INTO password_resets (user_id, token_hash, expires_at)
		VALUES($1, $2, $3) ON CONFLICT (user_id)
		DO UPDATE
		SET token_hash = $2, expires_at = $3
		RETURNING id`, pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
	err = row.Scan(&pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create password: %w", err)
	}

	return &pwReset, nil
}

// Validate token/ extract pw_rest(token, expiration)
// -> select user info affiliated with the password_reset
// -> Delete pw_rest so that it cant be used again
func (service *PasswordResetService) Consume(token string) (*User, error) {
	tokenHash := service.hash(token)
	var user User
	var pwReset PasswordReset
	row := service.DB.QueryRow(`
	SELECT password_resets.id, password_resets.expires_at,
		users.id,
		users.email,
		users.password_hash
	FROM password_resets
		JOIN users on users.id = password_resets.user_id
	WHERE password_resets.token_hash = $1`, tokenHash)
	err := row.Scan(&pwReset.ID, &pwReset.ExpiresAt,
		&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}
	//checking token expiration
	if time.Now().After(pwReset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %w", err)
	}
	err = service.delete(pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("cosume: %w", err)
	}

	return &user, nil
}

func (service *PasswordResetService) delete(id int) error {
	_, err := service.DB.Exec(`
		DELETE FROM password_resets
		WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete %w", err)
	}
	return nil
}

func (service *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
