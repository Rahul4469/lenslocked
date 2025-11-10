package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/Rahul4469/cloud-memory/rand"
)

// minimum number of bytes to be used for each session token
const MinBytesPerToken = 32

type Session struct {
	ID     int
	UserID int
	//Token is only set when creating a new session.when looking up a session
	//this will be left empy, as we only store the hash of a session token
	//in our database a we cannot reverse it into a raw token.
	Token     string
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
	//BytesPerToken is used to determine how many bytes to use when generating
	//each session token. Is this value is not set or is less than the
	// MinBytesPerToken const  it will be ignored and MinBytesPerToken will be used.
	BytesPerToken int
}

func (ss *SessionService) Create(userID int) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	session := Session{
		UserID:    userID,
		Token:     token,
		TokenHash: ss.hash(token),
	}
	//Updating existing session
	//1. Try to update the existing user's session
	//2. If err, create new session(err= no session_ID for that user_ID)
	row := ss.DB.QueryRow(`
		INSERT INTO sessions (user_id, token_hash)
		VALUES($1, $2) ON CONFLICT (user_id)
		DO UPDATE
		SET token_hash = $2
		RETURNING id`, session.UserID, session.TokenHash)
	err = row.Scan(&session.ID) //IMP: row.Scan populates the data from query into the empty struct/object
	// if err == sql.ErrNoRows {
	// 	row := ss.DB.QueryRow(`
	// 		INSERT INTO sessions (user_id, token_hash)
	// 		VALUES ($1, $2)
	// 		RETURNING id;`, session.UserID, session.TokenHash)
	// 	err = row.Scan(&session.ID)
	// }

	// row := ss.DB.QueryRow(`
	// 	INSERT INTO sessions (user_id, token_hash)
	// 	VALUES ($1, $2)
	// 	RETURNING id;`, session.UserID, session.TokenHash)
	// err = row.Scan(&session.ID)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	return &session, nil
}

// r.token -> session table(match usedID from matching token and store in row) ->
// user table(select email,passwordHash using the userID)
func (ss *SessionService) User(token string) (*User, error) {
	//1. Hash the session tokens
	tokenHash := ss.hash(token)
	//2. Query for the sessions with that hash
	var user User
	row := ss.DB.QueryRow(`
		SELECT users.id,
    		users.email,
    		users.password_hash
		FROM sessions
    	JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1;`, tokenHash)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	// row := ss.DB.QueryRow(`
	// 	SELECT user_id
	// 	FROM sessions
	// 	WHERE token_hash = $1;`, tokenHash)
	// err := row.Scan(&user.ID)
	// if err != nil {
	// 	return nil, fmt.Errorf("user: %w", err)
	// }
	// //3. Using the USerID from the session, we need to query for that user
	// row = ss.DB.QueryRow(`
	// 	SELECT email, password_hash
	// 	FROM users
	// 	WHERE id = $1;`, user.ID)
	// err = row.Scan(&user.Email, &user.PasswordHash)
	// if err != nil {
	// 	return nil, fmt.Errorf("user: %w", err)
	// }
	//4. Return the user
	return &user, nil
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)
	_, err := ss.DB.Exec(`
	DELETE FROM sessions
	WHERE token_hash = $1;`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
