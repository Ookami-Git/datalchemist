package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

// deriveKey génère une clé à partir de la passphrase (issue de viper) et d'un salt fixe.
func deriveKey(secretkey string) ([]byte, error) {
	passphrase := secretkey
	if passphrase == "" {
		return nil, errors.New("secretkey parameter is required")
	}
	salt := []byte("datalchemist_salt_2024")
	return scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, 32)
}

// Encrypt chiffre un texte en clair, retourne une string base64 (nonce + ciphertext)
func Encrypt(plaintext string, secretkey string) (string, error) {
	key, err := deriveKey(secretkey)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)
	result := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt déchiffre une string base64 (nonce + ciphertext), retourne le texte en clair
func Decrypt(encrypted string, secretkey string) (string, error) {
	key, err := deriveKey(secretkey)
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("données chiffrées trop courtes")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("échec du déchiffrement: %w", err)
	}

	return string(plaintext), nil
}