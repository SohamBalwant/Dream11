package matchbuddy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/inox/inox/backend/internal/matchbuddy"
)

func TestTemplateCommentator(t *testing.T) {
	tc := matchbuddy.NewTemplateCommentator()
	ctx := context.Background()

	// Over summary
	summary, err := tc.GenerateOverSummary(ctx, 1, "12", 0, "Alice", 14.5)
	if err != nil || !strings.Contains(summary, "Alice") {
		t.Fatalf("expected summary to contain Alice, got %q (err: %v)", summary, err)
	}

	// Swing commentary - wicket
	swing, err := tc.GenerateSwingCommentary(ctx, "Rohit Sharma", "Jasprit Bumrah", 0, false, true, "Bob", 25.0)
	if err != nil || !strings.Contains(swing, "TIMBER") {
		t.Fatalf("expected timber commentary, got %q (err: %v)", swing, err)
	}

	// Answer question
	state := matchbuddy.MatchStateSummary{
		ScoreStr: "45/2",
		OversStr: "3.4",
		Leaderboard: []matchbuddy.LeaderboardEntry{
			{Username: "Alice", Points: 88, CaptainName: "Virat Kohli"},
		},
	}
	ans, err := tc.AnswerQuestion(ctx, "Who is winning the match?", state)
	if err != nil || !strings.Contains(ans, "Alice") {
		t.Fatalf("expected answer to mention leader Alice, got %q", ans)
	}

	// What if
	whatIf, err := tc.ExplainWhatIf(ctx, "six", "Virat Kohli", "Jasprit Bumrah", "Alice", 12.0)
	if err != nil || !strings.Contains(whatIf, "SIX") {
		t.Fatalf("expected whatif explanation, got %q", whatIf)
	}

	// Catch up
	catchUp, err := tc.GenerateCatchUp(ctx, state)
	if err != nil || !strings.Contains(catchUp, "45/2") {
		t.Fatalf("expected catchup summary, got %q", catchUp)
	}
}

func TestGeminiCommentator_FallbackWhenNoKey(t *testing.T) {
	gc := matchbuddy.NewGeminiCommentator("", "gemini-1.5-flash")
	ctx := context.Background()

	// Should seamlessly fall back to template commentator
	summary, err := gc.GenerateOverSummary(ctx, 2, "20", 1, "Alice", 6.0)
	if err != nil {
		t.Fatalf("expected fallback not to error, got %v", err)
	}
	if !strings.Contains(summary, "Alice") {
		t.Fatalf("expected fallback output, got %q", summary)
	}
}

func TestGeminiCommentator_MockServerSuccess(t *testing.T) {
	mockResponse := map[string]any{
		"candidates": []map[string]any{
			{
				"content": map[string]any{
					"parts": []map[string]any{
						{
							"text": "Huge moment! Mumbai Blasters are putting on a batting masterclass!",
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	gc := matchbuddy.NewGeminiCommentator("fake-key", "gemini-1.5-flash")
	gc.SetBaseURL(server.URL)

	ctx := context.Background()
	text, err := gc.GenerateOverSummary(ctx, 3, "35", 1, "Alice", 8.0)
	if err != nil {
		t.Fatalf("expected success from mock server, got %v", err)
	}
	if !strings.Contains(text, "batting masterclass") {
		t.Fatalf("expected mock text, got %q", text)
	}
}

func TestGeminiCommentator_MockServerFailureFallsBack(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "quota exceeded", http.StatusTooManyRequests)
	}))
	defer server.Close()

	gc := matchbuddy.NewGeminiCommentator("fake-key", "gemini-1.5-flash")
	gc.SetBaseURL(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 500/429 should fall back to template commentator
	text, err := gc.GenerateOverSummary(ctx, 4, "48", 2, "Charlie", 4.0)
	if err != nil {
		t.Fatalf("expected fallback on HTTP error, got %v", err)
	}
	if !strings.Contains(text, "Charlie") {
		t.Fatalf("expected fallback commentary, got %q", text)
	}
}
