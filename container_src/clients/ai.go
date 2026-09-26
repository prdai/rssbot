package clients

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/luci/go-render/render"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"
	"github.com/prdai/rssbot/services"
	"github.com/prdai/rssbot/utils"
)

const (
	DefaultModelName = "space-bunny-free"
	DefaultBaseURL   = "https://opencode.ai/zen/v1"
	SystemPromptPath = "../prompts/index.j2"
)

type AIClient struct {
	client openai.Client
	model  string
}

func NewAIClient() (*AIClient, error) {
	apiKey := os.Getenv("OPENCODE_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENCODE_API_KEY is not set")
	}
	model := os.Getenv("OPENCODE_MODEL")
	if model == "" {
		model = DefaultModelName
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(DefaultBaseURL),
	)
	return &AIClient{client: client, model: model}, nil
}

type EmailContent struct {
	Title    string `json:"title"`
	HTMLBody string `json:"body"`
}

func (a *AIClient) GenerateEmail(rssFeedsItems []*services.NewItems) (EmailContent, error) {
	input := render.Render(rssFeedsItems) + render.Render(time.Now())
	completion, err := a.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Model: a.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(utils.LoadTemplate(SystemPromptPath)),
			openai.UserMessage(input),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		},
	})
	if err != nil {
		return EmailContent{}, err
	}
	if len(completion.Choices) == 0 {
		return EmailContent{}, errors.New("no completion choices returned")
	}
	rawOutput := strings.TrimSpace(completion.Choices[0].Message.Content)
	if rawOutput == "" {
		return EmailContent{}, errors.New("no content generated")
	}
	var emailContent EmailContent
	dec := json.NewDecoder(strings.NewReader(rawOutput))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&emailContent); err != nil {
		return EmailContent{}, err
	}
	return emailContent, nil
}
