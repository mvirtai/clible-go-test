package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

// BookMetadata vastaa yhtä kirjaa JSON-tiedostossa
type BookMetadata struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"` // int, koska JSONissa tämä on numero!
}

type BibleStructure struct {
	Books []BookMetadata `json:"books"`
}

//go:embed bible_structure.json
var bibleStructureRaw []byte

// OrderedBookIDs palauttaa kirjakoodit kanonisessa järjestyksessä
func OrderedBookIDs() ([]string, error) {
	var structure BibleStructure

	// Puretaan JSON-tavut structiin
	err := json.Unmarshal(bibleStructureRaw, &structure)
	if err != nil {
		return nil, fmt.Errorf("JSON-parse failed: %w", err)
	}

	// Järjestetään kirjat position-numeron mukaan
	sort.Slice(structure.Books, func(i, j int) bool {
		return structure.Books[i].Position < structure.Books[j].Position
	})

	// Kerätään vain ID:t listaksi
	ids := make([]string, len(structure.Books))
	for i, book := range structure.Books {
		ids[i] = book.ID
	}

	return ids, nil
}
