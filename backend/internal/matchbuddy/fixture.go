package matchbuddy

import "time"

// PlayerRole denotes a cricket player's role in the fantasy lineup.
type PlayerRole string

const (
	RoleBat    PlayerRole = "BAT"
	RoleBowl   PlayerRole = "BOWL"
	RoleAllR   PlayerRole = "AR"
	RoleWicket PlayerRole = "WK"
)

// Player represents a fictional player in the simulated fixture.
type Player struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Team      string     `json:"team"` // e.g. "MB" or "DT"
	Role      PlayerRole `json:"role"`
	Credits   float64    `json:"credits"`
	AvatarURL string     `json:"avatar_url,omitempty"`
}

// WicketInfo details a dismissal.
type WicketInfo struct {
	Kind      string `json:"kind"` // "bowled", "caught", "lbw", "run_out"
	PlayerOut string `json:"player_out"`
	BowlerID  string `json:"bowler_id,omitempty"`
}

// BallEvent describes one legal or extra delivery in the simulated match.
type BallEvent struct {
	Index          int         `json:"index"` // 0-based sequence
	Over           int         `json:"over"`  // 1-based over (1 to 6)
	Ball           int         `json:"ball"`  // 1-based ball within over (1 to 6)
	StrikerID      string      `json:"striker_id"`
	NonStrikerID   string      `json:"non_striker_id"`
	BowlerID       string      `json:"bowler_id"`
	RunsOffBat     int         `json:"runs_off_bat"`
	Extras         int         `json:"extras"`
	ExtraType      string      `json:"extra_type"` // "none", "wide", "no_ball"
	IsBoundary4    bool        `json:"is_boundary_4"`
	IsSix          bool        `json:"is_six"`
	Wicket         *WicketInfo `json:"wicket,omitempty"`
	CommentaryText string      `json:"commentary_text"`
}

// FantasyTeam represents a user's chosen 11 with Captain and Vice-Captain.
type FantasyTeam struct {
	UserID        string     `json:"user_id"`
	Username      string     `json:"username"`
	PlayerIDs     []string   `json:"player_ids"`
	CaptainID     string     `json:"captain_id"`
	ViceCaptainID string     `json:"vice_captain_id"`
	LockedAt      *time.Time `json:"locked_at,omitempty"`
	IsBot         bool       `json:"is_bot,omitempty"`
	PresetID      string     `json:"preset_id,omitempty"`
}

// TeamChange logs pre-match adjustments.
type TeamChange struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Kind     string    `json:"kind"` // "player_swap", "captain_change", "vc_change", "team_created"
	From     string    `json:"from,omitempty"`
	To       string    `json:"to,omitempty"`
	At       time.Time `json:"at"`
}

// PresetDefinition holds pre-configured 11 selections for one-click entry.
type PresetDefinition struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	PlayerIDs     []string `json:"player_ids"`
	CaptainID     string   `json:"captain_id"`
	ViceCaptainID string   `json:"vice_captain_id"`
}

// Fictional teams:
const (
	TeamMumbai = "MB" // Mumbai Blasters
	TeamDelhi  = "DT" // Delhi Titans
)

