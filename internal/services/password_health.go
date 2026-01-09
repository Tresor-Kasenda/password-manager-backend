package services

import (
	"math"
	"regexp"
	"time"
	"unicode"

	"github.com/tresor/password-manager/internal/models"
)

type PasswordHealthService struct{}

func NewPasswordHealthService() *PasswordHealthService {
	return &PasswordHealthService{}
}

func (s *PasswordHealthService) GenerateHealthReport(vaults []models.Vault) map[string]interface{} {
	totalPasswords := len(vaults)

	excellentCount := 0
	goodCount := 0
	weakCount := 0
	breachedCount := 0
	reusedCount := 0
	oldCount := 0

	details := []map[string]interface{}{}
	priorityActions := []map[string]interface{}{}

	// Simplified analysis (in production, decrypt and analyze properly)
	for _, vault := range vaults {
		score := 75 // Dummy score
		ageDays := int(time.Since(vault.UpdatedAt).Hours() / 24)

		detail := map[string]interface{}{
			"vault_id":        vault.ID,
			"title":           vault.Title,
			"website":         vault.Website,
			"score":           score,
			"strength":        "Good",
			"age_days":        ageDays,
			"is_reused":       false,
			"is_breached":     false,
			"issues":          []string{},
			"recommendations": []string{},
		}

		if score >= 80 {
			excellentCount++
		} else if score >= 60 {
			goodCount++
		} else {
			weakCount++
		}

		if ageDays > 180 {
			oldCount++
		}

		details = append(details, detail)
	}

	averageScore := 75
	if totalPasswords > 0 {
		// Calculate actual average
	}

	return map[string]interface{}{
		"overall_score":   averageScore,
		"total_passwords": totalPasswords,
		"statistics": map[string]interface{}{
			"excellent_passwords": excellentCount,
			"good_passwords":      goodCount,
			"weak_passwords":      weakCount,
			"breached_passwords":  breachedCount,
			"reused_passwords":    reusedCount,
			"old_passwords":       oldCount,
		},
		"details":          details,
		"priority_actions": priorityActions,
	}
}

func (s *PasswordHealthService) CalculateStrength(password string, lastChanged time.Time) map[string]interface{} {
	score := 0
	issues := []string{}
	suggestions := []string{}

	// Length check
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

	// Character variety
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsDigit(char) {
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

	// Pattern detection
	if matched, _ := regexp.MatchString(`(.)\1{2,}`, password); matched {
		issues = append(issues, "Avoid repeated characters")
		score -= 10
	}

	if matched, _ := regexp.MatchString(`(012|123|234|345|456|567|678|789|890)`, password); matched {
		issues = append(issues, "Avoid sequential numbers")
		score -= 10
	}

	// Entropy calculation
	entropy := s.calculateEntropy(password)
	if entropy < 28 {
		issues = append(issues, "Low entropy")
		score -= 15
	}

	// Age check
	ageDays := int(time.Since(lastChanged).Hours() / 24)
	if ageDays > 365 {
		issues = append(issues, "Password is over 1 year old")
		score -= 15
	} else if ageDays > 180 {
		issues = append(issues, "Password is over 6 months old")
		score -= 10
	}

	// Normalize score
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	// Determine strength
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

func (s *PasswordHealthService) calculateEntropy(password string) float64 {
	if len(password) == 0 {
		return 0
	}

	freq := make(map[rune]int)
	for _, char := range password {
		freq[char]++
	}

	length := float64(len(password))
	entropy := 0.0

	for _, count := range freq {
		probability := float64(count) / length
		entropy -= probability * math.Log2(probability)
	}

	return entropy * length
}
