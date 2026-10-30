package matchbuddy_test

import (
	"testing"

	"github.com/inox/inox/backend/internal/matchbuddy"
)

func TestValidateTeam(t *testing.T) {
	squadMap := matchbuddy.SquadByID

	t.Run("Valid preset team", func(t *testing.T) {
		preset := matchbuddy.Presets[0]
		team := matchbuddy.FantasyTeam{
			UserID:        "user1",
			Username:      "TestUser",
			PlayerIDs:     preset.PlayerIDs,
			CaptainID:     preset.CaptainID,
			ViceCaptainID: preset.ViceCaptainID,
		}
		if err := matchbuddy.ValidateTeam(team, squadMap); err != nil {
			t.Fatalf("expected valid team, got err: %v", err)
		}
	})

	t.Run("Invalid size (< 11)", func(t *testing.T) {
		team := matchbuddy.FantasyTeam{
			UserID:        "user1",
			PlayerIDs:     matchbuddy.Presets[0].PlayerIDs[:10],
			CaptainID:     matchbuddy.Presets[0].CaptainID,
			ViceCaptainID: matchbuddy.Presets[0].ViceCaptainID,
		}
		if err := matchbuddy.ValidateTeam(team, squadMap); err == nil {
			t.Fatalf("expected error for 10 players, got nil")
		}
	})

	t.Run("Duplicate players", func(t *testing.T) {
		pids := make([]string, 11)
		copy(pids, matchbuddy.Presets[0].PlayerIDs)
		pids[1] = pids[0] // duplicate
		team := matchbuddy.FantasyTeam{
			UserID:        "user1",
			PlayerIDs:     pids,
			CaptainID:     pids[0],
			ViceCaptainID: pids[2],
		}
		if err := matchbuddy.ValidateTeam(team, squadMap); err == nil {
			t.Fatalf("expected error for duplicate player, got nil")
		}
	})

	t.Run("More than 7 from one real team", func(t *testing.T) {
		// All 8 Mumbai players + 3 Delhi players = 11 players, 8 from MB
		mbPlayers := []string{
			"p_arjun_v", "p_rohan_s", "p_kabir_r", "p_dev_p",
			"p_tariq_m", "p_jas_b", "p_varun_c", "p_kunal_k", // 8 MB
			"p_vikram_k", "p_ishan_g", "p_aarav_p", // 3 DT
		}
		team := matchbuddy.FantasyTeam{
			UserID:        "user1",
			PlayerIDs:     mbPlayers,
			CaptainID:     mbPlayers[0],
			ViceCaptainID: mbPlayers[1],
		}
		if err := matchbuddy.ValidateTeam(team, squadMap); err == nil {
			t.Fatalf("expected error for 8 players from MB, got nil")
		}
	})

	t.Run("Captain and VC are the same player", func(t *testing.T) {
		preset := matchbuddy.Presets[0]
		team := matchbuddy.FantasyTeam{
			UserID:        "user1",
			PlayerIDs:     preset.PlayerIDs,
			CaptainID:     preset.PlayerIDs[0],
			ViceCaptainID: preset.PlayerIDs[0],
		}
		if err := matchbuddy.ValidateTeam(team, squadMap); err == nil {
			t.Fatalf("expected error when Captain == VC, got nil")
		}
	})

	t.Run("Captain not in 11", func(t *testing.T) {
		preset := matchbuddy.Presets[0]
		team := matchbuddy.FantasyTeam{
			UserID:        "user1",
			PlayerIDs:     preset.PlayerIDs,
			CaptainID:     "p_varun_c", // not in preset 0
			ViceCaptainID: preset.ViceCaptainID,
		}
		if err := matchbuddy.ValidateTeam(team, squadMap); err == nil {
			t.Fatalf("expected error when Captain not in 11, got nil")
		}
	})
}

