package matchbuddy

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/inox/inox/backend/internal/ws"
)

type MatchStatus string

const (
	StatusLobby     MatchStatus = "lobby"
	StatusPlaying   MatchStatus = "playing"
	StatusPaused    MatchStatus = "paused"
	StatusCompleted MatchStatus = "completed"
)

// FeedCardKind denotes the classification of a feed broadcast.
type FeedCardKind string

const (
	FeedKindSwing       FeedCardKind = "swing"
	FeedKindMilestone   FeedCardKind = "milestone"
	FeedKindOverSummary FeedCardKind = "over_summary"
	FeedKindWhatIf      FeedCardKind = "whatif"
	FeedKindAnswer      FeedCardKind = "answer"
	FeedKindCatchUp     FeedCardKind = "catchup"
	FeedKindBot         FeedCardKind = "bot"
	FeedKindTeamChange  FeedCardKind = "team_change"
)

// FeedCard represents a structured companion update rendered in chat or match tab.
type FeedCard struct {
	ID        string            `json:"id"`
	Kind      FeedCardKind      `json:"kind"`
	Headline  string            `json:"headline"`
	BodyText  string            `json:"body_text"`
	Over      int               `json:"over,omitempty"`
	Ball      int               `json:"ball,omitempty"`
	Timestamp int64             `json:"timestamp"`
	Deltas    map[string]float64 `json:"deltas,omitempty"`
}

// MatchStatePayload contains the complete room snapshot sent on join or state transitions.
type MatchStatePayload struct {
	Status          MatchStatus          `json:"status"`
	SpeedMultiplier float64              `json:"speed_multiplier"`
	CurrentBallIdx  int                  `json:"current_ball_idx"`
	TotalBalls      int                  `json:"total_balls"`
	Runs            int                  `json:"runs"`
	Wickets         int                  `json:"wickets"`
	Overs           int                  `json:"overs"`
	BallsInOver     int                  `json:"balls_in_over"`
	Striker         *Player              `json:"striker,omitempty"`
	NonStriker      *Player              `json:"non_striker,omitempty"`
	Bowler          *Player              `json:"bowler,omitempty"`
	RecentBalls     []BallEvent          `json:"recent_balls"`
	TeamsLocked     bool                 `json:"teams_locked"`
	Leaderboard     []LeaderboardEntry   `json:"leaderboard"`
	Teams           map[string]FantasyTeam `json:"teams"`
	HasBot          bool                 `json:"has_bot"`
	Squad           []Player             `json:"squad"`
	Presets         []PresetDefinition   `json:"presets"`
}

// MatchBallPayload carries delivery updates to every room member.
type MatchBallPayload struct {
	Ball         BallEvent          `json:"ball"`
	PlayerDeltas map[string]float64 `json:"player_deltas"`
	MemberDeltas map[string]float64 `json:"member_deltas"`
	NewTotals    map[string]float64 `json:"new_totals"`
	RankChanges  []RankChange       `json:"rank_changes"`
	ScoreStr     string             `json:"score_str"`
	OversStr     string             `json:"overs_str"`
}

// WhatIfProjection contains deterministic preview data.
type WhatIfProjection struct {
	Scenario          string       `json:"scenario"`
	StrikerName       string       `json:"striker_name"`
	BowlerName        string       `json:"bowler_name"`
	MemberProjections []RankChange `json:"member_projections"`
	Explanation       string       `json:"explanation"`
}

// RoomEngine manages live match simulation for one watch party room.
type RoomEngine struct {
	mu          sync.RWMutex
	roomID      string
	status      MatchStatus
	speed       float64
	ballIndex   int
	teamsLocked bool

	teams       map[string]FantasyTeam
	teamChanges []TeamChange
	totals      map[string]float64
	playerStats map[string]PlayerRunningStats
	recentBalls []BallEvent
	lastLeader  string

	hasBot      bool
	commentator Commentator
	broadcast   func(*ws.Event)

	ctx    context.Context
	cancel context.CancelFunc
}

