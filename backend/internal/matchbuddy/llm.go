package matchbuddy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Commentator provides conversational match analysis, over summaries, and Q&A.
type Commentator interface {
	GenerateOverSummary(ctx context.Context, overNum int, scoreStr string, wickets int, leaderName string, gap float64) (string, error)
	GenerateSwingCommentary(ctx context.Context, batterName, bowlerName string, runs int, isBoundary, isWicket bool, gainerName string, pointsDelta float64) (string, error)
	AnswerQuestion(ctx context.Context, question string, state MatchStateSummary) (string, error)
	ExplainWhatIf(ctx context.Context, scenario, strikerName, bowlerName string, gainerName string, maxDelta float64) (string, error)
	GenerateCatchUp(ctx context.Context, state MatchStateSummary) (string, error)
}

// MatchStateSummary holds structured context provided to the commentator.
type MatchStateSummary struct {
	ScoreStr       string             `json:"score"`
	OversStr       string             `json:"overs"`
	CurrentBatters []string           `json:"current_batters"`
	CurrentBowler  string             `json:"current_bowler"`
	RecentEvents   []string           `json:"recent_events"`
	Leaderboard    []LeaderboardEntry `json:"leaderboard"`
	TopPerformer   string             `json:"top_performer"`
}

// LeaderboardEntry reflects a user's standing.
type LeaderboardEntry struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	Points      float64 `json:"points"`
	Rank        int     `json:"rank"`
	CaptainName string  `json:"captain_name"`
	VCName      string  `json:"vc_name"`
	IsBot       bool    `json:"is_bot,omitempty"`
}

// TemplateCommentator generates witty, authentic cricket commentary without network dependencies.
type TemplateCommentator struct{}

func NewTemplateCommentator() *TemplateCommentator {
	return &TemplateCommentator{}
}

func (t *TemplateCommentator) GenerateOverSummary(
	_ context.Context,
	overNum int,
	scoreStr string,
	wickets int,
	leaderName string,
	gap float64,
) (string, error) {
	if leaderName == "" {
		return fmt.Sprintf("End of Over %d: Mumbai Blasters are %s/%d. Game is wide open!", overNum, scoreStr, wickets), nil
	}
	if gap <= 5 {
		return fmt.Sprintf("End of Over %d (%s/%d). Nail-biter! %s clings to the top spot with a razor-thin %.1f pt lead!",
			overNum, scoreStr, wickets, leaderName, gap), nil
	}
	return fmt.Sprintf("End of Over %d (%s/%d). %s is bossing the fantasy leaderboard with a commanding %.1f pt advantage!",
		overNum, scoreStr, wickets, leaderName, gap), nil
}

func (t *TemplateCommentator) GenerateSwingCommentary(
	_ context.Context,
	batterName, bowlerName string,
	runs int,
	isBoundary, isWicket bool,
	gainerName string,
	pointsDelta float64,
) (string, error) {
	if isWicket {
		return fmt.Sprintf("TIMBER! %s dismisses %s! Massive fantasy swing: %s racks up +%.1f pts on this dismissal!",
			bowlerName, batterName, gainerName, pointsDelta), nil
	}
	if runs >= 6 {
		return fmt.Sprintf("BOOM! %s launches a towering maximum! %s leaps forward by +%.1f pts!",
			batterName, gainerName, pointsDelta), nil
	}
	if isBoundary {
		return fmt.Sprintf("Classy stroke! %s finds the fence. %s banks +%.1f pts!",
			batterName, gainerName, pointsDelta), nil
	}
	return fmt.Sprintf("%s rotates strike; %s picks up +%.1f pts.", batterName, gainerName, pointsDelta), nil
}

