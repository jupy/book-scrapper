package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
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

	out, err := exec.Command("translate", "ru", "en", query).Output()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	fmt.Printf("%s\n", string(out))

	vec := strings.Split(string(out), "\n")
	for _, line := range vec {
		if strings.HasPrefix(line, "en: ") {
			trans = strings.TrimPrefix(line, "en: ")
			translator.data[query] = trans
			return trans
		}
	}
	return ""
}