// NewRoomEngine constructs an engine for a room.
func NewRoomEngine(
	roomID string,
	commentator Commentator,
	broadcast func(*ws.Event),
) *RoomEngine {
	ctx, cancel := context.WithCancel(context.Background())
	e := &RoomEngine{
		roomID:      roomID,
		status:      StatusLobby,
		speed:       1.0,
		ballIndex:   0,
		teamsLocked: false,
		teams:       make(map[string]FantasyTeam),
		teamChanges: make([]TeamChange, 0),
		totals:      make(map[string]float64),
		playerStats: make(map[string]PlayerRunningStats),
		recentBalls: make([]BallEvent, 0, 6),
		hasBot:      false,
		commentator: commentator,
		broadcast:   broadcast,
		ctx:         ctx,
		cancel:      cancel,
	}

	return e
}

// Stop cleanly terminates any background ticker.
func (e *RoomEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.status = StatusPaused
	if e.cancel != nil {
		e.cancel()
	}
}

// Start begins live simulation, locking all teams.
func (e *RoomEngine) Start() error {
	e.mu.Lock()
	if e.status == StatusPlaying {
		e.mu.Unlock()
		return nil
	}
	if e.ballIndex >= len(SimulatedFixture) {
		e.mu.Unlock()
		return fmt.Errorf("match already completed; reset to start over")
	}

	// Lock all teams
	e.teamsLocked = true
	now := time.Now()
	for uid, t := range e.teams {
		if t.LockedAt == nil {
			t.LockedAt = &now
			e.teams[uid] = t
		}
	}
	e.status = StatusPlaying

	e.cancel()
	e.ctx, e.cancel = context.WithCancel(context.Background())
	go e.runLoop(e.ctx)

	state := e.buildSnapshotLocked()
	e.mu.Unlock()

	e.broadcastState(state)
	e.broadcastFeed(FeedCard{
		ID:        fmt.Sprintf("feed-start-%d", time.Now().UnixMilli()),
		Kind:      FeedKindMilestone,
		Headline:  "🏏 Match Started! Teams Locked!",
		BodyText:  "Mumbai Blasters innings underway. Watch each ball impact your fantasy points in real-time!",
		Timestamp: time.Now().UnixMilli(),
	})
	return nil
}

// Pause stops the simulation loop.
func (e *RoomEngine) Pause() {
	e.mu.Lock()
	if e.status != StatusPlaying {
		e.mu.Unlock()
		return
	}
	e.status = StatusPaused
	e.cancel()
	state := e.buildSnapshotLocked()
	e.mu.Unlock()

	e.broadcastState(state)
}

// Resume restarts simulation if paused.
func (e *RoomEngine) Resume() {
	e.mu.Lock()
	if e.status != StatusPaused {
		e.mu.Unlock()
		return
	}
	e.status = StatusPlaying
	e.cancel()
	e.ctx, e.cancel = context.WithCancel(context.Background())
	go e.runLoop(e.ctx)
	state := e.buildSnapshotLocked()
	e.mu.Unlock()

	e.broadcastState(state)
}

// SetSpeed changes playback rate (1x, 2x, 5x).
func (e *RoomEngine) SetSpeed(multiplier float64) {
	e.mu.Lock()
	if multiplier < 0.5 {
		multiplier = 1.0
	} else if multiplier > 10.0 {
		multiplier = 5.0
	}
	e.speed = multiplier
	state := e.buildSnapshotLocked()
	wasPlaying := e.status == StatusPlaying
	if wasPlaying {
		e.cancel()
		e.ctx, e.cancel = context.WithCancel(context.Background())
		go e.runLoop(e.ctx)
	}
	e.mu.Unlock()

	e.broadcastState(state)
}

// SkipToNextOver fast-forwards through all remaining balls in current over.
func (e *RoomEngine) SkipToNextOver() {
	e.mu.Lock()
	if e.ballIndex >= len(SimulatedFixture) {
		e.mu.Unlock()
		return
	}

	curOver := SimulatedFixture[e.ballIndex].Over
	var ballsToApply []BallEvent
	for e.ballIndex < len(SimulatedFixture) && SimulatedFixture[e.ballIndex].Over == curOver {
		ballsToApply = append(ballsToApply, SimulatedFixture[e.ballIndex])
		e.ballIndex++
	}

	// Apply all skipped balls
	for _, b := range ballsToApply {
		e.applyBallLocked(b)
	}

	if e.ballIndex >= len(SimulatedFixture) {
		e.status = StatusCompleted
		e.cancel()
	}

	state := e.buildSnapshotLocked()
	e.mu.Unlock()

	e.broadcastState(state)
}

