package translation

import "context"

// Provider abstracts external translation service providers (Google, LibreTranslate, DeepL).
type Provider interface {
	Name() string
	Translate(ctx context.Context, text string, targetLang string) (string, error)
}
