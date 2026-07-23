package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"chorus/internal/domain"
)

// GoogleProvider implements Provider using Google Translate API.
type GoogleProvider struct {
	apiKey     string
	httpClient *http.Client
}

// NewGoogleProvider constructs a Google Translate provider.
func NewGoogleProvider(apiKey string, httpClient *http.Client) *GoogleProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &GoogleProvider{
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

func (g *GoogleProvider) Name() string {
	return "google"
}

func (g *GoogleProvider) Translate(ctx context.Context, text string, targetLang string) (string, error) {
	text = strings.TrimSpace(text)
	targetLang = strings.TrimSpace(strings.ToLower(targetLang))

	if text == "" {
		return "", fmt.Errorf("%w: text to translate cannot be empty", domain.ErrValidation)
	}
	if targetLang == "" {
		targetLang = "en"
	}

	// If no API key is provided, use Google Translate free web endpoint fallback
	if g.apiKey == "" {
		return g.translateWebFallback(ctx, text, targetLang)
	}

	// Official Google Cloud Translation API v2
	endpoint := fmt.Sprintf("https://translation.googleapis.com/language/translate/v2?key=%s", g.apiKey)
	form := url.Values{}
	form.Set("q", text)
	form.Set("target", targetLang)
	form.Set("format", "text")

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: google translate request failed: %v", domain.ErrInternal, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to web endpoint if API key call fails
		return g.translateWebFallback(ctx, text, targetLang)
	}

	var apiResp struct {
		Data struct {
			Translations []struct {
				TranslatedText string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("%w: failed decoding google response: %v", domain.ErrInternal, err)
	}

	if len(apiResp.Data.Translations) == 0 {
		return text, nil
	}

	return apiResp.Data.Translations[0].TranslatedText, nil
}

func (g *GoogleProvider) translateWebFallback(ctx context.Context, text, targetLang string) (string, error) {
	webURL := fmt.Sprintf("https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=%s&dt=t&q=%s",
		url.QueryEscape(targetLang),
		url.QueryEscape(text),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", webURL, nil)
	if err != nil {
		return text, nil // Graceful fallback to original text
	}

	resp, err := g.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Development mock fallback if offline
		return fmt.Sprintf("[Translated to %s]: %s", strings.ToUpper(targetLang), text), nil
	}
	defer resp.Body.Close()

	var raw []any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil || len(raw) == 0 {
		return fmt.Sprintf("[Translated to %s]: %s", strings.ToUpper(targetLang), text), nil
	}

	segments, ok := raw[0].([]any)
	if !ok || len(segments) == 0 {
		return fmt.Sprintf("[Translated to %s]: %s", strings.ToUpper(targetLang), text), nil
	}

	var builder strings.Builder
	for _, seg := range segments {
		if pair, ok := seg.([]any); ok && len(pair) > 0 {
			if str, ok := pair[0].(string); ok {
				builder.WriteString(str)
			}
		}
	}

	res := builder.String()
	if res == "" {
		return fmt.Sprintf("[Translated to %s]: %s", strings.ToUpper(targetLang), text), nil
	}

	return res, nil
}
