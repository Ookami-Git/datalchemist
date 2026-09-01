package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"datalchemist/database"

	"github.com/spf13/viper"
	"golang.org/x/crypto/scrypt"
)

// Paramètres scrypt et tailles. Ils font partie du format : les modifier rend
// illisible tout ce qui a déjà été chiffré, en base comme dans une archive.
const (
	scryptN    = 32768
	scryptR    = 8
	scryptP    = 1
	keyLength  = 32
	SaltLength = 32
)

// Key est une clé dérivée prête à l'emploi. La dérivation scrypt coûte une
// centaine de millisecondes : un appelant qui traite plusieurs secrets (export,
// import, migration de passphrase) dérive une fois et réutilise la même Key.
type Key struct {
	aead cipher.AEAD
}

// NewKey dérive une clé à partir d'une passphrase et d'un salt explicites. Le
// salt n'est pas secret, mais il doit accompagner les données qu'il a servi à
// protéger : sans lui, la passphrase seule ne suffit pas à déchiffrer.
func NewKey(passphrase string, salt []byte) (*Key, error) {
	raw, err := deriveKeyWith(passphrase, salt)
	if err != nil {
		return nil, err
	}
	return newKey(raw)
}

// NewSalt génère un salt aléatoire, à stocker en clair auprès des données
// chiffrées correspondantes.
func NewSalt() ([]byte, error) {
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// Encrypt chiffre un texte en clair, retourne une string base64 (nonce + ciphertext)
func (k *Key) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, k.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := k.aead.Seal(nil, nonce, []byte(plaintext), nil)
	result := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt déchiffre une string base64 (nonce + ciphertext), retourne le texte en clair
func (k *Key) Decrypt(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	nonceSize := k.aead.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("données chiffrées trop courtes")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	// GCM est authentifié : une mauvaise passphrase ou un mauvais salt échoue
	// ici plutôt que de rendre un clair erroné.
	plaintext, err := k.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("échec du déchiffrement: %w", err)
	}

	return string(plaintext), nil
}

func newKey(raw []byte) (*Key, error) {
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Key{aead: aead}, nil
}

// deriveKeyWith dérive une clé depuis une passphrase et un salt explicites.
func deriveKeyWith(passphrase string, salt []byte) ([]byte, error) {
	if passphrase == "" {
		return nil, errors.New("secretkey parameter is required")
	}
	if len(salt) == 0 {
		return nil, errors.New("salt is required")
	}
	return scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, keyLength)
}

// deriveKey génère une clé à partir de la passphrase (issue de viper) et d'un salt.
func deriveKey(secretkey string) ([]byte, error) {
	passphrase := secretkey
	if passphrase == "" {
		return nil, errors.New("secretkey parameter is required")
	}
	salt_value, err := database.ParameterGetValue("secret_salt")
	if err != nil {
		return nil, err
	}
	salt := []byte(salt_value.Value)
	return deriveKeyWith(passphrase, salt)
}

// localKey construit la clé de cette instance : la passphrase fournie et le
// salt stocké en base (paramètre secret_salt).
func localKey(secretkey string) (*Key, error) {
	raw, err := deriveKey(secretkey)
	if err != nil {
		return nil, err
	}
	return newKey(raw)
}

// Encrypt chiffre un texte en clair avec la clé de l'instance, retourne une
// string base64 (nonce + ciphertext)
func Encrypt(plaintext string) (string, error) {
	key, err := localKey(viper.GetString("secretkey"))
	if err != nil {
		return "", err
	}
	return key.Encrypt(plaintext)
}

// Decrypt déchiffre une string base64 (nonce + ciphertext) avec le salt de
// l'instance, retourne le texte en clair
func Decrypt(encrypted string, secretkey string) (string, error) {
	key, err := localKey(secretkey)
	if err != nil {
		return "", err
	}
	return key.Decrypt(encrypted)
}