func TestPointsForBall_Batting(t *testing.T) {
	stats := make(map[string]matchbuddy.PlayerRunningStats)

	// Ball 1: Single run
	b1 := matchbuddy.BallEvent{
		StrikerID:  "p_arjun_v",
		RunsOffBat: 1,
	}
	deltas, stats := matchbuddy.PointsForBall(b1, stats)
	if deltas["p_arjun_v"] != 1.0 {
		t.Errorf("expected 1.0 pt, got %v", deltas["p_arjun_v"])
	}
	if stats["p_arjun_v"].Runs != 1 || stats["p_arjun_v"].BallsFaced != 1 {
		t.Errorf("expected 1 run, 1 ball faced; got %+v", stats["p_arjun_v"])
	}

	// Ball 2: Four (4 runs + 4 bonus = 8 pts)
	b2 := matchbuddy.BallEvent{
		StrikerID:   "p_arjun_v",
		RunsOffBat:  4,
		IsBoundary4: true,
	}
	deltas, stats = matchbuddy.PointsForBall(b2, stats)
	if deltas["p_arjun_v"] != 8.0 {
		t.Errorf("expected 8.0 pts for four, got %v", deltas["p_arjun_v"])
	}
	if stats["p_arjun_v"].Runs != 5 || stats["p_arjun_v"].Fours != 1 {
		t.Errorf("expected 5 runs, 1 four; got %+v", stats["p_arjun_v"])
	}

	// Ball 3: Six (6 runs + 6 bonus = 12 pts)
	b3 := matchbuddy.BallEvent{
		StrikerID:  "p_arjun_v",
		RunsOffBat: 6,
		IsSix:      true,
	}
	deltas, stats = matchbuddy.PointsForBall(b3, stats)
	if deltas["p_arjun_v"] != 12.0 {
		t.Errorf("expected 12.0 pts for six, got %v", deltas["p_arjun_v"])
	}
	if stats["p_arjun_v"].Runs != 11 || stats["p_arjun_v"].Sixes != 1 {
		t.Errorf("expected 11 runs, 1 six; got %+v", stats["p_arjun_v"])
	}
}

func TestPointsForBall_Milestones(t *testing.T) {
	stats := map[string]matchbuddy.PlayerRunningStats{
		"p_dev_p": {Runs: 28, BallsFaced: 10},
	}

	// Batter at 28 runs hits a four -> moves to 32 runs.
	// 4 runs + 4 bonus (boundary) + 8 bonus (30 milestone) = 16 pts
	b := matchbuddy.BallEvent{
		StrikerID:   "p_dev_p",
		RunsOffBat:  4,
		IsBoundary4: true,
	}
	deltas, newStats := matchbuddy.PointsForBall(b, stats)
	expected := 4.0 + 4.0 + 8.0 // 16.0
	if deltas["p_dev_p"] != expected {
		t.Fatalf("expected %v pts for boundary + 30 milestone, got %v", expected, deltas["p_dev_p"])
	}
	if !newStats["p_dev_p"].Milestone30Awarded {
		t.Fatalf("expected Milestone30Awarded = true")
	}

	// Another boundary takes him to 36 runs -> no 30 milestone again!
	b2 := matchbuddy.BallEvent{
		StrikerID:   "p_dev_p",
		RunsOffBat:  4,
		IsBoundary4: true,
	}
	deltas2, newStats2 := matchbuddy.PointsForBall(b2, newStats)
	if deltas2["p_dev_p"] != 8.0 {
		t.Fatalf("expected 8.0 pts (no repeat 30 milestone), got %v", deltas2["p_dev_p"])
	}

	// Batter at 48 runs hits a six -> 54 runs (crosses 50)
	// 6 runs + 6 bonus + 8 bonus (50 milestone replaces 30, so net +8) = 20 pts
	newStats2["p_dev_p"] = matchbuddy.PlayerRunningStats{
		Runs:               48,
		BallsFaced:         15,
		Milestone30Awarded: true,
	}
	b3 := matchbuddy.BallEvent{
		StrikerID:  "p_dev_p",
		RunsOffBat: 6,
		IsSix:      true,
	}
	deltas3, newStats3 := matchbuddy.PointsForBall(b3, newStats2)
	expected50 := 6.0 + 6.0 + 8.0 // 20.0
	if deltas3["p_dev_p"] != expected50 {
		t.Fatalf("expected %v pts for 50 milestone upgrade, got %v", expected50, deltas3["p_dev_p"])
	}
	if !newStats3["p_dev_p"].Milestone50Awarded {
		t.Fatalf("expected Milestone50Awarded = true")
	}
}

func TestPointsForBall_DuckAndWicket(t *testing.T) {
	stats := make(map[string]matchbuddy.PlayerRunningStats)

	// Golden duck: Kabir Ray (0 runs) bowled by Samir Khan
	b := matchbuddy.BallEvent{
		StrikerID: "p_kabir_r",
		BowlerID:  "p_samir_k",
		Wicket: &matchbuddy.WicketInfo{
			Kind:      "bowled",
			PlayerOut: "p_kabir_r",
			BowlerID:  "p_samir_k",
		},
	}

	deltas, newStats := matchbuddy.PointsForBall(b, stats)

	// Kabir should get -2 pts (duck)
	if deltas["p_kabir_r"] != -2.0 {
		t.Errorf("expected -2.0 duck penalty for Kabir, got %v", deltas["p_kabir_r"])
	}
	if !newStats["p_kabir_r"].IsOut {
		t.Errorf("expected Kabir IsOut = true")
	}

	// Samir Khan gets +25 pts for wicket
	if deltas["p_samir_k"] != 25.0 {
		t.Errorf("expected +25.0 pts for bowler, got %v", deltas["p_samir_k"])
	}
	if newStats["p_samir_k"].Wickets != 1 {
		t.Errorf("expected Samir Wickets = 1")
	}
}

