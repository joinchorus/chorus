package translation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/gitstore"
)

// TranslationRecord holds cached translation data.
type TranslationRecord struct {
	MessageID      string `json:"message_id"`
	TargetLang     string `json:"target_lang"`
	TranslatedText string `json:"translated_text"`
	Provider       string `json:"provider"`
}

// Service manages message translation and caching.
type Service struct {
	mu       sync.RWMutex
	provider Provider
	dataDir  string
}

// NewService constructs a concrete translation Service.
func NewService(provider Provider, dataDir string) *Service {
	if provider == nil {
		provider = NewGoogleProvider("", nil)
	}
	return &Service{
		provider: provider,
		dataDir:  filepath.Clean(dataDir),
	}
}

// TranslateMessage fetches translation for message content, using cache when available.
func (s *Service) TranslateMessage(ctx context.Context, threadID, messageID, text, targetLang string) (*TranslationRecord, error) {
	threadID = strings.TrimSpace(threadID)
	messageID = strings.TrimSpace(messageID)
	text = strings.TrimSpace(text)
	targetLang = strings.TrimSpace(strings.ToLower(targetLang))

	if threadID == "" || messageID == "" {
		return nil, fmt.Errorf("%w: thread id and message id are required", domain.ErrValidation)
	}
	if text == "" {
		return nil, fmt.Errorf("%w: text to translate cannot be empty", domain.ErrValidation)
	}
	if targetLang == "" {
		targetLang = "en"
	}

	// 1. Check disk cache
	cachePath := s.getCachePath(threadID, messageID, targetLang)
	if rec, err := gitstore.ReadJSONFile[TranslationRecord](cachePath); err == nil && rec != nil {
		return rec, nil
	}

	// 2. Perform translation via provider
	translatedText, err := s.provider.Translate(ctx, text, targetLang)
	if err != nil {
		return nil, err
	}

	record := &TranslationRecord{
		MessageID:      messageID,
		TargetLang:     targetLang,
		TranslatedText: translatedText,
		Provider:       s.provider.Name(),
	}

	// 3. Save to disk cache asynchronously or synchronously
	s.mu.Lock()
	_ = gitstore.WriteJSONFile(cachePath, record)
	s.mu.Unlock()

	return record, nil
}

func (s *Service) getCachePath(threadID, messageID, targetLang string) string {
	fileName := fmt.Sprintf("%s_%s.json", messageID, targetLang)
	return filepath.Join(s.dataDir, "boards", "general", "threads", threadID, "translations", fileName)
}

func (s *Service) ClearCache(threadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	transDir := filepath.Join(s.dataDir, "boards", "general", "threads", threadID, "translations")
	return os.RemoveAll(transDir)
}
