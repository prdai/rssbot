package clients

import (
	"context"
	"log/slog"

	"google.golang.org/genai"
)

func NewGeminiClient(apiKey string) (genai.Client, error) {
	client, err := genai.NewClient(context.TODO(), &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		slog.Error(err.Error())
		return genai.Client{}, err
	}
	return *client, nil
}
