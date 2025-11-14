package models

type Account struct {
	AccountID            string `json:"account_id" dynamodbav:"account_id"`
	Email                string `json:"email" dynamodbav:"email"`
	AuthProvider         string `json:"auth_provider" dynamodbav:"auth_provider"`
	IsPaid               bool   `json:"is_paid" dynamodbav:"is_paid"`
	StripeCustomerID     string `json:"stripe_customer_id,omitempty" dynamodbav:"stripe_customer_id,omitempty"`
	StripeSubscriptionID string `json:"stripe_subscription_id,omitempty" dynamodbav:"stripe_subscription_id,omitempty"`
	CreatedAt            int64  `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt            int64  `json:"updated_at" dynamodbav:"updated_at"`
}

type PasteMeta struct {
	PasteID               string   `json:"paste_id" dynamodbav:"paste_id"`
	OriginalLanguage      string   `json:"original_language" dynamodbav:"original_language"`
	Tone                  string   `json:"tone" dynamodbav:"tone"`
	CreatorIPHash         string   `json:"creator_ip_hash" dynamodbav:"creator_ip_hash"`
	CreatorAccountID      string   `json:"creator_account_id,omitempty" dynamodbav:"creator_account_id,omitempty"`
	CreatedAt             int64    `json:"created_at" dynamodbav:"created_at"`
	CharacterCount        int      `json:"character_count" dynamodbav:"character_count"`
	AvailableTranslations []string `json:"available_translations" dynamodbav:"available_translations"`
}

type RateLimit struct {
	Identifier string `json:"identifier" dynamodbav:"identifier"`
	Date       string `json:"date" dynamodbav:"date"`
	PasteCount int    `json:"paste_count" dynamodbav:"paste_count"`
	LimitType  string `json:"limit_type" dynamodbav:"limit_type"`
	TTL        int64  `json:"ttl" dynamodbav:"ttl"`
}

type Paste struct {
	Meta         PasteMeta
	Original     string
	Translations map[string]string
}

type CreatePasteRequest struct {
	Content string `json:"content"`
	Tone    string `json:"tone"`
}

type CreatePasteResponse struct {
	PasteID            string   `json:"paste_id"`
	OriginalLanguage   string   `json:"original_language"`
	AvailableLanguages []string `json:"available_languages"`
}

type GetPasteResponse struct {
	PasteID               string            `json:"paste_id"`
	OriginalLanguage      string            `json:"original_language"`
	Tone                  string            `json:"tone"`
	CreatedAt             int64             `json:"created_at"`
	Original              string            `json:"original"`
	Translations          map[string]string `json:"translations"`
	AvailableTranslations []string          `json:"available_translations"`
}

type TranslateRequest struct {
	Language string `json:"language"`
}

type TranslateResponse struct {
	Language    string `json:"language"`
	Translation string `json:"translation"`
}
