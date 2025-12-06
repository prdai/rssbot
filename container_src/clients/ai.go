package clients

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/luci/go-render/render"
	"github.com/prdai/rssbot/services"
	"github.com/prdai/rssbot/utils"
	"google.golang.org/genai"
)

const (
	ModelName        = "gemini-2.5-flash"
	SystemPromptPath = "../prompts/index.j2"
)

type AIClient struct {
	GeminiClient     *genai.Client
	Chat             *genai.Chat
	GenerativeConfig *genai.GenerateContentConfig
}

func NewAIClient() (*AIClient, error) {
	client, err := genai.NewClient(context.TODO(), &genai.ClientConfig{APIKey: os.Getenv("GOOGLE_API_KEY"), Backend: genai.BackendGeminiAPI})
	if err != nil {
		slog.Error(err.Error())
		return &AIClient{}, err
	}
	aiClient := &AIClient{GeminiClient: client}
	go aiClient.createChat()
	return aiClient, nil
}

func i64(v int64) *int64 { return &v }

func RSSBotTitleBodySchema() *genai.Schema {
	titlePattern := `^\[[^\[\]]{1,60}\] \| RSSBot Sync \[(\d{4}-\d{2}-\d{2})(T\d{2}:\d{2}(:\d{2})?(Z|[+\-]\d{2}:?\d{2})?)?\]$`

	return &genai.Schema{
		Title:       "RSSBotSyncTitleBody",
		Type:        genai.TypeObject,
		Description: "LLM response for an RSS sync digest. Only title + body (HTML).",
		PropertyOrdering: []string{
			"title", "body",
		},
		Required: []string{"title", "body"},
		Properties: map[string]*genai.Schema{
			"title": {
				Type:        genai.TypeString,
				Description: "Must be formatted exactly as: [x] | RSSBot Sync [date]. x is a short generated label. date is YYYY-MM-DD or full ISO datetime.",
				Pattern:     titlePattern,
				MinLength:   i64(16),
				MaxLength:   i64(120),
			},
			"body": {
				Type:        genai.TypeString,
				Description: "A minimal, good-looking single-column HTML digest. Inline CSS only. Include feed names and new items with links. No external assets required.",
				MinLength:   i64(200),
				MaxLength:   i64(200000),
			},
		},
	}
}

func (a *AIClient) createChat() {
	client := *a
	sysPrompt := &genai.Content{
		Parts: []*genai.Part{
			{Text: utils.LoadTemplate(SystemPromptPath)},
		},
	}
	schema := RSSBotTitleBodySchema()
	generativeConfig := &genai.GenerateContentConfig{SystemInstruction: sysPrompt, ResponseSchema: schema, ResponseMIMEType: "application/json"}
	a.GenerativeConfig = generativeConfig
	chat, err := client.GeminiClient.Chats.Create(context.TODO(), ModelName, generativeConfig, nil)
	if err != nil {
		slog.Error(err.Error())
	}
	a.Chat = chat
}

type EmailContent struct {
	Title    string `json:"title"`
	HTMLBody string `json:"body"`
}

func (a *AIClient) GenerateEmail(rssFeedsItems []*services.NewItems) (EmailContent, error) {
	stringRssFeedsItems := render.Render(rssFeedsItems)
	stringRssFeedsItems += render.Render(time.Now())
	contents := []*genai.Content{
		genai.NewContentFromText(stringRssFeedsItems, genai.RoleUser),
	}
	response, err := a.Chat.GenerateContent(context.TODO(), ModelName, contents, a.GenerativeConfig)
	if err != nil {
		slog.Error(err.Error())
		return EmailContent{}, err
	}
	var rawOutput strings.Builder
	for _, p := range response.Candidates[0].Content.Parts {
		if p != nil && p.Text != "" {
			rawOutput.WriteString(p.Text)
		}
	}
	cleanedRawOutput := strings.TrimSpace(rawOutput.String())
	if cleanedRawOutput == "" {
		return EmailContent{}, errors.New("no content generated")
	}
	var emailContent EmailContent
	dec := json.NewDecoder(strings.NewReader(cleanedRawOutput))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&emailContent); err != nil {
		slog.Error(err.Error())
		return EmailContent{}, err
	}
	return emailContent, nil
}
