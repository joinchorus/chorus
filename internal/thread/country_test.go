package thread_test

import (
	"context"
	"testing"
	"time"

	"chorus/internal/conversationname"
	"chorus/internal/idgen"
	"chorus/internal/repository/memory"
	"chorus/internal/thread"
)

// The HTTP handler resolves the visitor's country (from CDN/GeoIP headers) and
// passes the resulting ISO code to the service. These tests pin that the
// service shows exactly that code, instead of re-resolving it as if it were an
// IP address (which fell back to "TR" for every visitor).
func newCountryTestService() *thread.Service {
	fixedTime := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	return thread.NewService(
		memory.NewThreadRepository(),
		idgen.NewRandomIDGenerator(),
		conversationname.NewDefaultGenerator([]string{"Ash", "River", "Echo"}),
		func() time.Time { return fixedTime },
	)
}

func TestCreateThread_UsesResolvedCountry(t *testing.T) {
	tests := []struct {
		name            string
		resolvedCountry string
		showCountry     bool
		want            *string
	}{
		{name: "germany stays germany", resolvedCountry: "DE", showCountry: true, want: strPtr("DE")},
		{name: "lowercase is normalized", resolvedCountry: " us ", showCountry: true, want: strPtr("US")},
		{name: "unknown country is shown as UN, not TR", resolvedCountry: "", showCountry: true, want: strPtr("UN")},
		{name: "opt-out hides country", resolvedCountry: "DE", showCountry: false, want: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newCountryTestService()
			th, err := svc.CreateThread(context.Background(), thread.CreateThreadInput{
				Title:       "Country flag check",
				Body:        "Body",
				ShowCountry: tc.showCountry,
			}, tc.resolvedCountry)
			if err != nil {
				t.Fatalf("CreateThread returned error: %v", err)
			}
			assertCountry(t, th.Country, tc.want)
		})
	}
}

func TestAddMessage_UsesResolvedCountry(t *testing.T) {
	svc := newCountryTestService()
	ctx := context.Background()

	th, err := svc.CreateThread(ctx, thread.CreateThreadInput{Title: "Reply flags", Body: "Body"}, "TR")
	if err != nil {
		t.Fatalf("CreateThread returned error: %v", err)
	}

	msg, err := svc.AddMessage(ctx, th.ID, thread.CreateMessageInput{
		Body:        "Hello from Berlin",
		ShowCountry: true,
	}, "DE")
	if err != nil {
		t.Fatalf("AddMessage returned error: %v", err)
	}
	assertCountry(t, msg.Country, strPtr("DE"))
}

func strPtr(s string) *string { return &s }

func assertCountry(t *testing.T, got, want *string) {
	t.Helper()
	switch {
	case want == nil && got != nil:
		t.Errorf("expected no country, got %q", *got)
	case want != nil && got == nil:
		t.Errorf("expected country %q, got nil", *want)
	case want != nil && *got != *want:
		t.Errorf("expected country %q, got %q", *want, *got)
	}
}
