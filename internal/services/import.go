package services

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/tresor/password-manager/internal/models"
)

type ImportService struct{}

func NewImportService() *ImportService {
	return &ImportService{}
}

func (s *ImportService) ParseImportFile(content, source string) ([]models.ImportEntry, error) {
	switch source {
	case "1password":
		return s.parseOnePassword(content)
	case "lastpass":
		return s.parseLastPass(content)
	case "bitwarden":
		return s.parseBitwarden(content)
	case "chrome":
		return s.parseChrome(content)
	case "keepass":
		return s.parseKeePass(content)
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}
}

func (s *ImportService) parseOnePassword(content string) ([]models.ImportEntry, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("empty import file")
	}

	// Expected format: Title,Website,Username,Password,Notes,Type,Favorite
	entries := []models.ImportEntry{}

	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) < 7 {
			continue
		}

		website := stringOrNil(record[1])
		username := stringOrNil(record[2])
		notes := stringOrNil(record[4])
		folder := stringOrNil(record[5])

		entry := models.ImportEntry{
			Title:    record[0],
			Website:  website,
			Username: username,
			Password: record[3],
			Notes:    notes,
			Folder:   folder,
			Favorite: strings.ToLower(record[6]) == "true",
			Source:   "1Password",
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *ImportService) parseLastPass(content string) ([]models.ImportEntry, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("empty import file")
	}

	// Expected format: url,username,password,extra,name,grouping,fav
	entries := []models.ImportEntry{}

	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) < 7 {
			continue
		}

		website := stringOrNil(record[0])
		username := stringOrNil(record[1])
		notes := stringOrNil(record[3])
		folder := stringOrNil(record[5])

		entry := models.ImportEntry{
			Title:    record[4],
			Website:  website,
			Username: username,
			Password: record[2],
			Notes:    notes,
			Folder:   folder,
			Favorite: record[6] == "1",
			Source:   "LastPass",
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *ImportService) parseBitwarden(content string) ([]models.ImportEntry, error) {
	var data struct {
		Items []struct {
			Type     int    `json:"type"`
			Name     string `json:"name"`
			Notes    string `json:"notes"`
			Favorite bool   `json:"favorite"`
			Login    struct {
				Username string `json:"username"`
				Password string `json:"password"`
				URIs     []struct {
					URI string `json:"uri"`
				} `json:"uris"`
			} `json:"login"`
			FolderID string `json:"folderId"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, err
	}

	entries := []models.ImportEntry{}

	for _, item := range data.Items {
		if item.Type != 1 { // Type 1 is Login
			continue
		}

		var website *string
		if len(item.Login.URIs) > 0 {
			website = &item.Login.URIs[0].URI
		}

		username := stringOrNil(item.Login.Username)
		notes := stringOrNil(item.Notes)
		folder := stringOrNil(item.FolderID)

		entry := models.ImportEntry{
			Title:    item.Name,
			Website:  website,
			Username: username,
			Password: item.Login.Password,
			Notes:    notes,
			Folder:   folder,
			Favorite: item.Favorite,
			Source:   "Bitwarden",
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *ImportService) parseChrome(content string) ([]models.ImportEntry, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("empty import file")
	}

	// Expected format: name,url,username,password
	entries := []models.ImportEntry{}

	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) < 4 {
			continue
		}

		website := stringOrNil(record[1])
		username := stringOrNil(record[2])

		entry := models.ImportEntry{
			Title:    record[0],
			Website:  website,
			Username: username,
			Password: record[3],
			Folder:   stringPtr("Browser Passwords"),
			Favorite: false,
			Source:   "Chrome",
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *ImportService) parseKeePass(content string) ([]models.ImportEntry, error) {
	type KeePassEntry struct {
		String []struct {
			Key   string `xml:"Key"`
			Value string `xml:"Value"`
		} `xml:"String"`
	}

	type KeePassXML struct {
		Entries []KeePassEntry `xml:"Root>Group>Entry"`
	}

	var data KeePassXML
	if err := xml.Unmarshal([]byte(content), &data); err != nil {
		return nil, err
	}

	entries := []models.ImportEntry{}

	for _, kpEntry := range data.Entries {
		entry := models.ImportEntry{
			Source: "KeePass",
		}

		for _, str := range kpEntry.String {
			switch str.Key {
			case "Title":
				entry.Title = str.Value
			case "URL":
				if str.Value != "" {
					entry.Website = &str.Value
				}
			case "UserName":
				if str.Value != "" {
					entry.Username = &str.Value
				}
			case "Password":
				entry.Password = str.Value
			case "Notes":
				if str.Value != "" {
					entry.Notes = &str.Value
				}
			}
		}

		if entry.Title != "" || entry.Password != "" {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func stringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringPtr(s string) *string {
	return &s
}