// Squad represents all 16 fictional players available for drafting.
var Squad = []Player{
	// Mumbai Blasters (8 players)
	{ID: "p_arjun_v", Name: "Arjun Verma", Team: TeamMumbai, Role: RoleBat, Credits: 9.5},
	{ID: "p_rohan_s", Name: "Rohan Sharma", Team: TeamMumbai, Role: RoleBat, Credits: 9.0},
	{ID: "p_kabir_r", Name: "Kabir Ray", Team: TeamMumbai, Role: RoleWicket, Credits: 8.5},
	{ID: "p_dev_p", Name: "Dev Pandya", Team: TeamMumbai, Role: RoleAllR, Credits: 9.5},
	{ID: "p_tariq_m", Name: "Tariq Malik", Team: TeamMumbai, Role: RoleAllR, Credits: 8.0},
	{ID: "p_jas_b", Name: "Jaspreet Bir", Team: TeamMumbai, Role: RoleBowl, Credits: 9.0},
	{ID: "p_varun_c", Name: "Varun Chawla", Team: TeamMumbai, Role: RoleBowl, Credits: 8.5},
	{ID: "p_kunal_k", Name: "Kunal Kumble", Team: TeamMumbai, Role: RoleBowl, Credits: 7.5},

	// Delhi Titans (8 players)
	{ID: "p_vikram_k", Name: "Vikram Kohli", Team: TeamDelhi, Role: RoleBat, Credits: 9.5},
	{ID: "p_ishan_g", Name: "Ishan Gill", Team: TeamDelhi, Role: RoleBat, Credits: 8.5},
	{ID: "p_rahul_t", Name: "Rahul Tiwari", Team: TeamDelhi, Role: RoleBat, Credits: 8.0},
	{ID: "p_aarav_p", Name: "Aarav Pant", Team: TeamDelhi, Role: RoleWicket, Credits: 9.0},
	{ID: "p_samir_k", Name: "Samir Khan", Team: TeamDelhi, Role: RoleAllR, Credits: 9.0},
	{ID: "p_neeraj_y", Name: "Neeraj Yadav", Team: TeamDelhi, Role: RoleBowl, Credits: 9.0},
	{ID: "p_anand_s", Name: "Anand Shami", Team: TeamDelhi, Role: RoleBowl, Credits: 8.5},
	{ID: "p_zayn_a", Name: "Zayn Akhtar", Team: TeamDelhi, Role: RoleBowl, Credits: 8.5},
}

// SquadByID maps player ID to Player for quick lookup.
var SquadByID = func() map[string]Player {
	m := make(map[string]Player, len(Squad))
	for _, p := range Squad {
		m[p.ID] = p
	}
	return m
}()

// Presets provides 3 distinct strategic choices.
var Presets = []PresetDefinition{
	{
		ID:          "preset_batting",
		Name:        "Aggressive Batting",
		Description: "Top-order powerhouse with Arjun (C) and Vikram (VC)",
		PlayerIDs: []string{
			"p_arjun_v", "p_rohan_s", "p_kabir_r", "p_dev_p", "p_tariq_m", "p_jas_b", // 6 MB
			"p_vikram_k", "p_ishan_g", "p_aarav_p", "p_samir_k", "p_neeraj_y",         // 5 DT
		},
		CaptainID:     "p_arjun_v",
		ViceCaptainID: "p_vikram_k",
	},
	{
		ID:          "preset_bowling",
		Name:        "Bowler Heavy",
		Description: "Wicket hunters with Neeraj (C) and Zayn (VC)",
		PlayerIDs: []string{
			"p_dev_p", "p_jas_b", "p_varun_c", "p_kunal_k", "p_arjun_v", // 5 MB
			"p_vikram_k", "p_samir_k", "p_neeraj_y", "p_anand_s", "p_zayn_a", "p_aarav_p", // 6 DT
		},
		CaptainID:     "p_neeraj_y",
		ViceCaptainID: "p_zayn_a",
	},
	{
		ID:          "preset_balanced",
		Name:        "Balanced All-Round",
		Description: "Well-rounded squad captained by Dev Pandya",
		PlayerIDs: []string{
			"p_arjun_v", "p_rohan_s", "p_dev_p", "p_jas_b", "p_varun_c", "p_kabir_r", // 6 MB
			"p_vikram_k", "p_aarav_p", "p_samir_k", "p_neeraj_y", "p_anand_s",         // 5 DT
		},
		CaptainID:     "p_dev_p",
		ViceCaptainID: "p_anand_s",
	},
}