func TestPointsForBall_RunOutNoBowlerPoints(t *testing.T) {
	stats := map[string]matchbuddy.PlayerRunningStats{
		"p_arjun_v": {Runs: 15},
	}

	b := matchbuddy.BallEvent{
		StrikerID: "p_arjun_v",
		BowlerID:  "p_samir_k",
		Wicket: &matchbuddy.WicketInfo{
			Kind:      "run_out",
			PlayerOut: "p_arjun_v",
		},
	}

	deltas, _ := matchbuddy.PointsForBall(b, stats)
	if deltas["p_samir_k"] != 0 {
		t.Errorf("bowler should not receive wicket points for run out, got %v", deltas["p_samir_k"])
	}
	if deltas["p_arjun_v"] != 0 {
		t.Errorf("batter with 15 runs should not get duck penalty on out, got %v", deltas["p_arjun_v"])
	}
}

func TestMultipliers_CaptainAndViceCaptain(t *testing.T) {
	team := matchbuddy.FantasyTeam{
		UserID:        "user1",
		PlayerIDs:     []string{"p_arjun_v", "p_vikram_k", "p_dev_p"},
		CaptainID:     "p_arjun_v",  // 2x
		ViceCaptainID: "p_vikram_k", // 1.5x
	}

	playerDeltas := map[string]float64{
		"p_arjun_v":  12.0, // Captain: 12 * 2 = 24
		"p_vikram_k": 8.0,  // VC: 8 * 1.5 = 12
		"p_dev_p":    5.0,  // Normal: 5 * 1 = 5
	}

	score := matchbuddy.TeamScore(team, playerDeltas)
	expected := (12.0 * 2.0) + (8.0 * 1.5) + (5.0 * 1.0) // 24 + 12 + 5 = 41
	if score != expected {
		t.Fatalf("expected team score %v, got %v", expected, score)
	}
}

// Golden test: runs the full 36-ball fixture and verifies stability & lead changes.
func TestSimulatedFixture_GoldenRun(t *testing.T) {
	presetBat := matchbuddy.Presets[0]
	presetBowl := matchbuddy.Presets[1]

	teams := []matchbuddy.FantasyTeam{
		{
			UserID:        "u_bat",
			Username:      "BattingFan",
			PlayerIDs:     presetBat.PlayerIDs,
			CaptainID:     presetBat.CaptainID,
			ViceCaptainID: presetBat.ViceCaptainID,
		},
		{
			UserID:        "u_bowl",
			Username:      "BowlingFan",
			PlayerIDs:     presetBowl.PlayerIDs,
			CaptainID:     presetBowl.CaptainID,
			ViceCaptainID: presetBowl.ViceCaptainID,
		},
	}

	stats := make(map[string]matchbuddy.PlayerRunningStats)
	totals := map[string]float64{
		"u_bat":  0,
		"u_bowl": 0,
	}

	leadFlips := 0
	lastLeader := ""

	for _, ball := range matchbuddy.SimulatedFixture {
		var memberDeltas map[string]float64
		var rankChanges []matchbuddy.RankChange

		memberDeltas, totals, _, stats, rankChanges = matchbuddy.ApplyBall(
			ball, stats, teams, totals,
		)

		_ = memberDeltas
		leader := rankChanges[0].UserID
		if lastLeader != "" && leader != lastLeader {
			leadFlips++
		}
		lastLeader = leader
	}

	t.Logf("Fixture completed! Final scores: BattingFan=%v, BowlingFan=%v, Lead flips=%d",
		totals["u_bat"], totals["u_bowl"], leadFlips)

	if len(matchbuddy.SimulatedFixture) != 36 {
		t.Fatalf("expected 36 balls in fixture, got %d", len(matchbuddy.SimulatedFixture))
	}

	if leadFlips == 0 {
		t.Errorf("expected at least one lead flip between the preset teams, got 0")
	}

	// Verify both teams scored meaningful points
	if totals["u_bat"] <= 50 || totals["u_bowl"] <= 50 {
		t.Errorf("expected healthy scores > 50, got bat=%v, bowl=%v", totals["u_bat"], totals["u_bowl"])
	}
}
