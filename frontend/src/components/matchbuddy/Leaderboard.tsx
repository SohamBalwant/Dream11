import React from "react";
import type { LeaderboardEntry } from "../../types/matchbuddy";
import { Trophy, Bot } from "lucide-react";
import styles from "./Leaderboard.module.css";

interface LeaderboardProps {
  entries: LeaderboardEntry[];
  recentDeltas?: Record<string, number>;
}

export const Leaderboard: React.FC<LeaderboardProps> = ({
  entries,
  recentDeltas = {},
}) => {
  if (entries.length === 0) {
    return (
      <div className={styles.container}>
        <div className={styles.headerRow}>
          <span className={styles.title}>Fantasy Leaderboard</span>
        </div>
        <div className="p-4 text-center text-sm text-[var(--color-text-muted)] bg-[var(--color-surface-1)] rounded-lg border border-[var(--color-border-subtle)]">
          No teams drafted yet. Pick your 11 or select a preset to enter!
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.headerRow}>
        <span className={styles.title}>Fantasy Leaderboard</span>
        <span className="text-xs text-[var(--color-text-muted)]">
          {entries.length} participant{entries.length === 1 ? "" : "s"}
        </span>
      </div>

      <div className={styles.teamList}>
        {entries.map((entry) => {
          const delta = recentDeltas[entry.user_id];
          const hasDelta = delta !== undefined && delta !== 0;
          const isRank1 = entry.rank === 1;

          return (
            <div
              key={entry.user_id}
              className={`${styles.memberCard} ${isRank1 ? styles.rank1 : ""}`}
            >
              <div className={styles.leftCluster}>
                <div
                  className={`${styles.rankBadge} ${
                    isRank1 ? styles.rank1Badge : ""
                  }`}
                >
                  {isRank1 ? <Trophy size={13} /> : `#${entry.rank}`}
                </div>

                <div className={styles.userInfo}>
                  <div className={styles.nameRow}>
                    <span className={styles.username}>{entry.username}</span>
                    {entry.is_bot && (
                      <span className={styles.botTag}>
                        <Bot size={10} className="inline mr-0.5" /> Bot
                      </span>
                    )}
                  </div>
                  <div className={styles.picksRow}>
                    <span className={styles.pickItem}>
                      <span className={styles.badgeC}>C</span> {entry.captain_name}
                    </span>
                    <span>·</span>
                    <span className={styles.pickItem}>
                      <span className={styles.badgeVC}>VC</span> {entry.vc_name}
                    </span>
                  </div>
                </div>
              </div>

              <div className={styles.rightCluster}>
                <span className={styles.points}>
                  {entry.points.toFixed(1)} <span className="text-xs text-[var(--color-text-muted)]">pts</span>
                </span>
                {hasDelta && (
                  <span
                    className={`${styles.deltaBadge} ${
                      delta > 0 ? styles.deltaPositive : styles.deltaNegative
                    }`}
                  >
                    {delta > 0 ? `+${delta.toFixed(1)}` : delta.toFixed(1)}
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
