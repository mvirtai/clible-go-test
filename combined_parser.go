package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

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

// parseUSFX käsittelee litteän milestone-pohjaisen USFX-XML-datan tilakoneella
func (p *CombinedParser) parseUSFX(decoder *xml.Decoder) ([]Verse, error) {
	var verses []Verse

	var currentBook string
	var currentChapter int
	var currentVerse int
	var textBuffer strings.Builder // tehokas tapa kerryttää merkkijonoja Go:ssa

	// Sisäinen apufunktio: tallentaa puskurissa olevan jakeen listaan ja tyhjentää puskurin
	flushCurrentVerse := func() {
		// tallennetaan vain, jos on validi sijainti ja puskurissa tekstiä
		rawText := textBuffer.String()
		trimmedText := strings.TrimSpace(rawText)

		if currentBook != "" && currentChapter > 0 && currentVerse > 0 && trimmedText != "" {
			verses = append(verses, Verse{
				BookID:  currentBook,
				Chapter: currentChapter,
				Verse:   currentVerse,
				Text:    trimmedText,
			})
		}
		// tyhjennetään puskuri seuraavaa jaetta varten
		textBuffer.Reset()
	}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			// kriittinen! kun tiedosto loppuu, muista flushata viimeinen jae
			flushCurrentVerse()
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading USFX token: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			tagName := strings.ToLower(se.Name.Local)

			switch tagName {
			case "book":
				// <book id="GEN">
				flushCurrentVerse()
				currentBook = p.getAttrValue(se.Attr, "id")
				currentChapter = 0
				currentVerse = 0

			case "c":
				// <c id="1"/>
				flushCurrentVerse() // Edellisen luvun viimeinen jae päättyi tähän
				chStr := p.getAttrValue(se.Attr, "id")
				chNum, err := strconv.Atoi(chStr)
				if err != nil {
					return nil, fmt.Errorf("invalid USFX chapter id '%s': %w", chStr, err)
				}
				currentChapter = chNum
				currentVerse = 0

			case "v":
				// <v id="1"/>
				flushCurrentVerse() // Edellinen jae päättyi, uusi alkaa
				vStr := p.getAttrValue(se.Attr, "id")
				vNum, err := strconv.Atoi(vStr)
				if err != nil {
					return nil, fmt.Errorf("invalid USFX verse id '%s': %w", vStr, err)
				}
				currentVerse = vNum
			}

		case xml.CharData:
			// Jos olemme parhaillaan validin jakeen sisällä, kirjoitetaan teksti puskuriin.
			// Tämä poimii tekstin riippumatta siitä, onko välissä muita tageja (kuten <p> tai muotoiluja).
			if currentBook != "" && currentChapter > 0 && currentVerse > 0 {
				textBuffer.Write(se)
			}
		}
	}

	return verses, nil
}

// parseOSIS parsii monimutkaisen OSIS-formaatin (tukee container- ja milestone-tyylejä)
func (p *CombinedParser) parseOSIS(decoder *xml.Decoder) ([]Verse, error) {
	var verses []Verse

	// Tilakoneen muuttujat
	var currentBook string
	var currentChapter int
	var currentVerse int
	var textBuffer strings.Builder

	// Lippulappu, jolla muistetaan ollanko container- vai milestone-tilassa
	isContainerMode := false

	// Sisäinen apufunktio puskurin tyhjentämiseen (flush)
	flushCurrentVerse := func() {
		trimmedText := strings.TrimSpace(textBuffer.String())
		if currentBook != "" && currentChapter > 0 && currentVerse > 0 && trimmedText != "" {
			verses = append(verses, Verse{
				BookID:  currentBook,
				Chapter: currentChapter,
				Verse:   currentVerse,
				Text:    trimmedText,
			})
		}
		textBuffer.Reset()
	}

	// Sisäinen apufunktio OSIS-tunnisteen pilkkomiseen (esim. "Gen.1.1" -> "GEN", 1, 1)
	parseOsisID := func(idStr string) (string, int, int, error) {
		parts := strings.Split(idStr, ".")
		if len(parts) < 3 {
			return "", 0, 0, fmt.Errorf("invalid OSIS ID format '%s'", idStr)
		}

		book := strings.ToUpper(parts[0]) // Muutetaan standardiksi (esim. Gen -> GEN)


		ch, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, 0, fmt.Errorf("invalid chapter in OSIS ID '%s'", idStr, err)
		}

		v, err := strconv.Atoi(parts[2])
		if err != nil {
			return "", 0, 0, fmt.Errorf("invalid verse in OSIS ID '%s': %w", idStr, err)
		}

		return book, ch, v, nil
	}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			flushCurrentVerse() // Viimeinen jae talteen
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading OSIS token: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			tagName := strings.ToLower(se.Name.Local)

			if tagName == "verse" {
				osisID := p.getAttrValue(se.Attr, "osisID")
				sID := p.getAttrValue(se.Attr, "sID")
				eID := p.getAttrValue(se.Attr, "eID")

				if osisID != "" {
					// 1. CONTAINER: <verse osisID="Gen.1.1">
					flushCurrentVerse()
					b, c, v, err := parseOsisID(osisID)
					if err != nil {
						return nil, err
					}
					currentBook = b
					currentChapter = c
					currentVerse = v
					isContainerMode = true

				} else if sID != "" {
					// 2. MILESTONE-START: <verse sID="Gen.1.1" />
					flushCurrentVerse()
					b, c, v, err := parseOsisID(sID)
					if err != nil {
						return nil, err
					}
					currentBook = b
					currentChapter = c
					currentVerse = v
					isContainerMode = false

				} else if eID != "" {
					// 3. MILESTONE-END: <verse eID="Gen.1.1" />
					flushCurrentVerse()
					currentVerse = 0 // Tila nollataan, odotetaan seuraavaa jaetta

				}
			}

		case xml.EndElement:
			tagname := strings.ToLower(se.Name.Local)
			// Jos olimme container-tilassa ja </verse> sulkeutuu, tallennetaan data
			if tagName == "verse" && isContainerMode {
				flushCurrentVerse()
				currentVerse = 0
				isContainerMode = false
			}

		case xml.CharData:
			// Kerätään tekstiä vain jos tilakone on validissa jakeessa
			if currentBook != "" && currentChapter > 0 && currentVerse > 0 {
				textBuffer.Write(se)
			}
		}

	} verses, nil
}

func (p *CombinedParser) parseBeblia(decoder *xml.Decoder) ([]Verse, error) {
	return nil, nil
}


func (p *CombinedParser) getAttrValue(attrs []xml.Attr, name string) string {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}