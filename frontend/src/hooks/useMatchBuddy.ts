import { useState, useEffect, useCallback, useRef } from "react";
import { wsService } from "../services/ws/ws.service";
import type {
  MatchStatePayload,
  MatchBallPayload,
  FeedCard,
  TeamChange,
} from "../types/matchbuddy";
import type { WSMessage } from "../types/ws";

export function useMatchBuddy(roomId?: string) {
  const [matchState, setMatchState] = useState<MatchStatePayload | null>(null);
  const [feedCards, setFeedCards] = useState<FeedCard[]>([]);
  const [teamChanges, setTeamChanges] = useState<TeamChange[]>([]);
  const [lastBall, setLastBall] = useState<MatchBallPayload | null>(null);
  const [recentDeltas, setRecentDeltas] = useState<Record<string, number>>({});
  const deltaTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!roomId) return;

    const unsubs = [
      wsService.on("MATCH_STATE", (msg: WSMessage<MatchStatePayload>) => {
        if (msg.payload) {
          setMatchState(msg.payload);
        }
      }),

      wsService.on("MATCH_BALL", (msg: WSMessage<MatchBallPayload>) => {
        if (msg.payload) {
          const payload = msg.payload;
          setLastBall(payload);
          setRecentDeltas(payload.member_deltas || {});

          if (deltaTimerRef.current) clearTimeout(deltaTimerRef.current);
          deltaTimerRef.current = setTimeout(() => {
            setRecentDeltas({});
          }, 3500);

          setMatchState((prev) => {
            if (!prev) return prev;
            // Update leaderboard & score in snapshot
            const updatedLeaderboard = prev.leaderboard.map((entry) => {
              const newPts = payload.new_totals[entry.user_id];
              return newPts !== undefined ? { ...entry, points: newPts } : entry;
            });
            // Re-sort leaderboard
            updatedLeaderboard.sort((a, b) => b.points - a.points);
            updatedLeaderboard.forEach((e, idx) => {
              e.rank = idx + 1;
            });

            return {
              ...prev,
              current_ball_idx: payload.ball.index + 1,
              runs: (payload.ball.over - 1) * 6 + payload.ball.ball, // or authoritative from state
              recent_balls: [...prev.recent_balls.slice(-5), payload.ball],
              leaderboard: updatedLeaderboard,
            };
          });
        }
      }),

      wsService.on("MATCH_FEED", (msg: WSMessage<FeedCard>) => {
        if (msg.payload) {
          const card = msg.payload;
          setFeedCards((prev) => [card, ...prev.slice(0, 49)]); // keep newest 50
        }
      }),

      wsService.on("MATCH_TEAM_CHANGES", (msg: WSMessage<TeamChange[]>) => {
        if (msg.payload && Array.isArray(msg.payload)) {
          setTeamChanges(msg.payload);
        }
      }),
    ];

    return () => {
      unsubs.forEach((u) => u());
      if (deltaTimerRef.current) clearTimeout(deltaTimerRef.current);
    };
  }, [roomId]);

  // Actions
  const startMatch = useCallback(() => {
    wsService.send("MATCH_CONTROL", { action: "start" });
  }, []);

  const pauseMatch = useCallback(() => {
    wsService.send("MATCH_CONTROL", { action: "pause" });
  }, []);

  const resumeMatch = useCallback(() => {
    wsService.send("MATCH_CONTROL", { action: "resume" });
  }, []);

  const setSpeed = useCallback((speed: number) => {
    wsService.send("MATCH_CONTROL", { action: "set_speed", speed });
  }, []);

  const skipOver = useCallback(() => {
    wsService.send("MATCH_CONTROL", { action: "skip_over" });
  }, []);

  const resetMatch = useCallback(() => {
    wsService.send("MATCH_CONTROL", { action: "reset" });
    setFeedCards([]);
    setLastBall(null);
    setRecentDeltas({});
  }, []);

  const toggleSolo = useCallback(() => {
    wsService.send("MATCH_CONTROL", { action: "try_solo" });
  }, []);

  const setTeam = useCallback(
    (
      playerIds?: string[],
      captainId?: string,
      viceCaptainId?: string,
      presetId?: string,
    ) => {
      wsService.send("MATCH_TEAM_SET", {
        player_ids: playerIds,
        captain_id: captainId,
        vice_captain_id: viceCaptainId,
        preset_id: presetId,
      });
    },
    [],
  );

  const askBuddy = useCallback((question: string) => {
    if (!question.trim()) return;
    wsService.send("MATCH_ASK", { question: question.trim() });
  }, []);

  const whatIf = useCallback((scenario: "six" | "four" | "wicket" | "dot") => {
    wsService.send("MATCH_WHATIF", { scenario });
  }, []);

  const catchUp = useCallback(() => {
    wsService.send("MATCH_CATCHUP", {});
  }, []);

  return {
    matchState,
    feedCards,
    teamChanges,
    lastBall,
    recentDeltas,
    startMatch,
    pauseMatch,
    resumeMatch,
    setSpeed,
    skipOver,
    resetMatch,
    toggleSolo,
    setTeam,
    askBuddy,
    whatIf,
    catchUp,
  };
}