func (t *TemplateCommentator) AnswerQuestion(
	_ context.Context,
	question string,
	state MatchStateSummary,
) (string, error) {
	q := strings.ToLower(question)

	if strings.Contains(q, "who is winning") || strings.Contains(q, "leader") || strings.Contains(q, "first") {
		if len(state.Leaderboard) > 0 {
			top := state.Leaderboard[0]
			var runnerUp string
			if len(state.Leaderboard) > 1 {
				runnerUp = fmt.Sprintf(" with %s closely trailing on %.1f pts", state.Leaderboard[1].Username, state.Leaderboard[1].Points)
			}
			return fmt.Sprintf("%s leads the pack on %.1f pts%s. Their captain %s is delivering key returns!",
				top.Username, top.Points, runnerUp, top.CaptainName), nil
		}
		return "The fantasy contest is in the lobby stage. Lock your teams to see the leader!", nil
	}

	if strings.Contains(q, "score") || strings.Contains(q, "match") || strings.Contains(q, "how many") {
		return fmt.Sprintf("Mumbai Blasters are at %s after %s overs. Batters in the middle: %s; Bowler: %s.",
			state.ScoreStr, state.OversStr, strings.Join(state.CurrentBatters, " & "), state.CurrentBowler), nil
	}

	if strings.Contains(q, "best") || strings.Contains(q, "top") || strings.Contains(q, "player") {
		return fmt.Sprintf("Dev Pandya and Arjun Verma are generating the biggest fantasy swings right now with their boundary firepower."), nil
	}

	return fmt.Sprintf("Current state: %s after %s overs. %s holds rank #1 with %.1f pts. Keep watching the live deltas after each delivery!",
		state.ScoreStr, state.OversStr, state.Leaderboard[0].Username, state.Leaderboard[0].Points), nil
}

func (t *TemplateCommentator) ExplainWhatIf(
	_ context.Context,
	scenario, strikerName, bowlerName string,
	gainerName string,
	maxDelta float64,
) (string, error) {
	switch scenario {
	case "six":
		return fmt.Sprintf("If %s hits a SIX next ball (+12 pts raw), %s gains the most (+%.1f pts with multipliers) and tightens their leaderboard grip!",
			strikerName, gainerName, maxDelta), nil
	case "four":
		return fmt.Sprintf("If %s strikes a FOUR (+8 pts raw), %s collects +%.1f pts to push up the rankings.",
			strikerName, gainerName, maxDelta), nil
	case "wicket":
		return fmt.Sprintf("If %s takes a WICKET (+25 pts raw), bowler-heavy teams gain +%.1f pts, while %s owners take a major hit!",
			bowlerName, maxDelta, strikerName), nil
	case "dot":
		return fmt.Sprintf("A dot ball puts pressure on the batters. Dot streaks boost economy rates and favour bowling lineups."), nil
	default:
		return fmt.Sprintf("Projected scenario '%s' for %s bowling to %s delivers up to +%.1f pts for %s.",
			scenario, bowlerName, strikerName, maxDelta, gainerName), nil
	}
}

func (t *TemplateCommentator) GenerateCatchUp(
	_ context.Context,
	state MatchStateSummary,
) (string, error) {
	if len(state.Leaderboard) == 0 {
		return "Match Buddy is ready! Pre-match teams are drafting in the lobby.", nil
	}
	top := state.Leaderboard[0]
	return fmt.Sprintf("Catch-up: Score is %s in %s overs. %s leads the room on %.1f pts with Captain %s. Recent action: %s.",
		state.ScoreStr, state.OversStr, top.Username, top.Points, top.CaptainName,
		strings.Join(state.RecentEvents, "; ")), nil
}

// GeminiCommentator calls Google's Gemini REST API with strict timeouts and automatic template fallback.
type GeminiCommentator struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	fallback   Commentator
}

func NewGeminiCommentator(apiKey, model string) *GeminiCommentator {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &GeminiCommentator{
		apiKey:     apiKey,
		model:      model,
		baseURL:    "https://generativelanguage.googleapis.com",
		httpClient: &http.Client{Timeout: 6 * time.Second},
		fallback:   NewTemplateCommentator(),
	}
}

// SetBaseURL allows overriding the Gemini API host (for testing).
func (g *GeminiCommentator) SetBaseURL(url string) {
	g.baseURL = url
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (g *GeminiCommentator) callGemini(ctx context.Context, prompt string) (string, error) {
	if g.apiKey == "" {
		return "", fmt.Errorf("gemini api key not configured")
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", g.baseURL, g.model, g.apiKey)

	reqPayload := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: "You are 'Match Buddy', a witty and strategic fantasy cricket companion inside a live watch-party room. Keep responses short (under 2 sentences), punchy, and focused on fantasy point impacts and banter. Fictional match: Mumbai Blasters vs Delhi Titans.\n\n" + prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 150,
		},
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			lastErr = err
			continue
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := g.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("gemini api error (status %d): %s", resp.StatusCode, string(respBody))
			continue
		}
		if readErr != nil {
			lastErr = readErr
			continue
		}

		var parsed geminiResponse
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			lastErr = err
			continue
		}

		if len(parsed.Candidates) > 0 && len(parsed.Candidates[0].Content.Parts) > 0 {
			txt := strings.TrimSpace(parsed.Candidates[0].Content.Parts[0].Text)
			if txt != "" {
				return txt, nil
			}
		}

		lastErr = fmt.Errorf("empty gemini response candidates")
	}

	return "", lastErr
}