// Reset clears simulation progress back to pre-match lobby.
func (e *RoomEngine) Reset() {
	e.mu.Lock()
	e.cancel()
	e.status = StatusLobby
	e.ballIndex = 0
	e.teamsLocked = false
	e.totals = make(map[string]float64)
	e.playerStats = make(map[string]PlayerRunningStats)
	e.recentBalls = make([]BallEvent, 0, 6)
	e.lastLeader = ""

	for uid, t := range e.teams {
		t.LockedAt = nil
		e.teams[uid] = t
		e.totals[uid] = 0
	}

	state := e.buildSnapshotLocked()
	e.mu.Unlock()

	e.broadcastState(state)
	e.broadcastFeed(FeedCard{
		ID:        fmt.Sprintf("feed-reset-%d", time.Now().UnixMilli()),
		Kind:      FeedKindMilestone,
		Headline:  "🔄 Match Reset to Lobby",
		BodyText:  "Lineups unlocked. Adjust your 11, captain, and vice-captain before restarting!",
		Timestamp: time.Now().UnixMilli(),
	})
}

// ToggleSolo adds or removes the virtual friend "Rahul (bot)".
func (e *RoomEngine) ToggleSolo() {
	e.mu.Lock()
	e.hasBot = !e.hasBot

	if e.hasBot {
		preset := Presets[1] // Bowler Heavy preset
		e.teams["user_bot_rahul"] = FantasyTeam{
			UserID:        "user_bot_rahul",
			Username:      "Rahul (bot)",
			PlayerIDs:     preset.PlayerIDs,
			CaptainID:     preset.CaptainID,
			ViceCaptainID: preset.ViceCaptainID,
			IsBot:         true,
			PresetID:      preset.ID,
		}
		e.totals["user_bot_rahul"] = 0
		e.recordTeamChangeLocked("user_bot_rahul", "Rahul (bot)", "team_created", "", preset.Name)
	} else {
		delete(e.teams, "user_bot_rahul")
		delete(e.totals, "user_bot_rahul")
	}

	state := e.buildSnapshotLocked()
	e.mu.Unlock()

	e.broadcastState(state)
	e.broadcastTeamChanges()
}

// SetTeam saves a user's fantasy selection.
func (e *RoomEngine) SetTeam(
	userID, username string,
	playerIDs []string,
	captainID, viceCaptainID string,
	presetID string,
) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.teamsLocked {
		return fmt.Errorf("teams are locked; match in progress")
	}

	if presetID != "" {
		preset := PresetByID(presetID)
		if preset == nil {
			return fmt.Errorf("unknown preset: %s", presetID)
		}
		playerIDs = preset.PlayerIDs
		captainID = preset.CaptainID
		viceCaptainID = preset.ViceCaptainID
	}

	team := FantasyTeam{
		UserID:        userID,
		Username:      username,
		PlayerIDs:     playerIDs,
		CaptainID:     captainID,
		ViceCaptainID: viceCaptainID,
		PresetID:      presetID,
	}

	if err := ValidateTeam(team, SquadByID); err != nil {
		return err
	}

	existing, hadExisting := e.teams[userID]
	e.teams[userID] = team
	if _, ok := e.totals[userID]; !ok {
		e.totals[userID] = 0
	}

	// Audit trail
	if !hadExisting {
		e.recordTeamChangeLocked(userID, username, "team_created", "", "Custom 11")
	} else {
		if existing.CaptainID != captainID {
			e.recordTeamChangeLocked(userID, username, "captain_change", existing.CaptainID, captainID)
		}
		if existing.ViceCaptainID != viceCaptainID {
			e.recordTeamChangeLocked(userID, username, "vc_change", existing.ViceCaptainID, viceCaptainID)
		}
		if existing.PresetID != presetID && presetID != "" {
			e.recordTeamChangeLocked(userID, username, "player_swap", existing.PresetID, presetID)
		}
	}

	state := e.buildSnapshotLocked()
	go func() {
		e.broadcastState(state)
		e.broadcastTeamChanges()
	}()

	return nil
}

// Snapshot returns the current room state.
func (e *RoomEngine) Snapshot() MatchStatePayload {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.buildSnapshotLocked()
}

// TeamChangesSnapshot returns the audit history.
func (e *RoomEngine) TeamChangesSnapshot() []TeamChange {
	e.mu.RLock()
	defer e.mu.RUnlock()
	copied := make([]TeamChange, len(e.teamChanges))
	copy(copied, e.teamChanges)
	return copied
}

