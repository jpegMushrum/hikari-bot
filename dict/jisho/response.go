package jisho

import "strings"

type Meta struct {
	Status int `json:"status"`
}

type Japanese struct {
	Reading string `json:"reading"`
	Word    string `json:"word"`
}

type Sense struct {
	SpeechParts []string `json:"parts_of_speech"`
	EnglishDef  []string `json:"english_definitions"`
}

type Data struct {
	IsCommon bool       `json:"is_common"`
	Jlpt     []string   `json:"jlpt"`
	Japanese []Japanese `json:"japanese"`
	Senses   []Sense    `json:"senses"`
}

type JishoResponse struct {
	Meta Meta   `json:"meta"`
	Data []Data `json:"data"`
}

func (jsr *JishoResponse) HasEntries() bool {
	return len(jsr.Data) > 0
}

// Unsafe
func (jsr *JishoResponse) RelevantDefinition() string {
	return jsr.Data[0].Senses[0].EnglishDef[0]
}

// Unsafe
func (jsr *JishoResponse) RelevantKana() string {
	return jsr.Data[0].Japanese[0].Reading
}

// Unsafe
func (jsr *JishoResponse) RelevantWord() string {
	word := jsr.Data[0].Japanese[0].Word
	if word == "" { // Kana only case
		return jsr.Data[0].Japanese[0].Reading
	}
	return word
}

// Unsafe
func (jsr *JishoResponse) RelevantSpeechPart() string {
	return strings.ToLower(jsr.Data[0].Senses[0].SpeechParts[0])
}

// Unsafe
func (jsr *JishoResponse) Words() []string {
	words := []string{}
	for _, dt := range jsr.Data {
		for _, jp := range dt.Japanese {
			if jp.Word == "" { // Kana only
				words = append(words, jp.Reading)
			} else {
				words = append(words, jp.Word)
			}
		}
	}
	return words
}

func (jsr *JishoResponse) Kanas() []string {
	readings := []string{}
	for _, dt := range jsr.Data {
		for _, jp := range dt.Japanese {
			readings = append(readings, jp.Reading)
		}
	}
	return readings
}