func (g *GeminiCommentator) GenerateOverSummary(
	ctx context.Context,
	overNum int,
	scoreStr string,
	wickets int,
	leaderName string,
	gap float64,
) (string, error) {
	prompt := fmt.Sprintf("Over %d completed. Score: %s/%d. Fantasy leader: %s with a %.1f point lead. Give a brief, exciting over summary emphasizing the fantasy swing.",
		overNum, scoreStr, wickets, leaderName, gap)
	text, err := g.callGemini(ctx, prompt)
	if err != nil {
		slog.Debug("Gemini over summary failed; falling back to template", "error", err)
		return g.fallback.GenerateOverSummary(ctx, overNum, scoreStr, wickets, leaderName, gap)
	}
	return text, nil
}

func (g *GeminiCommentator) GenerateSwingCommentary(
	ctx context.Context,
	batterName, bowlerName string,
	runs int,
	isBoundary, isWicket bool,
	gainerName string,
	pointsDelta float64,
) (string, error) {
	prompt := fmt.Sprintf("Ball action: %s bowling to %s. Outcome: %d runs (boundary: %t, wicket: %t). Biggest fantasy gainer: %s (+%.1f pts). Give a 1-2 sentence dramatic reaction celebrating this fantasy swing.",
		bowlerName, batterName, runs, isBoundary, isWicket, gainerName, pointsDelta)
	text, err := g.callGemini(ctx, prompt)
	if err != nil {
		slog.Debug("Gemini swing commentary failed; falling back to template", "error", err)
		return g.fallback.GenerateSwingCommentary(ctx, batterName, bowlerName, runs, isBoundary, isWicket, gainerName, pointsDelta)
	}
	return text, nil
}

func (g *GeminiCommentator) AnswerQuestion(
	ctx context.Context,
	question string,
	state MatchStateSummary,
) (string, error) {
	prompt := fmt.Sprintf("Fan question: \"%s\". Current match context: Score %s in %s overs; Batters: %s; Bowler: %s; Top leaderboard: %+v; Recent events: %s. Answer accurately in 1-2 sentences with fantasy insight and friendly banter.",
		question, state.ScoreStr, state.OversStr, strings.Join(state.CurrentBatters, ", "), state.CurrentBowler, state.Leaderboard, strings.Join(state.RecentEvents, "; "))
	text, err := g.callGemini(ctx, prompt)
	if err != nil {
		slog.Debug("Gemini AnswerQuestion failed; falling back to template", "error", err)
		return g.fallback.AnswerQuestion(ctx, question, state)
	}
	return text, nil
}

func (g *GeminiCommentator) ExplainWhatIf(
	ctx context.Context,
	scenario, strikerName, bowlerName string,
	gainerName string,
	maxDelta float64,
) (string, error) {
	prompt := fmt.Sprintf("Scenario: What if next ball is a '%s'? Striker: %s, Bowler: %s. Projected biggest winner: %s (+%.1f pts). Explain the strategic fantasy outcome in 1-2 clear, punchy sentences.",
		scenario, strikerName, bowlerName, gainerName, maxDelta)
	text, err := g.callGemini(ctx, prompt)
	if err != nil {
		slog.Debug("Gemini ExplainWhatIf failed; falling back to template", "error", err)
		return g.fallback.ExplainWhatIf(ctx, scenario, strikerName, bowlerName, gainerName, maxDelta)
	}
	return text, nil
}

func (g *GeminiCommentator) GenerateCatchUp(
	ctx context.Context,
	state MatchStateSummary,
) (string, error) {
	prompt := fmt.Sprintf("Generate a 2-sentence match and fantasy recap for a user who just joined the watch party. Score: %s in %s overs. Leaderboard: %+v. Recent events: %s.",
		state.ScoreStr, state.OversStr, state.Leaderboard, strings.Join(state.RecentEvents, "; "))
	text, err := g.callGemini(ctx, prompt)
	if err != nil {
		slog.Debug("Gemini GenerateCatchUp failed; falling back to template", "error", err)
		return g.fallback.GenerateCatchUp(ctx, state)
	}
	return text, nil
}