func (e *RoomEngine) recordTeamChangeLocked(userID, username, kind, from, to string) {
	e.teamChanges = append(e.teamChanges, TeamChange{
		UserID:   userID,
		Username: username,
		Kind:     kind,
		From:     from,
		To:       to,
		At:       time.Now(),
	})
}

func (e *RoomEngine) runLoop(ctx context.Context) {
	for {
		e.mu.RLock()
		spd := e.speed
		idx := e.ballIndex
		total := len(SimulatedFixture)
		e.mu.RUnlock()

		if idx >= total {
			e.mu.Lock()
			e.status = StatusCompleted
			state := e.buildSnapshotLocked()
			e.mu.Unlock()
			e.broadcastState(state)
			return
		}

		intervalMs := int(5000.0 / spd)
		if intervalMs < 200 {
			intervalMs = 200
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(intervalMs) * time.Millisecond):
			e.tickOneBall()
		}
	}
}

func (e *RoomEngine) tickOneBall() {
	e.mu.Lock()
	if e.status != StatusPlaying || e.ballIndex >= len(SimulatedFixture) {
		e.mu.Unlock()
		return
	}

	ball := SimulatedFixture[e.ballIndex]
	e.ballIndex++
	completed := e.ballIndex >= len(SimulatedFixture)
	if completed {
		e.status = StatusCompleted
	}

	memberDeltas, newTotals, playerDeltas, rankChanges := e.applyBallLocked(ball)
	state := e.buildSnapshotLocked()
	e.mu.Unlock()

	// Broadcast ball update
	scoreStr := fmt.Sprintf("%d/%d", state.Runs, state.Wickets)
	oversStr := fmt.Sprintf("%d.%d", state.Overs, state.BallsInOver)

	ballEvt := &ws.Event{
		Type:      ws.EventType("MATCH_BALL"),
		RoomID:    e.roomID,
		Timestamp: time.Now().UnixMilli(),
	}
	ballPayload := MatchBallPayload{
		Ball:         ball,
		PlayerDeltas: playerDeltas,
		MemberDeltas: memberDeltas,
		NewTotals:    newTotals,
		RankChanges:  rankChanges,
		ScoreStr:     scoreStr,
		OversStr:     oversStr,
	}
	ballEvt.Payload = mustMarshal(ballPayload)
	e.broadcast(ballEvt)

	// Check feed card rules (anti-spam filter)
	go e.evaluateFeedCards(ball, memberDeltas, rankChanges, state)
}

func (e *RoomEngine) applyBallLocked(ball BallEvent) (
	memberDeltas map[string]float64,
	newTotals map[string]float64,
	playerDeltas map[string]float64,
	rankChanges []RankChange,
) {
	teamList := make([]FantasyTeam, 0, len(e.teams))
	for _, t := range e.teams {
		teamList = append(teamList, t)
	}

	memberDeltas, newTotals, playerDeltas, e.playerStats, rankChanges = ApplyBall(
		ball, e.playerStats, teamList, e.totals,
	)
	e.totals = newTotals

	// Append to recent balls (keep last 6)
	e.recentBalls = append(e.recentBalls, ball)
	if len(e.recentBalls) > 6 {
		e.recentBalls = e.recentBalls[len(e.recentBalls)-6:]
	}

	return memberDeltas, newTotals, playerDeltas, rankChanges
}

