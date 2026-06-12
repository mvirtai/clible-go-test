package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("=== Clible-v2 Integroitavuustesti ===")
	fmt.Println("-------------------------------------")

	// 1. TESTI: Testataan bible_structure.json ja sen järjestäminen
	fmt.Println("1. Ladataan staattinen Raamatun rakenne...")
	kirjaLista, err := OrderedBookIDs()
	if err != nil {
		log.Fatalf("❌ JSON-luku epäonnistui: %v", err)
	}
	fmt.Printf("   ✓ Metadata toimii! Löytyi %d kirjaa (Ensimmäinen: %s)\n", len(kirjaLista), kirjaLista[0])

	// 2. TESTI: Testataan Extractor (Yhteys GitHubin API:in)
	fmt.Println("\n2. Otetaan yhteys GitHub-luetteloon (Extract)...")
	extractor := NewBibleExtractor()
	käännökset, err := extractor.FetchAvailableTranslations()
	if err != nil {
		log.Fatalf("❌ Käännösten haku epäonnistui: %v", err)
	}

	fmt.Printf("   ✓ Yhteys pelittää! GitHubista löytyi yhteensä %d XML-käännöstä.\n", len(käännökset))

	// Näytetään ensimmäiset 3 käännöstä esimerkkinä siitä, että data pilkkoutuu oikein
	fmt.Println("   Esimerkki luettelon dynaamisesta parsinnasta (Top 3):")
	for i := 0; i < 3 && i < len(käännökset); i++ {
		k := käännökset[i]
		fmt.Printf("     - [%s] %s (Formaatti: %s)\n", k.Language, k.Name, k.Format)
	}

	// 3. TESTI: Testataan CombinedParser lennosta luotavalla dummy-datalla
	fmt.Println("\n3. Testataan CombinedParserin reititystä ja Zefania-aliparseria...")

	// Luodaan pieni validi Zefania-XML-rakenne merkkijonona
	dummyZefaniaXML := `<?xml version="1.0" encoding="utf-8"?>
<xmlbible>
	<BIBLEBOOK bname="GEN">
		<CHAPTER cnumber="1">
			<VERS vnumber="1">Alussa Jumala loi taivaan ja maan.</VERS>
			<VERS vnumber="2">Maa oli autio ja tyhjä.</VERS>
		</CHAPTER>
	</BIBLEBOOK>
</xmlbible>`

	testTiedosto := "testi_asennus_zefania.xml"

	// Kirjoitetaan se hetkeksi levylle testitiedostoksi
	err = os.WriteFile(testTiedosto, []byte(dummyZefaniaXML), 0644)
	if err != nil {
		log.Fatalf("❌ Testitiedoston luonti epäonnistui: %v", err)
	}
	// Varmistetaan, että testitiedosto siivotaan pois levyltä kun ohjelma sulkeutuu
	defer os.Remove(testTiedosto)

	// Kutsutaan parseria
	parser := &CombinedParser{}
	jakeet, err := parser.ParseFile(testTiedosto)
	if err != nil {
		log.Fatalf("❌ Parsinta epäonnistui: %v", err)
	}

	// Tulostetaan lopputulos
	fmt.Printf("   ✓ Parsinta onnistui! Tunnistettiin formaatti ja saatiin %d jaetta:\n", len(jakeet))
	for _, j := range jakeet {
		fmt.Printf("     [%s %d:%d] -> %s\n", j.BookID, j.Chapter, j.Verse, j.Text)
	}

	fmt.Println("\n-------------------------------------")
	fmt.Println("🚀 Kaikki järjestelmät toimivat nimellisesti! Clible-v2 runko on valmis laajennettavaksi.")
}
