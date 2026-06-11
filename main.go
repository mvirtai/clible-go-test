package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("1. Kutsutaan OrderedBookIDs()-funktiota toisesta tiedostosta...")

	// Koska molemmat tiedostot ovat samaa 'package main' -pakettia,
	// main.go näkee OrderedBookIDs-funktion suoraan ilman mitään import-kikkailuja!
	kirjaLista, err := OrderedBookIDs()
	if err != nil {
		log.Fatalf("Hups, JSON-luku epäonnistui: %v", err)
	}

	fmt.Println("\n2. Onnistui! Tässä kaikki Cliblen 66 kirjaa oikeassa järjestyksessä:")
	fmt.Printf("%v\n", kirjaLista)
}