// Anti-spam rules: Emit feed card on notable moments only.
func (e *RoomEngine) evaluateFeedCards(
	ball BallEvent,
	memberDeltas map[string]float64,
	rankChanges []RankChange,
	state MatchStatePayload,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	var maxDelta float64
	var maxGainerID string
	for uid, d := range memberDeltas {
		if d > maxDelta {
			maxDelta = d
			maxGainerID = uid
		}
	}

	maxGainerName := maxGainerID
	for _, t := range state.Teams {
		if t.UserID == maxGainerID {
			maxGainerName = t.Username
		}
	}

	strikerName := ball.StrikerID
	if p, ok := SquadByID[ball.StrikerID]; ok {
		strikerName = p.Name
	}
	bowlerName := ball.BowlerID
	if p, ok := SquadByID[ball.BowlerID]; ok {
		bowlerName = p.Name
	}

	// Rule 1: Leader Change
	var leaderChanged bool
	newLeader := ""
	if len(rankChanges) > 0 {
		newLeader = rankChanges[0].UserID
		if e.lastLeader != "" && newLeader != e.lastLeader {
			leaderChanged = true
		}
		e.lastLeader = newLeader
	}

	if leaderChanged {
		leaderName := rankChanges[0].Username
		gap := 0.0
		if len(rankChanges) > 1 {
			gap = rankChanges[0].NewTotal - rankChanges[1].NewTotal
		}
		headline := fmt.Sprintf("🔥 Lead Change! %s takes Rank #1!", leaderName)
		body := fmt.Sprintf("%s surges into 1st place with %.1f pts (+%.1f lead).",
			leaderName, rankChanges[0].NewTotal, gap)

		e.broadcastFeed(FeedCard{
			ID:        fmt.Sprintf("feed-lead-%d-%d", ball.Over, ball.Ball),
			Kind:      FeedKindSwing,
			Headline:  headline,
			BodyText:  body,
			Over:      ball.Over,
			Ball:      ball.Ball,
			Timestamp: time.Now().UnixMilli(),
			Deltas:    memberDeltas,
		})
	}

	// Rule 2: Wicket
	if ball.Wicket != nil {
		outName := ball.Wicket.PlayerOut
		if p, ok := SquadByID[ball.Wicket.PlayerOut]; ok {
			outName = p.Name
		}
		headline := fmt.Sprintf("💥 WICKET! %s dismissed!", outName)
		body, _ := e.commentator.GenerateSwingCommentary(ctx, outName, bowlerName, 0, false, true, maxGainerName, maxDelta)

		e.broadcastFeed(FeedCard{
			ID:        fmt.Sprintf("feed-wicket-%d-%d", ball.Over, ball.Ball),
			Kind:      FeedKindSwing,
			Headline:  headline,
			BodyText:  body,
			Over:      ball.Over,
			Ball:      ball.Ball,
			Timestamp: time.Now().UnixMilli(),
			Deltas:    memberDeltas,
		})

		// Bot reaction on wicket if applicable
		if state.HasBot && maxGainerID != "user_bot_rahul" && rand.Float64() < 0.6 {
			botLines := []string{
				"Bhai, tera captain gaya 😅",
				"Ye wicket se mera game ban gaya! 🚀",
				"Arre re, catch of the tournament! 🏏",
			}
			e.broadcastFeed(FeedCard{
				ID:        fmt.Sprintf("feed-bot-%d", time.Now().UnixMilli()),
				Kind:      FeedKindBot,
				Headline:  "Rahul (bot)",
				BodyText:  botLines[rand.Intn(len(botLines))],
				Timestamp: time.Now().UnixMilli(),
			})
		}
		return
	}

	// Rule 3: Big Boundary / Six / Delta >= 10
	if ball.IsSix || maxDelta >= 10.0 {
		headline := fmt.Sprintf("🚀 Maximum! %s hits a SIX!", strikerName)
		body, _ := e.commentator.GenerateSwingCommentary(ctx, strikerName, bowlerName, ball.RunsOffBat, false, false, maxGainerName, maxDelta)

		e.broadcastFeed(FeedCard{
			ID:        fmt.Sprintf("feed-six-%d-%d", ball.Over, ball.Ball),
			Kind:      FeedKindSwing,
			Headline:  headline,
			BodyText:  body,
			Over:      ball.Over,
			Ball:      ball.Ball,
			Timestamp: time.Now().UnixMilli(),
			Deltas:    memberDeltas,
		})
		return
	}

	// Rule 4: End of Over Summary
	if ball.Ball == 6 && ball.ExtraType != "wide" && ball.ExtraType != "no_ball" {
		leaderName := ""
		gap := 0.0
		if len(state.Leaderboard) > 0 {
			leaderName = state.Leaderboard[0].Username
			if len(state.Leaderboard) > 1 {
				gap = state.Leaderboard[0].Points - state.Leaderboard[1].Points
			}
		}
		scoreStr := fmt.Sprintf("%d", state.Runs)
		summary, _ := e.commentator.GenerateOverSummary(ctx, ball.Over, scoreStr, state.Wickets, leaderName, gap)

		e.broadcastFeed(FeedCard{
			ID:        fmt.Sprintf("feed-over-%d", ball.Over),
			Kind:      FeedKindOverSummary,
			Headline:  fmt.Sprintf("📊 End of Over %d Summary", ball.Over),
			BodyText:  summary,
			Over:      ball.Over,
			Ball:      ball.Ball,
			Timestamp: time.Now().UnixMilli(),
		})
	}
}

