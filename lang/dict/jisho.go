package dict

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const JishoURL = "https://jisho.org/api/v1/search/words?keyword=%s"

type Meta struct {
	Status int `json:"status"`
}

type Japanese struct {
	Reading string `json:"reading"`
	Word    string `json:"word"`
}

type Data struct {
	IsCommon bool       `json:"is_common"`
	Jlpt     []string   `json:"jlpt"`
	Japanese []Japanese `json:"japanese"`
}

type JishoResponse struct {
	Meta Meta   `json:"meta"`
	Data []Data `json:"data"`
}

// Also need to add Get part_of_spech for Noun Check!!!
func (jr *JishoResponse) GetKana() string {
	// Crack
	// Also suppose that if has Data then has at least one Reading
	if len(jr.Data) == 0 {
		return ""
	}
	return jr.Data[0].Japanese[0].Reading
}

type JishoDict struct{}

func (jisho *JishoDict) Search(key string) (JishoResponse, error) {
	var jr JishoResponse

	responses, err := http.Get(fmt.Sprintf(JishoURL, url.QueryEscape(key)))

	if err != nil {
		return jr, err
	}

	responsesBytes, err := io.ReadAll(responses.Body)

	if err != nil {
		return jr, err
	}

	// fmt.Println(string(responsesBytes))

	json.Unmarshal(responsesBytes, &jr)

	// fmt.Println(jr.Data[0].Japanese[0].Reading)
	return jr, err
}
