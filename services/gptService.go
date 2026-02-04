package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type GptModel struct {
	Url    string `json:"url"`
	Model  string `json:"model"`
	ApiKey string
}

type GptPromptResponse struct {
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content,omitempty"`
	} `json:"output,omitempty"`
}

var gptModel = GptModel{
	Url:    "https://api.openai.com/v1/responses",
	Model:  "gpt-5-nano",
	ApiKey: os.Getenv("OPEN_AI_KEY"),
}

func GptService(ctx context.Context, description string) (string, error) {
	// Check for edge cases
	if description == "" {
		return "Uncategorized", errors.New("description is empty")
	}

	// Create payload and convert to json
	payload := map[string]interface{}{
		"model": gptModel.Model,
		// Todo : ubah ini menjadi reusable prompt
		"input": "Categorize this financial statement {" + description + "} in one word",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "Failed json convertion", err
	}

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "POST", gptModel.Url, bytes.NewReader(jsonPayload))
	if err != nil {
		return "Failed request to OPENAI", err
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+gptModel.ApiKey)

	// Send the payload
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "Uncategorized", err
	}

	if resp.StatusCode != http.StatusOK {
		return "Uncategorized", fmt.Errorf("API error: status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var result GptPromptResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	if len(result.Output) == 0 || len(result.Output[0].Content) == 0 {
		return "Uncategorized", errors.New("AI returned an empty response")
	}

	generatedCategory := strings.TrimSpace(result.Output[0].Content[0].Text)

	// Get and clean the output from GPT response
	return generatedCategory, nil
}