// WhatIf performs a deterministic projection of the next delivery.
func (e *RoomEngine) WhatIf(scenario string) (*WhatIfProjection, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.ballIndex >= len(SimulatedFixture) {
		return nil, fmt.Errorf("match already completed")
	}

	nextBall := SimulatedFixture[e.ballIndex]
	strikerID := nextBall.StrikerID
	bowlerID := nextBall.BowlerID

	strikerName := strikerID
	if p, ok := SquadByID[strikerID]; ok {
		strikerName = p.Name
	}
	bowlerName := bowlerID
	if p, ok := SquadByID[bowlerID]; ok {
		bowlerName = p.Name
	}

	// Build synthetic ball
	synthBall := BallEvent{
		Over:       nextBall.Over,
		Ball:       nextBall.Ball,
		StrikerID:  strikerID,
		BowlerID:   bowlerID,
		ExtraType:  "none",
	}

	switch scenario {
	case "six":
		synthBall.RunsOffBat = 6
		synthBall.IsSix = true
	case "four":
		synthBall.RunsOffBat = 4
		synthBall.IsBoundary4 = true
	case "wicket":
		synthBall.RunsOffBat = 0
		synthBall.Wicket = &WicketInfo{
			Kind:      "bowled",
			PlayerOut: strikerID,
			BowlerID:  bowlerID,
		}
	case "dot":
		synthBall.RunsOffBat = 0
	default:
		return nil, fmt.Errorf("unsupported scenario: %s", scenario)
	}

	teamList := make([]FantasyTeam, 0, len(e.teams))
	for _, t := range e.teams {
		teamList = append(teamList, t)
	}

	_, _, _, _, rankChanges := ApplyBall(synthBall, e.playerStats, teamList, e.totals)

	var maxDelta float64
	var maxGainer string
	for _, rc := range rankChanges {
		if rc.PointsDelta > maxDelta {
			maxDelta = rc.PointsDelta
			maxGainer = rc.Username
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	explanation, _ := e.commentator.ExplainWhatIf(ctx, scenario, strikerName, bowlerName, maxGainer, maxDelta)

	proj := &WhatIfProjection{
		Scenario:          scenario,
		StrikerName:       strikerName,
		BowlerName:        bowlerName,
		MemberProjections: rankChanges,
		Explanation:       explanation,
	}

	return proj, nil
}

// Ask queries the commentator with match state context.
func (e *RoomEngine) Ask(question string) (string, error) {
	e.mu.RLock()
	state := e.buildSnapshotLocked()
	e.mu.RUnlock()

	summary := MatchStateSummary{
		ScoreStr: fmt.Sprintf("%d/%d", state.Runs, state.Wickets),
		OversStr: fmt.Sprintf("%d.%d", state.Overs, state.BallsInOver),
		CurrentBatters: []string{
			nameOrEmpty(state.Striker),
			nameOrEmpty(state.NonStriker),
		},
		CurrentBowler: nameOrEmpty(state.Bowler),
		Leaderboard:   state.Leaderboard,
	}
	for _, b := range state.RecentBalls {
		summary.RecentEvents = append(summary.RecentEvents, b.CommentaryText)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	return e.commentator.AnswerQuestion(ctx, question, summary)
}

// CatchUp produces a brief catch-up recap.
func (e *RoomEngine) CatchUp() (string, error) {
	e.mu.RLock()
	state := e.buildSnapshotLocked()
	e.mu.RUnlock()

	summary := MatchStateSummary{
		ScoreStr: fmt.Sprintf("%d/%d", state.Runs, state.Wickets),
		OversStr: fmt.Sprintf("%d.%d", state.Overs, state.BallsInOver),
		CurrentBatters: []string{
			nameOrEmpty(state.Striker),
			nameOrEmpty(state.NonStriker),
		},
		CurrentBowler: nameOrEmpty(state.Bowler),
		Leaderboard:   state.Leaderboard,
	}
	for _, b := range state.RecentBalls {
		summary.RecentEvents = append(summary.RecentEvents, b.CommentaryText)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	return e.commentator.GenerateCatchUp(ctx, summary)
}

func nameOrEmpty(p *Player) string {
	if p == nil {
		return ""
	}
	return p.Name
}

func (e *RoomEngine) buildSnapshotLocked() MatchStatePayload {
	var totalRuns, totalWickets, overs, ballsInOver int
	var striker, nonStriker, bowler *Player

	if e.ballIndex > 0 && e.ballIndex <= len(SimulatedFixture) {
		lastBall := SimulatedFixture[e.ballIndex-1]
		overs = lastBall.Over - 1
		ballsInOver = lastBall.Ball
		if ballsInOver == 6 {
			overs++
			ballsInOver = 0
		}
	}

	// Calculate score from fixture events up to current ballIndex
	for i := 0; i < e.ballIndex; i++ {
		b := SimulatedFixture[i]
		totalRuns += b.RunsOffBat + b.Extras
		if b.Wicket != nil {
			totalWickets++
		}
	}

	// Active batters and bowler
	if e.ballIndex < len(SimulatedFixture) {
		next := SimulatedFixture[e.ballIndex]
		if p, ok := SquadByID[next.StrikerID]; ok {
			striker = &p
		}
		if p, ok := SquadByID[next.NonStrikerID]; ok {
			nonStriker = &p
		}
		if p, ok := SquadByID[next.BowlerID]; ok {
			bowler = &p
		}
	} else if len(SimulatedFixture) > 0 {
		last := SimulatedFixture[len(SimulatedFixture)-1]
		if p, ok := SquadByID[last.StrikerID]; ok {
			striker = &p
		}
		if p, ok := SquadByID[last.NonStrikerID]; ok {
			nonStriker = &p
		}
		if p, ok := SquadByID[last.BowlerID]; ok {
			bowler = &p
		}
	}

	// Leaderboard calculation
	type userRank struct {
		team  FantasyTeam
		score float64
	}
	list := make([]userRank, 0, len(e.teams))
	for uid, t := range e.teams {
		list = append(list, userRank{
			team:  t,
			score: e.totals[uid],
		})
	}
	// Sort descending by points
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].score > list[i].score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	leaderboard := make([]LeaderboardEntry, len(list))
	for i, item := range list {
		cName := item.team.CaptainID
		if p, ok := SquadByID[item.team.CaptainID]; ok {
			cName = p.Name
		}
		vcName := item.team.ViceCaptainID
		if p, ok := SquadByID[item.team.ViceCaptainID]; ok {
			vcName = p.Name
		}
		leaderboard[i] = LeaderboardEntry{
			UserID:      item.team.UserID,
			Username:    item.team.Username,
			Points:      item.score,
			Rank:        i + 1,
			CaptainName: cName,
			VCName:      vcName,
			IsBot:       item.team.IsBot,
		}
	}

	// Copy teams
	copiedTeams := make(map[string]FantasyTeam, len(e.teams))
	for k, v := range e.teams {
		copiedTeams[k] = v
	}

	copiedRecent := make([]BallEvent, len(e.recentBalls))
	copy(copiedRecent, e.recentBalls)

	return MatchStatePayload{
		Status:          e.status,
		SpeedMultiplier: e.speed,
		CurrentBallIdx:  e.ballIndex,
		TotalBalls:      len(SimulatedFixture),
		Runs:            totalRuns,
		Wickets:         totalWickets,
		Overs:           overs,
		BallsInOver:     ballsInOver,
		Striker:         striker,
		NonStriker:      nonStriker,
		Bowler:          bowler,
		RecentBalls:     copiedRecent,
		TeamsLocked:     e.teamsLocked,
		Leaderboard:     leaderboard,
		Teams:           copiedTeams,
		HasBot:          e.hasBot,
		Squad:           Squad,
		Presets:         Presets,
	}
}

func (e *RoomEngine) broadcastState(state MatchStatePayload) {
	evt := &ws.Event{
		Type:      ws.EventType("MATCH_STATE"),
		RoomID:    e.roomID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   mustMarshal(state),
	}
	e.broadcast(evt)
}

func (e *RoomEngine) broadcastFeed(card FeedCard) {
	evt := &ws.Event{
		Type:      ws.EventType("MATCH_FEED"),
		RoomID:    e.roomID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   mustMarshal(card),
	}
	e.broadcast(evt)
}

func (e *RoomEngine) broadcastTeamChanges() {
	e.mu.RLock()
	changes := make([]TeamChange, len(e.teamChanges))
	copy(changes, e.teamChanges)
	e.mu.RUnlock()

	evt := &ws.Event{
		Type:      ws.EventType("MATCH_TEAM_CHANGES"),
		RoomID:    e.roomID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   mustMarshal(changes),
	}
	e.broadcast(evt)
}
