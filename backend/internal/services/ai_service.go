package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AIService struct {
	apiKey string
	client *http.Client
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

type EventExtraction struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

func NewAIService(apiKey string) *AIService {
	return &AIService{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *AIService) Research(query string) (string, error) {
	if s.apiKey == "" {
		return "", errors.New("Gemini API key not configured")
	}

	prompt := fmt.Sprintf(`You are a helpful AI assistant in a chat application. 
Please provide a clear, concise, and informative response to the following query:

%s

Format your response in a way that's easy to read and understand.`, query)

	return s.callGemini(prompt)
}

func (s *AIService) ExtractEvent(messageText string) (*EventExtraction, error) {
	if s.apiKey == "" {
		return nil, errors.New("Gemini API key not configured")
	}

	prompt := fmt.Sprintf(`Extract event information from the following text and return ONLY a valid JSON object with these fields:
- title: event name or description
- date: date in YYYY-MM-DD format (use context clues for year if not specified, default to current/next year)
- time: time in HH:MM format (24-hour), or "00:00" if not specified
- location: location or "Not specified"
- description: brief description or empty string

Text: "%s"

Return ONLY the JSON object, no other text.

Example output:
{"title":"Team Meeting","date":"2026-02-15","time":"14:00","location":"Conference Room A","description":"Weekly team sync"}`, messageText)

	response, err := s.callGemini(prompt)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var event EventExtraction
	if err := json.Unmarshal([]byte(response), &event); err != nil {
		// Try to clean the response
		response = cleanJSONResponse(response)
		if err := json.Unmarshal([]byte(response), &event); err != nil {
			return nil, fmt.Errorf("failed to parse event data: %w", err)
		}
	}

	return &event, nil
}

func (s *AIService) callGemini(prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", s.apiKey)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error: %s - %s", resp.Status, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func cleanJSONResponse(response string) string {
	// Remove markdown code blocks if present
	response = bytes.TrimPrefix([]byte(response), []byte("```json"))
	response = bytes.TrimPrefix(response, []byte("```"))
	response = bytes.TrimSuffix(response, []byte("```"))
	return string(bytes.TrimSpace(response))
}
