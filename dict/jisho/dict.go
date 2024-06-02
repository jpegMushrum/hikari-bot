package jisho

import (
	"bakalover/hikari-bot/dict"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	JishoUrl   = "https://jisho.org/api/v1/search/words?keyword=%s"
	MaxRetries = 3
	RetryDelay = 2 * time.Second
)

type Jisho struct{}

func NewJisho() *Jisho {
	return &Jisho{}
}

func (jisho *Jisho) Search(key string) (dict.Response, error) {
	var jr JishoResponse
	var err error

	for i := 0; i < MaxRetries; i++ {
		jr, err = jisho.searchAttempt(key)
		if err == nil {
			return &jr, nil
		}

		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			log.Printf("Timeout occurred, retrying (attempt %d/%d)...\n", i+1, MaxRetries)
			time.Sleep(RetryDelay)
			continue
		}

		return &jr, err
	}

	return &jr, fmt.Errorf("max retries exceeded: %w", err)
}

func (jisho *Jisho) searchAttempt(key string) (JishoResponse, error) {
	var jr JishoResponse

	responses, err := http.Get(fmt.Sprintf(JishoUrl, url.QueryEscape(key)))
	if err != nil {
		return jr, err
	}
	defer responses.Body.Close()

	responsesBytes, err := io.ReadAll(responses.Body)
	if err != nil {
		return jr, err
	}

	err = json.Unmarshal(responsesBytes, &jr)
	return jr, err
}

func (j *Jisho) NounRepr() string {
	return "Noun"
}

func (j *Jisho) Repr() string {
	return "Jisho"
}
