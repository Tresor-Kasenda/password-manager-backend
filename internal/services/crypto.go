package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

type CryptoService struct {
	httpClient *http.Client
}

func NewCryptoService() *CryptoService {
	return &CryptoService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *CryptoService) DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		3,       // iterations
		64*1024, // memory (64 MB)
		4,       // parallelism
		32,      // key length (256 bits)
	)
}

func (s *CryptoService) GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func (s *CryptoService) HashPassword(password string) (string, string, error) {
	salt, err := s.GenerateSalt()
	if err != nil {
		return "", "", err
	}

	hash := s.DeriveKey(password, salt)

	return base64.StdEncoding.EncodeToString(hash),
		base64.StdEncoding.EncodeToString(salt),
		nil
}

func (s *CryptoService) VerifyPassword(password, hashStr, saltStr string) bool {
	hash, err := base64.StdEncoding.DecodeString(hashStr)
	if err != nil {
		return false
	}

	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return false
	}

	derivedKey := s.DeriveKey(password, salt)

	return string(hash) == string(derivedKey)
}

func (s *CryptoService) EncryptData(plaintext, masterPassword string) (ciphertext, salt, nonce string, err error) {
	saltBytes, err := s.GenerateSalt()
	if err != nil {
		return "", "", "", err
	}

	key := s.DeriveKey(masterPassword, saltBytes)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", err
	}

	nonceBytes := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonceBytes); err != nil {
		return "", "", "", err
	}

	ciphertextBytes := gcm.Seal(nil, nonceBytes, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertextBytes),
		base64.StdEncoding.EncodeToString(saltBytes),
		base64.StdEncoding.EncodeToString(nonceBytes),
		nil
}

func (s *CryptoService) DecryptData(ciphertextStr, masterPassword, saltStr, nonceStr string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextStr)
	if err != nil {
		return "", err
	}

	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceStr)
	if err != nil {
		return "", err
	}

	key := s.DeriveKey(masterPassword, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: invalid password or corrupted data")
	}

	return string(plaintext), nil
}

func (s *CryptoService) GeneratePassword(length int, useSpecial bool) (string, error) {
	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		digits    = "0123456789"
		special   = "!@#$%^&*()-_=+[]{}|;:,.<>?"
	)

	charset := uppercase + lowercase + digits
	if useSpecial {
		charset += special
	}

	password := make([]byte, length)
	if _, err := rand.Read(password); err != nil {
		return "", err
	}

	for i := range password {
		password[i] = charset[int(password[i])%len(charset)]
	}

	return string(password), nil
}

func (s *CryptoService) CheckPasswordBreach(password string) (breached bool, count int, err error) {
	hasher := sha1.New()
	hasher.Write([]byte(password))
	hashBytes := hasher.Sum(nil)
	hashStr := strings.ToUpper(hex.EncodeToString(hashBytes))

	prefix := hashStr[:5]
	suffix := hashStr[5:]

	url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, err
	}

	req.Header.Set("User-Agent", "PasswordManager-Go")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, 0, fmt.Errorf("HIBP API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, 0, err
	}

	lines := strings.Split(string(body), "\r\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		if parts[0] == suffix {
			count, err := strconv.Atoi(parts[1])
			if err != nil {
				return true, 0, nil
			}
			return true, count, nil
		}
	}

	return false, 0, nil
}

func (s *CryptoService) CalculatePasswordStrength(password string) map[string]interface{} {
	score := 0
	var issues []string
	var suggestions []string

	length := len(password)
	if length < 8 {
		issues = append(issues, "Password is too short (minimum 8 characters)")
		suggestions = append(suggestions, "Use at least 12 characters")
	} else if length < 12 {
		score += 10
		suggestions = append(suggestions, "Consider using 16+ characters")
	} else if length >= 16 {
		score += 30
	} else {
		score += 20
	}

	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
		} else if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= '0' && char <= '9' {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}

	varietyCount := 0
	if hasLower {
		varietyCount++
	} else {
		suggestions = append(suggestions, "Add lowercase letters")
	}

	if hasUpper {
		varietyCount++
	} else {
		suggestions = append(suggestions, "Add uppercase letters")
	}

	if hasDigit {
		varietyCount++
	} else {
		suggestions = append(suggestions, "Add numbers")
	}

	if hasSpecial {
		varietyCount++
	} else {
		suggestions = append(suggestions, "Add special characters")
	}

	score += varietyCount * 15

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	var strength, color string
	if score >= 80 {
		strength = "Strong"
		color = "green"
	} else if score >= 60 {
		strength = "Medium"
		color = "orange"
	} else {
		strength = "Weak"
		color = "red"
	}

	return map[string]interface{}{
		"score":       score,
		"strength":    strength,
		"color":       color,
		"issues":      issues,
		"suggestions": suggestions,
	}
}
