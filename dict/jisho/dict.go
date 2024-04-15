package jisho

import (
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

type JishoDict struct{}

func (jisho *JishoDict) Search(key string) (JishoResponse, error) {
	var jr JishoResponse
	var err error

	for i := 0; i < MaxRetries; i++ {
		jr, err = jisho.searchAttempt(key)
		if err == nil {
			return jr, nil // Successful response
		}

		// Check if the error is a timeout error
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			log.Printf("Timeout occurred, retrying (attempt %d/%d)...\n", i+1, MaxRetries)
			time.Sleep(RetryDelay) 
			continue
		}

		return jr, err
	}

	return jr, fmt.Errorf("max retries exceeded: %w", err)
}

func (jisho *JishoDict) searchAttempt(key string) (JishoResponse, error) {
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
