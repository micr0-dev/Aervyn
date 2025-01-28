package models

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          string
	Username    string
	DisplayName string
	Bio         string
	Password    string
	PublicKey   string
	PrivateKey  string
	CreatedAt   time.Time
}

func CreateUser(username, password string) (*User, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	publicKeyPEM := &bytes.Buffer{}
	pem.Encode(publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	privateKeyPEM := &bytes.Buffer{}
	pem.Encode(privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = db.Exec(
		"INSERT INTO users (id, username, password, public_key, private_key) VALUES (?, ?, ?, ?, ?)",
		id, username, string(hashedPassword), publicKeyPEM.String(), privateKeyPEM.String(),
	)
	if err != nil {
		return nil, err
	}

	return &User{ID: id, Username: username}, nil
}

func (u *User) UpdateProfile(displayName, bio string) error {
	_, err := db.Exec(`
        UPDATE users 
        SET display_name = ?, bio = ?
        WHERE id = ?
    `, displayName, bio, u.ID)
	return err
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.QueryRow(
		"SELECT id, username, password FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id string) (*User, error) {
	var user User
	err := db.QueryRow(
		"SELECT id, username, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
