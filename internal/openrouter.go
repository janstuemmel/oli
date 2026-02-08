package internal

import (
	"encoding/json"
	"errors"
	"strings"
)

type Delta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index int   `json:"index"`
	Delta Delta `json:"delta"`
}

type Usage struct {
	Cost float64
}

type Chunk struct {
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage"`
}

func trim(data string) string {
	return strings.TrimSpace(strings.TrimPrefix(data, "data:"))
}

func HandleOpenRouterChunk(chunk string) (string, *Chunk, error) {
	var data *Chunk

	if !strings.HasPrefix(chunk, "data:") || strings.HasSuffix(chunk, "[DONE]") {
		return "", nil, nil
	}

	err := json.Unmarshal([]byte(trim(chunk)), &data)

	if err != nil {
		return "", data, errors.New("cannot unmarshal json")
	}

	if len(data.Choices) == 0 {
		return "", nil, nil
	}

	return data.Choices[0].Delta.Content, data, nil
}
