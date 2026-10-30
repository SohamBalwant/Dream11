package matchbuddy_test

import (
	"sync"
	"testing"
	"time"

	"github.com/inox/inox/backend/internal/matchbuddy"
	"github.com/inox/inox/backend/internal/ws"
)

func TestEngineLifecycleAndControls(t *testing.T) {
	var mu sync.Mutex
	events := make([]*ws.Event, 0)
	broadcast := func(e *ws.Event) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, e)
	}

	commentator := matchbuddy.NewTemplateCommentator()
	engine := matchbuddy.NewRoomEngine("test_room", commentator, broadcast)
	defer engine.Stop()

	// Initial State
	snap := engine.Snapshot()
	if snap.Status != matchbuddy.StatusLobby {
		t.Fatalf("expected initial status lobby, got %s", snap.Status)
	}
	if snap.TeamsLocked {
		t.Fatalf("expected teams to be unlocked initially")
	}

	// 1. Set Team with preset
	err := engine.SetTeam("user1", "FanOne", nil, "", "", "preset_batting")
	if err != nil {
		t.Fatalf("failed to set preset team: %v", err)
	}

	snap = engine.Snapshot()
	if len(snap.Teams) != 1 {
		t.Fatalf("expected 1 team in snapshot, got %d", len(snap.Teams))
	}
	if snap.Teams["user1"].CaptainID != "p_arjun_v" {
		t.Errorf("expected captain p_arjun_v, got %s", snap.Teams["user1"].CaptainID)
	}

	// 2. Toggle Solo mode ("Rahul (bot)")
	engine.ToggleSolo()
	snap = engine.Snapshot()
	if !snap.HasBot {
		t.Errorf("expected hasBot to be true")
	}
	if _, ok := snap.Teams["user_bot_rahul"]; !ok {
		t.Errorf("expected Rahul bot in teams")
	}

	// 3. Start Match (Teams lock)
	if err := engine.Start(); err != nil {
		t.Fatalf("failed to start match: %v", err)
	}
	snap = engine.Snapshot()
	if snap.Status != matchbuddy.StatusPlaying {
		t.Errorf("expected status playing, got %s", snap.Status)
	}
	if !snap.TeamsLocked {
		t.Errorf("expected teams to be locked after start")
	}

	// Verify team edits are rejected when locked
	err = engine.SetTeam("user1", "FanOne", nil, "", "", "preset_bowling")
	if err == nil {
		t.Errorf("expected error when setting team while locked, got nil")
	}

	// 4. Pause and Resume
	engine.Pause()
	snap = engine.Snapshot()
	if snap.Status != matchbuddy.StatusPaused {
		t.Errorf("expected status paused, got %s", snap.Status)
	}

	engine.Resume()
	snap = engine.Snapshot()
	if snap.Status != matchbuddy.StatusPlaying {
		t.Errorf("expected status playing, got %s", snap.Status)
	}

	// 5. Speed change
	engine.SetSpeed(5.0)
	snap = engine.Snapshot()
	if snap.SpeedMultiplier != 5.0 {
		t.Errorf("expected speed 5.0, got %v", snap.SpeedMultiplier)
	}

	// 6. Skip over
	engine.SkipToNextOver()
	snap = engine.Snapshot()
	if snap.Overs < 1 && snap.CurrentBallIdx < 6 {
		t.Errorf("expected over advancement after SkipToNextOver, got idx=%d", snap.CurrentBallIdx)
	}

	// 7. WhatIf projection (does not mutate engine state)
	idxBefore := snap.CurrentBallIdx
	proj, err := engine.WhatIf("six")
	if err != nil {
		t.Fatalf("WhatIf returned error: %v", err)
	}
	if proj.Scenario != "six" {
		t.Errorf("expected scenario six, got %s", proj.Scenario)
	}
	if len(proj.MemberProjections) == 0 {
		t.Errorf("expected projections for room members")
	}
	idxAfter := engine.Snapshot().CurrentBallIdx
	if idxBefore != idxAfter {
		t.Errorf("WhatIf mutated ball index! before=%d, after=%d", idxBefore, idxAfter)
	}

	// 8. Ask and CatchUp
	ans, err := engine.Ask("Who is winning?")
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if ans == "" {
		t.Errorf("expected non-empty answer from commentator")
	}

	catchUp, err := engine.CatchUp()
	if err != nil {
		t.Fatalf("CatchUp returned error: %v", err)
	}
	if catchUp == "" {
		t.Errorf("expected non-empty catch up message")
	}

	// 9. Reset
	engine.Reset()
	snap = engine.Snapshot()
	if snap.Status != matchbuddy.StatusLobby {
		t.Errorf("expected status lobby after reset, got %s", snap.Status)
	}
	if snap.TeamsLocked {
		t.Errorf("expected teams unlocked after reset")
	}
	if snap.CurrentBallIdx != 0 {
		t.Errorf("expected ball index 0 after reset, got %d", snap.CurrentBallIdx)
	}
}

func TestEngineConcurrency(t *testing.T) {
	broadcast := func(e *ws.Event) {}
	commentator := matchbuddy.NewTemplateCommentator()
	engine := matchbuddy.NewRoomEngine("test_race", commentator, broadcast)
	defer engine.Stop()

	_ = engine.SetTeam("u1", "U1", nil, "", "", "preset_batting")
	_ = engine.SetTeam("u2", "U2", nil, "", "", "preset_bowling")
	_ = engine.Start()
	engine.SetSpeed(10.0) // very fast ticking

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = engine.Snapshot()
				_, _ = engine.WhatIf("wicket")
				_, _ = engine.Ask("How's the game?")
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
}
