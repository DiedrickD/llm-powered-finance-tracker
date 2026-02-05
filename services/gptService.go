package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	Url:   "https://api.openai.com/v1/responses",
	Model: "gpt-5-nano",
}

func GptService(ctx context.Context, description string, category []string) (int, string, error) {
	// Check for edge cases
	if description == "" {
		return 0, "Uncategorized", errors.New("description is empty")
	}

	apiKey := os.Getenv("OPEN_AI_KEY")
	if apiKey == "" {
		return 0, "Uncategorized", errors.New("OPEN_AI_KEY is empty")
	}

	// Convert into a list
	listCategory := strings.Join(category, ", ")

	// Create payload and convert to json
	payload := map[string]interface{}{
		"model": gptModel.Model,
		"input": "Parse the amount and categorize it based on this list " + listCategory + "from this description:" + description + "return the output in this format [amount(int), category]",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return 0, "Failed json convertion", err
	}

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "POST", gptModel.Url, bytes.NewReader(jsonPayload))
	if err != nil {
		return 0, "Failed request to OPENAI", err
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiKey)

	// Send the payload
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return 0, "Uncategorized", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return 0, "Uncategorized", fmt.Errorf("API error: status code %d", resp.StatusCode)
	}

	var result GptPromptResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, "", err
	}

	var generatedResponse string
	found := false

	// Loop to find output text
	for _, out := range result.Output {
		if out.Type == "message" {
			for _, c := range out.Content {
				if c.Type == "output_text" {
					generatedResponse = c.Text
					found = true
				}
			}
		}
	}

	if !found || generatedResponse == "" {
		return 0, "Uncategorized", errors.New("AI returned an empty response")
	}

	fmt.Println(generatedResponse)
	// Clean response
	generatedResponse = strings.Trim(generatedResponse, "[]")
	parts := strings.Split(generatedResponse, ",")

	valAmount, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		panic(err)
	}

	// parse category
	valCategory := strings.TrimSpace(parts[1])

	// Get and clean the output from GPT response
	return valAmount, valCategory, nil
}
