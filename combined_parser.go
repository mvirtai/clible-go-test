package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Apufunktio attribuutin arvon poimimiseen lennosta
func getAttributeValue(attrs []xml.Attr, key string) string {
	for _, attr := range attrs {
		if attr.Name.Local == key {
			return attr.Value
		}
	}
	return ""
}

type Verse struct {
	BookID  string `json:"book_id"`
	Chapter int    `json:"chapter"`
	Verse   int    `json:"verse"`
	Text    string `json:"text"`
}

type CombinedParser struct{}

// ParseFile avaa tiedoston ja selvittää sen XML-formaatin
func (p *CombinedParser) ParseFile(filePath string) ([]Verse, error) {
	// 1. Avataan tiedosto
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file: %w", err)
	}
	defer file.Close()

	// 2. Luodaan striimi-decoder
	decoder := xml.NewDecoder(file)

	// 3. Etsitään ensimmäinen alkutagi
	var rootName string
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("corrupted XML-file: %w", err)
		}

		if startElem, ok := token.(xml.StartElement); ok {
			rootName = strings.ToLower(startElem.Name.Local)
			break
		}
	}

	// 4. REITITYS: Kutsutaan aliparseria ja PALAUTETAAN sen antama tulos (return)
	switch rootName {
	case "xmlbible":
		fmt.Println("Format recognized: Zefania")
		return p.parseZefania(decoder)

	case "usfx":
		fmt.Println("Format recognized: USFX")
		return p.parseUSFX(decoder)

	case "osis":
		fmt.Println("Format recognized: OSIS")
		return p.parseOSIS(decoder)

	case "bible":
		fmt.Println("Format recognized: Beblia")
		return p.parseBeblia(decoder)

	default:
		return nil, fmt.Errorf("unsupported XML-format: <%s>", rootName)
	}
}

// --- ALIPARSEREIDEN RUNKO-OSAT ---

// parseZefania parseroi puumuotoisen XML-datan ja palauttaa slice-listan jakeita
func (p *CombinedParser) parseZefania(decoder *xml.Decoder) ([]Verse, error) {
	var verses []Verse

	// Tilamuuttujat, joilla muistetaan missä kohdassa tiedostoa ollaan menossa
	var currentBook string
	var currentChapter int

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break // Tiedosto luettu onnistuneesti loppuun
		}
		if err != nil {
			return nil, fmt.Errorf("error reading XML token: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "BIBLEBOOK":
				// Haetaan kirjan tunniste (Zefaniassa usein bname tai bnumber)
				currentBook = getAttributeValue(se.Attr, "bname")
				if currentBook == "" {
					currentBook = getAttributeValue(se.Attr, "bnumber")
				}

			case "CHAPTER":
				chStr := getAttributeValue(se.Attr, "cnumber")
				chNum, err := strconv.Atoi(chStr)
				if err != nil {
					return nil, fmt.Errorf("invalid chapter number '%s': %w", chStr, err)
				}
				currentChapter = chNum

			case "VERS":
				vStr := getAttributeValue(se.Attr, "vnumber")
				vNum, err := strconv.Atoi(vStr)
				if err != nil {
					return nil, fmt.Errorf("invalid verse number '%s': %w", vStr, err)
				}

				// Koska <VERS> on lehtisolmu, napataan sen sisäinen teksti dynaamisesti.
				// DecodeElement purkaa elementin sisällön suoraan muuttujaan ja siirtää decoderin
				// automaattisesti kyseisen elementin sulkutagin yli (</VERS>).
				var verseText string
				if err := decoder.DecodeElement(&verseText, &se); err != nil {
					return nil, fmt.Errorf("failed to decode verse text: %w", err)
				}

				// Luodaan uusi Verse-olio ja liitetään se listaan
				newVerse := Verse{
					BookID:  currentBook, // TODO: Mäppäys "Genesis" -> "GEN" myöhemmin
					Chapter: currentChapter,
					Verse:   vNum,
					Text:    verseText,
				}
				verses = append(verses, newVerse)
			}
		}
	}

	return verses, nil
}

func (p *CombinedParser) parseUSFX(decoder *xml.Decoder) ([]Verse, error) {
	return nil, nil
}

func (p *CombinedParser) parseOSIS(decoder *xml.Decoder) ([]Verse, error) {
	return nil, nil
}

func (p *CombinedParser) parseBeblia(decoder *xml.Decoder) ([]Verse, error) {
	return nil, nil
}
