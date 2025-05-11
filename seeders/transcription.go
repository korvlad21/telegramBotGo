package seeders

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Phonetic struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

type DictionaryEntry struct {
	Word      string     `json:"word"`
	Phonetics []Phonetic `json:"phonetics"`
}

func getTranscription(word string) (string, error) {
	url := fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/en/%s", word)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var entries []DictionaryEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return "", err
	}

	for _, entry := range entries {
		for _, phonetic := range entry.Phonetics {
			if phonetic.Text != "" {
				return phonetic.Text, nil
			}
		}
	}

	return "", fmt.Errorf("no transcription found")
}

func AddTranscription(word string) {
	transcription, err := getTranscription(word)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Transcription for '%s': %s\n", word, transcription)
}
