package main

import (
	"log"
	"os"
)

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	if len(os.Args) > 4 {
		log.Fatalln("Usage: DolTranslator [path/to/dol] [path/to/json] [new address]")
	}

	dolPath := os.Args[1]
	dol, err := os.ReadFile(dolPath)
	checkError(err)

	translationsPath := os.Args[2]
	trans, err := os.ReadFile(translationsPath)
	checkError(err)

	newAddress := os.Args[3]
	_address, err := ParseAddress(newAddress)
	checkError(err)

	ctx, err := NewTranslateCtx(dol, trans, _address)
	checkError(err)

	err = ctx.ApplyTranslations()
	checkError(err)

	err = ctx.WriteTranslatedDol()
	checkError(err)
}
