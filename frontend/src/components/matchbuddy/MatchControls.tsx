import React from "react";
import type { MatchStatus } from "../../types/matchbuddy";
import { Play, Pause, FastForward, RotateCcw, Bot, Users } from "lucide-react";
import styles from "./MatchControls.module.css";

interface MatchControlsProps {
  status: MatchStatus;
  speed: number;
  hasBot: boolean;
  teamsLocked: boolean;
  onStart: () => void;
  onPause: () => void;
  onResume: () => void;
  onSetSpeed: (speed: number) => void;
  onSkipOver: () => void;
  onReset: () => void;
  onToggleSolo: () => void;
  onOpenTeamBuilder: () => void;
}

export const MatchControls: React.FC<MatchControlsProps> = ({
  status,
  speed,
  hasBot,
  teamsLocked,
  onStart,
  onPause,
  onResume,
  onSetSpeed,
  onSkipOver,
  onReset,
  onToggleSolo,
  onOpenTeamBuilder,
}) => {
  const isPlaying = status === "playing";
  const isPaused = status === "paused";
  const isLobby = status === "lobby";
  const isCompleted = status === "completed";

  return (
    <div className={styles.controlsGrid}>
      <div className={styles.primaryRow}>
        {isLobby && (
          <button type="button" className={styles.mainActionBtn} onClick={onStart}>
            <Play size={16} fill="currentColor" /> Lock & Start Replay
          </button>
        )}

        {isPlaying && (
          <button type="button" className={styles.mainActionBtn} onClick={onPause}>
            <Pause size={16} fill="currentColor" /> Pause Replay
          </button>
        )}

        {isPaused && (
          <button type="button" className={styles.mainActionBtn} onClick={onResume}>
            <Play size={16} fill="currentColor" /> Resume Replay
          </button>
        )}

        {isCompleted && (
          <button type="button" className={styles.mainActionBtn} onClick={onReset}>
            <RotateCcw size={16} /> Replay Again
          </button>
        )}

        <button
          type="button"
          className={styles.teamBtn}
          onClick={onOpenTeamBuilder}
          title={teamsLocked ? "Inspect Lineups" : "Edit Fantasy 11"}
        >
          <Users size={15} />
          {teamsLocked ? "View 11" : "Edit 11"}
        </button>
      </div>

      <div className={styles.secondaryRow}>
        {/* Speed Pills */}
        <div className={styles.speedGroup}>
          {[1.0, 2.0, 5.0].map((s) => (
            <button
              key={s}
              type="button"
              className={`${styles.speedBtn} ${speed === s ? styles.speedBtnActive : ""}`}
              onClick={() => onSetSpeed(s)}
            >
              {s}x
            </button>
          ))}
        </div>

        {/* Action buttons */}
        <div className={styles.utilityGroup}>
          <button
            type="button"
            className={`${styles.utilBtn} ${hasBot ? styles.soloActive : ""}`}
            onClick={onToggleSolo}
            title="Adds Rahul (bot) to compare points against a simulated friend"
          >
            <Bot size={13} /> {hasBot ? "Bot: Rahul (On)" : "Try it solo"}
          </button>

          {!isCompleted && !isLobby && (
            <button
              type="button"
              className={styles.utilBtn}
              onClick={onSkipOver}
              title="Fast-forward to next over"
            >
              <FastForward size={13} /> Skip Over
            </button>
          )}

          {!isLobby && (
            <button
              type="button"
              className={styles.utilBtn}
              onClick={onReset}
              title="Reset match to lobby"
            >
              <RotateCcw size={13} /> Reset
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
