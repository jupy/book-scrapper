package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	/* 	"os/exec" */

	"github.com/bregydoc/gtranslate"
)

/* var Translations = map[string]string{} */

type TagTranslator struct {
	data map[string]string
}

func (translator *TagTranslator) Load() {
	translator.data = make(map[string]string)

	content, err := ioutil.ReadFile("translations.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &translator.data)
	if err != nil {
		log.Fatal(err)
	}
}

func (translator *TagTranslator) Save() {
	file, _ := json.MarshalIndent(translator.data, "", " ")
	_ = ioutil.WriteFile("translations.json", file, 0644)
}

func (translator *TagTranslator) Translate(query string) string {
	trans := translator.data[query]
	if len(trans) > 0 {
		return trans
	}

	translated, err := gtranslate.TranslateWithParams(
		query,
		gtranslate.TranslationParams{
			From: "ru",
			To:   "en",
		},
	)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	fmt.Printf("%s\n", string(translated))
	translator.data[query] = translated
	return translated
}