// PresetByID provides lookup by preset ID.
func PresetByID(id string) *PresetDefinition {
	for _, p := range Presets {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// SimulatedFixture delivers a dramatic 6-over (36-delivery) T20 innings excerpt.
// Batting: Mumbai Blasters (MB) | Bowling: Delhi Titans (DT)
//
// Key drama highlights:
// - Over 1.2: Four by Arjun Verma.
// - Over 1.4: Huge SIX by Arjun Verma!
// - Over 2.2: No-ball by Anand Shami.
// - Over 2.5 - 3.2: 4-ball dot streak creating tension.
// - Over 3.1: Wicket! Kabir Ray dismissed for a DUCK (-2 pts).
// - Over 4.3: Wicket! Captain pick Arjun Verma dismissed for 24, caught off Zayn Akhtar! Lead swings to Bowler preset!
// - Over 4.5: Wide ball by Zayn Akhtar.
// - Over 5.3: Massive SIX by Dev Pandya!
// - Over 6.2: Four by Dev Pandya!
// - Over 6.4: Wicket of bowler Kunal Kumble.
// - Over 6.5: Dev Pandya crosses 30 runs with a SIX, triggering +8 milestone bonus! Lead flips back to Batting preset!
var SimulatedFixture = []BallEvent{
	// ── OVER 1 (Bowler: Neeraj Yadav, Batters: Arjun Verma & Rohan Sharma) ──
	{
		Index: 0, Over: 1, Ball: 1, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_neeraj_y",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Neeraj opens with a sharp good-length delivery. Arjun works it into the gap for a single.",
	},
	{
		Index: 1, Over: 1, Ball: 2, StrikerID: "p_rohan_s", NonStrikerID: "p_arjun_v", BowlerID: "p_neeraj_y",
		RunsOffBat: 4, Extras: 0, ExtraType: "none", IsBoundary4: true,
		CommentaryText: "FOUR! Rohan Sharma leans into a gorgeous cover drive, piercing the infield effortlessly!",
	},
	{
		Index: 2, Over: 1, Ball: 3, StrikerID: "p_rohan_s", NonStrikerID: "p_arjun_v", BowlerID: "p_neeraj_y",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Tucked away off the pads to deep square leg for one.",
	},
	{
		Index: 3, Over: 1, Ball: 4, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_neeraj_y",
		RunsOffBat: 6, Extras: 0, ExtraType: "none", IsSix: true,
		CommentaryText: "SIX! Stand and deliver! Arjun Verma launches it 88 meters over long-on for maximum!",
	},
	{
		Index: 4, Over: 1, Ball: 5, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_neeraj_y",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		CommentaryText: "Beaten by the bounce! Neeraj pulls his length back, beating the outside edge.",
	},
	{
		Index: 5, Over: 1, Ball: 6, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_neeraj_y",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Pushed to mid-off to retain the strike. End of over 1: 13/0.",
	},

	// ── OVER 2 (Bowler: Anand Shami, Batters: Arjun Verma & Rohan Sharma) ──
	{
		Index: 6, Over: 2, Ball: 1, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_anand_s",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Anand Shami steams in; Arjun drops it softly to midwicket for a quick single.",
	},
	{
		Index: 7, Over: 2, Ball: 2, StrikerID: "p_rohan_s", NonStrikerID: "p_arjun_v", BowlerID: "p_anand_s",
		RunsOffBat: 1, Extras: 1, ExtraType: "no_ball",
		CommentaryText: "NO BALL! Shami oversteps the crease! Rohan gets a single, and Free Hit coming up!",
	},
	{
		Index: 8, Over: 2, Ball: 3, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_anand_s",
		RunsOffBat: 4, Extras: 0, ExtraType: "none", IsBoundary4: true,
		CommentaryText: "FOUR! Free Hit dispatched! Arjun crunches a short ball over mid-wicket for four!",
	},
	{
		Index: 9, Over: 2, Ball: 4, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_anand_s",
		RunsOffBat: 2, Extras: 0, ExtraType: "none",
		CommentaryText: "Great running between the wickets, turning one into two with aggressive calling.",
	},
	{
		Index: 10, Over: 2, Ball: 5, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_anand_s",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		CommentaryText: "Dot ball. Excellent yorker right at the blockhole, dug out safely.",
	},
	{
		Index: 11, Over: 2, Ball: 6, StrikerID: "p_arjun_v", NonStrikerID: "p_rohan_s", BowlerID: "p_anand_s",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		CommentaryText: "Another dot! Shami follows the batter and cramps him for room. End of over 2: 22/0.",
	},

	// ── OVER 3 (Bowler: Samir Khan, Batters: Rohan Sharma & Kabir Ray) ──
	{
		Index: 12, Over: 3, Ball: 1, StrikerID: "p_rohan_s", NonStrikerID: "p_arjun_v", BowlerID: "p_samir_k",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		Wicket: &WicketInfo{Kind: "caught", PlayerOut: "p_rohan_s", BowlerID: "p_samir_k"},
		CommentaryText: "WICKET! Samir Khan strikes with his very first ball! Rohan Sharma chips to mid-on, caught easily!",
	},
	{
		Index: 13, Over: 3, Ball: 2, StrikerID: "p_kabir_r", NonStrikerID: "p_arjun_v", BowlerID: "p_samir_k",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		Wicket: &WicketInfo{Kind: "bowled", PlayerOut: "p_kabir_r", BowlerID: "p_samir_k"},
		CommentaryText: "WICKET! TWO IN TWO! Golden duck for Kabir Ray, clean bowled through the gate! Duck penalty applies!",
	},
	{
		Index: 14, Over: 3, Ball: 3, StrikerID: "p_dev_p", NonStrikerID: "p_arjun_v", BowlerID: "p_samir_k",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		CommentaryText: "Hat-trick ball defended solidly! Dev Pandya blocks right back to the bowler.",
	},
	{
		Index: 15, Over: 3, Ball: 4, StrikerID: "p_dev_p", NonStrikerID: "p_arjun_v", BowlerID: "p_samir_k",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Dev taps to backward point to get off the mark and rotate strike.",
	},
	{
		Index: 16, Over: 3, Ball: 5, StrikerID: "p_arjun_v", NonStrikerID: "p_dev_p", BowlerID: "p_samir_k",
		RunsOffBat: 2, Extras: 0, ExtraType: "none",
		CommentaryText: "Arjun drives past extra cover and scampers back for a brace.",
	},
	{
		Index: 17, Over: 3, Ball: 6, StrikerID: "p_arjun_v", NonStrikerID: "p_dev_p", BowlerID: "p_samir_k",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Worked into the on-side for a single. End of over 3: 26/2. Double-strike from Samir Khan!",
	},

	// ── OVER 4 (Bowler: Zayn Akhtar, Batters: Arjun Verma & Dev Pandya) ──
	{
		Index: 18, Over: 4, Ball: 1, StrikerID: "p_arjun_v", NonStrikerID: "p_dev_p", BowlerID: "p_zayn_a",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Zayn enters the attack with some flight; Arjun pushes to long-off for one.",
	},
	{
		Index: 19, Over: 4, Ball: 2, StrikerID: "p_dev_p", NonStrikerID: "p_arjun_v", BowlerID: "p_zayn_a",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Dev works the leg-spinner to deep backward square for a single.",
	},
	{
		Index: 20, Over: 4, Ball: 3, StrikerID: "p_arjun_v", NonStrikerID: "p_dev_p", BowlerID: "p_zayn_a",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		Wicket: &WicketInfo{Kind: "caught", PlayerOut: "p_arjun_v", BowlerID: "p_zayn_a"},
		CommentaryText: "WICKET! HUGE MOMENT! Captain Arjun Verma falls for 24, caught at long-on! Fantasy leaderboards in turmoil!",
	},
	{
		Index: 21, Over: 4, Ball: 4, StrikerID: "p_tariq_m", NonStrikerID: "p_dev_p", BowlerID: "p_zayn_a",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		CommentaryText: "New batter Tariq Malik plays and misses at a sharp turning leg-break.",
	},
	{
		Index: 22, Over: 4, Ball: 5, StrikerID: "p_tariq_m", NonStrikerID: "p_dev_p", BowlerID: "p_zayn_a",
		RunsOffBat: 0, Extras: 1, ExtraType: "wide",
		CommentaryText: "WIDE! Spilled down the leg side, wicketkeeper Aarav Pant tidies it up.",
	},
	{
		Index: 23, Over: 4, Ball: 6, StrikerID: "p_tariq_m", NonStrikerID: "p_dev_p", BowlerID: "p_zayn_a",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Tariq guides it to third man for a single to get off the mark. End of over 4: 30/3.",
	},

	// ── OVER 5 (Bowler: Neeraj Yadav, Batters: Tariq Malik & Dev Pandya) ──
	{
		Index: 24, Over: 5, Ball: 1, StrikerID: "p_tariq_m", NonStrikerID: "p_dev_p", BowlerID: "p_neeraj_y",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Neeraj returns for his second over; Tariq dabs a single to point.",
	},
	{
		Index: 25, Over: 5, Ball: 2, StrikerID: "p_dev_p", NonStrikerID: "p_tariq_m", BowlerID: "p_neeraj_y",
		RunsOffBat: 2, Extras: 0, ExtraType: "none",
		CommentaryText: "Dev Pandya clips firmly between mid-wicket and long-on for a sharp two.",
	},
	{
		Index: 26, Over: 5, Ball: 3, StrikerID: "p_dev_p", NonStrikerID: "p_tariq_m", BowlerID: "p_neeraj_y",
		RunsOffBat: 6, Extras: 0, ExtraType: "none", IsSix: true,
		CommentaryText: "SIX! MONSTER HIT! Dev Pandya smokes a 92m rocket over wide long-on into the top tier!",
	},
	{
		Index: 27, Over: 5, Ball: 4, StrikerID: "p_dev_p", NonStrikerID: "p_tariq_m", BowlerID: "p_neeraj_y",
		RunsOffBat: 4, Extras: 0, ExtraType: "none", IsBoundary4: true,
		CommentaryText: "FOUR! Pandya is in demolition mode! Driven straight back past the non-striker for four!",
	},
	{
		Index: 28, Over: 5, Ball: 5, StrikerID: "p_dev_p", NonStrikerID: "p_tariq_m", BowlerID: "p_neeraj_y",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Neeraj fires a yorker; Dev squeezes it out to cover for one.",
	},
	{
		Index: 29, Over: 5, Ball: 6, StrikerID: "p_tariq_m", NonStrikerID: "p_dev_p", BowlerID: "p_neeraj_y",
		RunsOffBat: 2, Extras: 0, ExtraType: "none",
		CommentaryText: "Tariq turns it into the vacant deep square gap for two. End of over 5: 46/3. Big 16-run over!",
	},

	// ── OVER 6 (Bowler: Anand Shami, Batters: Dev Pandya & Tariq Malik) ──
	{
		Index: 30, Over: 6, Ball: 1, StrikerID: "p_dev_p", NonStrikerID: "p_tariq_m", BowlerID: "p_anand_s",
		RunsOffBat: 2, Extras: 0, ExtraType: "none",
		CommentaryText: "Final over! Shami bowls full; Dev whips it through midwicket for two runs.",
	},
	{
		Index: 31, Over: 6, Ball: 2, StrikerID: "p_dev_p", NonStrikerID: "p_tariq_m", BowlerID: "p_anand_s",
		RunsOffBat: 4, Extras: 0, ExtraType: "none", IsBoundary4: true,
		CommentaryText: "FOUR! Slice of genius! Dev cuts late over backward point for a boundary!",
	},
	{
		Index: 32, Over: 6, Ball: 3, StrikerID: "p_dev_p", NonStrikerID: "p_tariq_m", BowlerID: "p_anand_s",
		RunsOffBat: 1, Extras: 0, ExtraType: "none",
		CommentaryText: "Dug out down to long-off for a single.",
	},
	{
		Index: 33, Over: 6, Ball: 4, StrikerID: "p_tariq_m", NonStrikerID: "p_dev_p", BowlerID: "p_anand_s",
		RunsOffBat: 0, Extras: 0, ExtraType: "none",
		Wicket: &WicketInfo{Kind: "bowled", PlayerOut: "p_tariq_m", BowlerID: "p_anand_s"},
		CommentaryText: "WICKET! Shami shatters the stumps with an unplayable 145kph toe-crusher! Tariq is out!",
	},
	{
		Index: 34, Over: 6, Ball: 5, StrikerID: "p_dev_p", NonStrikerID: "p_kunal_k", BowlerID: "p_anand_s",
		RunsOffBat: 6, Extras: 0, ExtraType: "none", IsSix: true,
		CommentaryText: "SIX! UNBELIEVABLE! Dev Pandya smashes a maximum and crosses 30 runs! 30-run bonus unlocked!",
	},
	{
		Index: 35, Over: 6, Ball: 6, StrikerID: "p_dev_p", NonStrikerID: "p_kunal_k", BowlerID: "p_anand_s",
		RunsOffBat: 2, Extras: 0, ExtraType: "none",
		CommentaryText: "Finished in style! Dev pushes to deep cover for two. Mumbai Blasters finish on 61/4!",
	},
}
