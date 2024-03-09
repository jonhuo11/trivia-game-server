package main

import (
	"encoding/json"
	"os"
)

type TriviaQuestion struct {
	Q string         `json:"q"`
	A []TriviaAnswer `json:"a"`
}

type TriviaAnswer struct {
	A       string `json:"a"`
	Correct bool   `json:"-"` // dont serialize this field
}

type TriviaBank struct {
	Questions []TriviaQuestion `json:"questions"`
}

func loadTriviaBankFromString(s string) (*TriviaBank, error) {
	tb := TriviaBank{}
	if err := json.Unmarshal([]byte(s), &tb); err != nil {
		return nil, err
	}
	return &tb, nil
}

func loadTriviaBankDefault() (TriviaBank, error) {
	// TODO make this load dynamically
	trivraw, err := os.ReadFile("./assets/trivia/lol.trivia")
	if err != nil {
		return TriviaBank{}, err
	}
	triv, err := loadTriviaBankFromString(string(trivraw))
	if err != nil {
		return TriviaBank{}, err
	}
	return *triv, nil
}
