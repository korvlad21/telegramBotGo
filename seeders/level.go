package seeders

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	oxfordAppID  = os.Getenv("OXFORD_APP_ID")
	oxfordAppKey = os.Getenv("OXFORD_APP_KEY")
)

// GetWordLevel — основная функция получения уровня по слову
func GetWordLevel(word string) (string, error) {
	word = strings.ToLower(strings.TrimSpace(word))
	level, err := fetchLevelFromOxfordAPI(word)
	if err != nil {
		log.Fatal(err)
		return "Unknown", nil
	}
	return level, nil
}

// fetchLevelFromOxfordAPI — делает реальный запрос к Oxford API
func fetchLevelFromOxfordAPI(word string) (string, error) {
	url := fmt.Sprintf("https://od-api.oxforddictionaries.com/api/v2/entries/en-gb/%s", word)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("app_id", oxfordAppID)
	req.Header.Set("app_key", oxfordAppKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New("Oxford API returned non-200 status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return extractLevelFromOxfordJSON(body)
}

// extractLevelFromOxfordJSON — ищет CEFR level в ответе JSON
func extractLevelFromOxfordJSON(body []byte) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	results, ok := data["results"].([]interface{})
	if !ok || len(results) == 0 {
		return "", errors.New("no results in Oxford response")
	}

	for _, r := range results {
		result := r.(map[string]interface{})
		lexicalEntries, ok := result["lexicalEntries"].([]interface{})
		if !ok {
			continue
		}

		for _, entry := range lexicalEntries {
			entryMap := entry.(map[string]interface{})
			if cefrLevel, ok := entryMap["cefrLevel"].(string); ok {
				return cefrLevel, nil
			}
			// иногда уровень может быть в nested "entries" → "senses"
			if entries, ok := entryMap["entries"].([]interface{}); ok {
				for _, e := range entries {
					em := e.(map[string]interface{})
					if senses, ok := em["senses"].([]interface{}); ok {
						for _, s := range senses {
							sm := s.(map[string]interface{})
							if cefrLevel, ok := sm["cefrLevel"].(string); ok {
								return cefrLevel, nil
							}
						}
					}
				}
			}
		}
	}
	return "", errors.New("CEFR level not found in response")
}
