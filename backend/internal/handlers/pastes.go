package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lingopaste/backend/internal/cache"
	"github.com/lingopaste/backend/internal/db"
	"github.com/lingopaste/backend/internal/middleware"
	"github.com/lingopaste/backend/internal/models"
	"github.com/lingopaste/backend/internal/storage"
	"github.com/lingopaste/backend/internal/translate"
	"github.com/lingopaste/backend/internal/utils"
)

type PasteHandler struct {
	db         *db.DynamoDB
	storage    *storage.S3Storage
	cache      *cache.LRUCache
	translator *translate.OpenAITranslator
	maxLength  int
}

func NewPasteHandler(
	db *db.DynamoDB,
	storage *storage.S3Storage,
	cache *cache.LRUCache,
	translator *translate.OpenAITranslator,
	maxLength int,
) *PasteHandler {
	return &PasteHandler{
		db:         db,
		storage:    storage,
		cache:      cache,
		translator: translator,
		maxLength:  maxLength,
	}
}

func (h *PasteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePasteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	if len(req.Content) > h.maxLength {
		http.Error(w, fmt.Sprintf("Content exceeds maximum length of %d characters", h.maxLength), http.StatusBadRequest)
		return
	}

	if req.Tone == "" {
		req.Tone = "default"
	}

	validTones := map[string]bool{"default": true, "professional": true, "friendly": true, "brusque": true}
	if !validTones[req.Tone] {
		http.Error(w, "Invalid tone. Must be: default, professional, friendly, or brusque", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Generate paste ID
	pasteID, err := utils.GeneratePasteID(8)
	if err != nil {
		log.Printf("Error generating paste ID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Detect original language
	originalLang, err := h.translator.DetectLanguage(ctx, req.Content)
	if err != nil {
		log.Printf("Error detecting language: %v", err)
		http.Error(w, "Failed to detect language", http.StatusInternalServerError)
		return
	}
	originalLang = strings.TrimSpace(strings.ToLower(originalLang))

	// Save original to S3
	if err := h.storage.SaveOriginal(ctx, pasteID, req.Content); err != nil {
		log.Printf("Error saving original to S3: %v", err)
		http.Error(w, "Failed to save paste", http.StatusInternalServerError)
		return
	}

	// Get IP and account info
	ip := middleware.GetIPFromContext(ctx)
	ipHash := utils.HashIP(ip)
	// TODO: Get account ID from JWT when auth is implemented
	accountID := ""

	// Create metadata
	meta := &models.PasteMeta{
		PasteID:               pasteID,
		OriginalLanguage:      originalLang,
		Tone:                  req.Tone,
		CreatorIPHash:         ipHash,
		CreatorAccountID:      accountID,
		CharacterCount:        len(req.Content),
		AvailableTranslations: []string{originalLang},
	}

	if err := h.db.CreatePasteMeta(ctx, meta); err != nil {
		log.Printf("Error saving paste metadata: %v", err)
		http.Error(w, "Failed to save paste metadata", http.StatusInternalServerError)
		return
	}

	// Cache the original
	cacheKey := fmt.Sprintf("%s:%s", pasteID, originalLang)
	h.cache.Set(cacheKey, req.Content)

	// Cache metadata
	h.cache.Set(fmt.Sprintf("meta:%s", pasteID), meta)

	resp := models.CreatePasteResponse{
		PasteID:            pasteID,
		OriginalLanguage:   originalLang,
		AvailableLanguages: []string{originalLang},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PasteHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pasteID := vars["id"]

	if pasteID == "" {
		http.Error(w, "Paste ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Try to get metadata from cache
	var meta *models.PasteMeta
	if cached, ok := h.cache.Get(fmt.Sprintf("meta:%s", pasteID)); ok {
		meta = cached.(*models.PasteMeta)
	} else {
		// Load from database
		var err error
		meta, err = h.db.GetPasteMeta(ctx, pasteID)
		if err != nil {
			log.Printf("Error getting paste metadata: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if meta == nil {
			http.Error(w, "Paste not found", http.StatusNotFound)
			return
		}
		// Cache it
		h.cache.Set(fmt.Sprintf("meta:%s", pasteID), meta)
	}

	// Get original content
	var original string
	cacheKey := fmt.Sprintf("%s:%s", pasteID, meta.OriginalLanguage)
	if cached, ok := h.cache.Get(cacheKey); ok {
		original = cached.(string)
	} else {
		var err error
		original, err = h.storage.GetOriginal(ctx, pasteID)
		if err != nil {
			log.Printf("Error getting original from S3: %v", err)
			http.Error(w, "Failed to load paste", http.StatusInternalServerError)
			return
		}
		h.cache.Set(cacheKey, original)
	}

	// Load all available translations
	translations := make(map[string]string)
	translations[meta.OriginalLanguage] = original

	for _, lang := range meta.AvailableTranslations {
		if lang == meta.OriginalLanguage {
			continue
		}

		langCacheKey := fmt.Sprintf("%s:%s", pasteID, lang)
		if cached, ok := h.cache.Get(langCacheKey); ok {
			translations[lang] = cached.(string)
		} else {
			// Try to load from S3
			trans, err := h.storage.GetTranslation(ctx, pasteID, lang)
			if err == nil {
				translations[lang] = trans
				h.cache.Set(langCacheKey, trans)
			}
		}
	}

	resp := models.GetPasteResponse{
		PasteID:               pasteID,
		OriginalLanguage:      meta.OriginalLanguage,
		Tone:                  meta.Tone,
		CreatedAt:             meta.CreatedAt,
		Original:              original,
		Translations:          translations,
		AvailableTranslations: meta.AvailableTranslations,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PasteHandler) Translate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pasteID := vars["id"]
	targetLang := r.URL.Query().Get("lang")

	if pasteID == "" || targetLang == "" {
		http.Error(w, "Paste ID and language are required", http.StatusBadRequest)
		return
	}

	targetLang = strings.TrimSpace(strings.ToLower(targetLang))

	ctx := r.Context()

	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", pasteID, targetLang)
	if cached, ok := h.cache.Get(cacheKey); ok {
		resp := models.TranslateResponse{
			Language:    targetLang,
			Translation: cached.(string),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Try to load from S3
	translation, err := h.storage.GetTranslation(ctx, pasteID, targetLang)
	if err == nil {
		h.cache.Set(cacheKey, translation)
		resp := models.TranslateResponse{
			Language:    targetLang,
			Translation: translation,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Need to translate - get metadata and original
	meta, err := h.db.GetPasteMeta(ctx, pasteID)
	if err != nil || meta == nil {
		http.Error(w, "Paste not found", http.StatusNotFound)
		return
	}

	original, err := h.storage.GetOriginal(ctx, pasteID)
	if err != nil {
		log.Printf("Error getting original from S3: %v", err)
		http.Error(w, "Failed to load paste", http.StatusInternalServerError)
		return
	}

	// Perform translation
	translation, err = h.translator.Translate(ctx, original, targetLang, meta.Tone)
	if err != nil {
		log.Printf("Error translating: %v", err)
		http.Error(w, "Translation failed", http.StatusInternalServerError)
		return
	}

	// Save translation to S3
	if err := h.storage.SaveTranslation(ctx, pasteID, targetLang, translation); err != nil {
		log.Printf("Error saving translation to S3: %v", err)
		// Continue anyway - we have the translation
	}

	// Update metadata to include new language
	if err := h.db.AddTranslationLanguage(ctx, pasteID, targetLang); err != nil {
		log.Printf("Error updating paste metadata: %v", err)
	}

	// Cache the translation
	h.cache.Set(cacheKey, translation)

	resp := models.TranslateResponse{
		Language:    targetLang,
		Translation: translation,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
