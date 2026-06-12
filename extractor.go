package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type GithubContent struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

type BibleTranslation struct {
	Language    string `json:"language"`
	Name        string `json:"name"`
	Format      string `json:"format"`
	DownloadURL string `json:"download_url"`
}

type BibleExtractor struct {
	RepoAPIURL string
}

func NewBibleExtractor() *BibleExtractor {
	return &BibleExtractor{
		RepoAPIURL: "https://api.github.com/repos/seven1m/open-bibles/contents/",
	}
}

// FetchAvailableTranslations hakee Githubista listan ja rakentaa Cliblen katalogin
func (e *BibleExtractor) FetchAvailableTranslations() ([]BibleTranslation, error) {
	resp, err := http.Get(e.RepoAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository contents: %w", err)
	}
	defer resp.Body.Close()

	var ghItems []GithubContent
	if err := json.NewDecoder(resp.Body).Decode(&ghItems); err != nil {
		return nil, fmt.Errorf("failed to decode repository contents: %w", err)
	}

	var translations []BibleTranslation

	for _, item := range ghItems {
		// Huomioidaan vain tiedostot, jotka loppuvat .xml
		if item.Type != "file" || !strings.HasSuffix(item.Name, ".xml") {
			continue
		}

		// Pilkotaan tiedostonnimi, esim. "eng-kjv.osis.xml"
		parts := strings.Split(item.Name, ".")
		if len(parts) < 3 {
			continue
		}

		nameAndLang := parts[0] // "eng-kjv" tai "eng-gb-webbe"
		format := parts[1]      // "osis", "usfx" ...

		// Erotetaan kieli ja käännöksen nimi ensimmäisen viivan (-) kohdalla
		dashIndex := strings.Index(nameAndLang, "-")
		var lang, transName string
		if dashIndex != -1 {
			lang = nameAndLang[:dashIndex]
			transName = nameAndLang[dashIndex+1:]
		} else {
			lang = "unknown"
			transName = nameAndLang
		}

		translations = append(translations, BibleTranslation{
			Language:    lang,
			Name:        strings.ToUpper(transName),
			Format:      format,
			DownloadURL: item.DownloadURL,
		})
	}

	return translations, nil
}

// DownloadTranslation lataa tiedoston suoraan levylle (O(1) muistinkulutus)
func (e *BibleExtractor) DownloadTranslation(downloadURL, localPath string) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
