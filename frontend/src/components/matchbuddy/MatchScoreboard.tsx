import React from "react";
import type { MatchStatePayload, BallEvent } from "../../types/matchbuddy";
import { Sparkles, Radio } from "lucide-react";
import styles from "./MatchScoreboard.module.css";

interface MatchScoreboardProps {
  state: MatchStatePayload | null;
}

function renderChip(ball: BallEvent, idx: number) {
  let chipClass = styles.ballChip;
  let text = String(ball.runs_off_bat);

  if (ball.wicket) {
    chipClass = `${styles.ballChip} ${styles.chipWicket}`;
    text = "W";
  } else if (ball.is_six) {
    chipClass = `${styles.ballChip} ${styles.chipSix}`;
    text = "6";
  } else if (ball.is_boundary_4) {
    chipClass = `${styles.ballChip} ${styles.chipFour}`;
    text = "4";
  } else if (ball.extra_type === "wide") {
    text = "Wd";
  } else if (ball.extra_type === "no_ball") {
    text = "Nb";
  } else if (ball.runs_off_bat === 0) {
    chipClass = `${styles.ballChip} ${styles.chipDot}`;
    text = "•";
  }

  return (
    <span key={`ball-${ball.index}-${idx}`} className={chipClass} title={ball.commentary_text}>
      {text}
    </span>
  );
}

export const MatchScoreboard: React.FC<MatchScoreboardProps> = ({ state }) => {
  if (!state) {
    return (
      <div className={styles.scoreboard}>
        <div className={styles.topRow}>
          <span className={styles.demoBadge}>
            <Sparkles size={11} /> Simulated demo match
          </span>
          <span className={styles.statusPill}>Connecting...</span>
        </div>
      </div>
    );
  }

  const runs = state.runs;
  const wickets = state.wickets;
  const overs = state.overs;
  const ballsInOver = state.balls_in_over;
  const totalOversDec = overs + ballsInOver / 6.0;
  const crr = totalOversDec > 0 ? (runs / totalOversDec).toFixed(2) : "0.00";

  return (
    <div className={styles.scoreboard}>
      <div className={styles.topRow}>
        <span className={styles.demoBadge}>
          <Sparkles size={11} /> Simulated demo match
        </span>
        <span
          className={`${styles.statusPill} ${
            state.status === "playing" ? styles.statusPlaying : ""
          }`}
        >
          {state.status === "playing" && <Radio size={10} className="inline mr-1" />}
          {state.status.toUpperCase()} {state.speed_multiplier > 1 ? `(${state.speed_multiplier}x)` : ""}
        </span>
      </div>

      <div className={styles.mainStats}>
        <div className={styles.scoreGroup}>
          <span className={styles.teamName}>MB</span>
          <span className={styles.score}>
            {runs}/{wickets}
          </span>
          <span className={styles.overs}>
            ({overs}.{ballsInOver} / 6 ov)
          </span>
        </div>
        <span className={styles.runRate}>CRR: {crr}</span>
      </div>

      <div className={styles.creaseInfo}>
        <div>
          <div className={styles.playerRole}>Striker 🏏</div>
          <div className={styles.playerName}>
            {state.striker?.name || "Arjun Verma*"}
          </div>
        </div>
        <div>
          <div className={styles.playerRole}>Bowler 🎯</div>
          <div className={styles.playerName}>
            {state.bowler?.name || "Neeraj Yadav"}
          </div>
        </div>
      </div>

      {state.recent_balls.length > 0 && (
        <div className={styles.recentBallsRow}>
          <span className={styles.recentLabel}>Recent:</span>
          {state.recent_balls.map((b, i) => renderChip(b, i))}
        </div>
      )}
    </div>
  );
};
