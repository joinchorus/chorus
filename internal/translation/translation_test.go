package translation_test

import (
	"context"
	"testing"

	"chorus/internal/translation"
)

func TestTranslationService_CacheAndProvider(t *testing.T) {
	tempDir := t.TempDir()
	provider := translation.NewGoogleProvider("", nil)
	svc := translation.NewService(provider, tempDir)

	ctx := context.Background()

	// 1. Initial Translation Call
	rec1, err := svc.TranslateMessage(ctx, "thd_123", "msg_456", "Hello world", "es")
	if err != nil {
		t.Fatalf("expected no error on translation, got %v", err)
	}
	if rec1.MessageID != "msg_456" {
		t.Errorf("expected message_id msg_456, got %s", rec1.MessageID)
	}
	if rec1.TargetLang != "es" {
		t.Errorf("expected target_lang es, got %s", rec1.TargetLang)
	}
	if rec1.TranslatedText == "" {
		t.Errorf("expected non-empty translated_text")
	}

	// 2. Second Call should hit Cache
	rec2, err := svc.TranslateMessage(ctx, "thd_123", "msg_456", "Hello world", "es")
	if err != nil {
		t.Fatalf("expected no error on cached translation, got %v", err)
	}
	if rec2.TranslatedText != rec1.TranslatedText {
		t.Errorf("expected cached translation text %q, got %q", rec1.TranslatedText, rec2.TranslatedText)
	}
}
