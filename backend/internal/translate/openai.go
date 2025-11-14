package translate

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAITranslator struct {
	client *openai.Client
	model  string
}

func NewOpenAITranslator(apiKey, model string) *OpenAITranslator {
	client := openai.NewClient(apiKey)
	return &OpenAITranslator{
		client: client,
		model:  model,
	}
}

func (t *OpenAITranslator) DetectLanguage(ctx context.Context, text string) (string, error) {
	systemPrompt := "You are a language detection assistant. Respond with ONLY the ISO 639-1 language code (e.g., 'en', 'es', 'fr', 'de', 'ja', 'zh') for the given text. No explanations, just the code."

	resp, err := t.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: t.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		Temperature: 0.1,
		MaxTokens:   10,
	})
	if err != nil {
		return "", fmt.Errorf("failed to detect language: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

func (t *OpenAITranslator) Translate(ctx context.Context, text, targetLanguage, tone string) (string, error) {
	systemPrompt := buildSystemPrompt(targetLanguage, tone)

	resp, err := t.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: t.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		Temperature: 0.3,
	})
	if err != nil {
		return "", fmt.Errorf("failed to translate: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

func buildSystemPrompt(targetLanguage, tone string) string {
	toneInstruction := getToneInstruction(tone)

	return fmt.Sprintf(`You are a professional translator. Translate the following text to %s.

Tone: %s

Important:
- Preserve all formatting (line breaks, spacing, etc.)
- Translate all content accurately
- Maintain the original meaning and context
- Return ONLY the translated text, nothing else

Target language: %s`, getLanguageName(targetLanguage), toneInstruction, targetLanguage)
}

func getToneInstruction(tone string) string {
	switch tone {
	case "professional":
		return "Use formal business language. Be polite, professional, and respectful."
	case "friendly":
		return "Use warm and conversational language. Be approachable and personable."
	case "brusque":
		return "Be direct and concise. Get straight to the point without unnecessary words."
	default:
		return "Use natural and accurate language. Be clear and appropriate for general use."
	}
}

func getLanguageName(code string) string {
	languages := map[string]string{
		"en": "English",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"ru": "Russian",
		"ja": "Japanese",
		"ko": "Korean",
		"zh": "Chinese",
		"ar": "Arabic",
		"hi": "Hindi",
		"nl": "Dutch",
		"pl": "Polish",
		"tr": "Turkish",
		"vi": "Vietnamese",
		"th": "Thai",
		"sv": "Swedish",
		"da": "Danish",
		"fi": "Finnish",
		"no": "Norwegian",
	}

	if name, ok := languages[code]; ok {
		return name
	}
	return code
}
