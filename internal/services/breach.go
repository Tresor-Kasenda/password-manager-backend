package services

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BreachService struct {
	apiKey     string
	httpClient *http.Client
}

func NewBreachService(apiKey string) *BreachService {
	return &BreachService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *BreachService) CheckBreach(password string) (breached bool, count int, err error) {
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

func (s *BreachService) CheckEmailBreaches(email string) ([]map[string]interface{}, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("HIBP API key not configured")
	}

	url := fmt.Sprintf("https://haveibeenpwned.com/api/v3/breachedaccount/%s", email)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("hibp-api-key", s.apiKey)
	req.Header.Set("User-Agent", "PasswordManager-Go")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []map[string]interface{}{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HIBP API returned status %d", resp.StatusCode)
	}

	// Parse JSON response
	// Implementation depends on your JSON library

	return []map[string]interface{}{}, nil
}
