export type PlayerRole = "BAT" | "BOWL" | "AR" | "WK";

export interface Player {
  id: string;
  name: string;
  team: string; // "MB" | "DT"
  role: PlayerRole;
  credits: number;
  avatar_url?: string;
}

export interface WicketInfo {
  kind: string;
  player_out: string;
  bowler_id?: string;
}

export interface BallEvent {
  index: number;
  over: number;
  ball: number;
  striker_id: string;
  non_striker_id: string;
  bowler_id: string;
  runs_off_bat: number;
  extras: number;
  extra_type: string;
  is_boundary_4: boolean;
  is_six: boolean;
  wicket?: WicketInfo;
  commentary_text: string;
}

export interface FantasyTeam {
  user_id: string;
  username: string;
  player_ids: string[];
  captain_id: string;
  vice_captain_id: string;
  locked_at?: string;
  is_bot?: boolean;
  preset_id?: string;
}

export interface TeamChange {
  user_id: string;
  username: string;
  kind: "player_swap" | "captain_change" | "vc_change" | "team_created";
  from?: string;
  to?: string;
  at: string;
}

export interface PresetDefinition {
  id: string;
  name: string;
  description: string;
  player_ids: string[];
  captain_id: string;
  vice_captain_id: string;
}

export type MatchStatus = "lobby" | "playing" | "paused" | "completed";

export interface LeaderboardEntry {
  user_id: string;
  username: string;
  points: number;
  rank: number;
  captain_name: string;
  vc_name: string;
  is_bot?: boolean;
}

export interface RankChange {
  user_id: string;
  username: string;
  old_rank: number;
  new_rank: number;
  points_delta: number;
  new_total: number;
}

export type FeedCardKind =
  | "swing"
  | "milestone"
  | "over_summary"
  | "whatif"
  | "answer"
  | "catchup"
  | "bot"
  | "team_change";

export interface FeedCard {
  id: string;
  kind: FeedCardKind;
  headline: string;
  body_text: string;
  over?: number;
  ball?: number;
  timestamp: number;
  deltas?: Record<string, number>;
}

export interface MatchStatePayload {
  status: MatchStatus;
  speed_multiplier: number;
  current_ball_idx: number;
  total_balls: number;
  runs: number;
  wickets: number;
  overs: number;
  balls_in_over: number;
  striker?: Player;
  non_striker?: Player;
  bowler?: Player;
  recent_balls: BallEvent[];
  teams_locked: boolean;
  leaderboard: LeaderboardEntry[];
  teams: Record<string, FantasyTeam>;
  has_bot: boolean;
  squad: Player[];
  presets: PresetDefinition[];
}

export interface MatchBallPayload {
  ball: BallEvent;
  player_deltas: Record<string, number>;
  member_deltas: Record<string, number>;
  new_totals: Record<string, number>;
  rank_changes: RankChange[];
  score_str: string;
  overs_str: string;
}

export interface WhatIfProjection {
  scenario: string;
  striker_name: string;
  bowler_name: string;
  member_projections: RankChange[];
  explanation: string;
}
