package jisho

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const JishoUrl = "https://jisho.org/api/v1/search/words?keyword=%s"

type JishoDict struct{}

func (jisho *JishoDict) Search(key string) (JishoResponse, error) {
	var jr JishoResponse

	responses, err := http.Get(fmt.Sprintf(JishoUrl, url.QueryEscape(key)))

	if err != nil {
		return jr, err
	}

	responsesBytes, err := io.ReadAll(responses.Body)

	if err != nil {
		return jr, err
	}

	err = json.Unmarshal(responsesBytes, &jr)

	return jr, err
}
