import React, { useState } from "react";
import type { TeamChange } from "../../types/matchbuddy";
import { History, ChevronDown, ChevronUp } from "lucide-react";
import styles from "./TeamChangesLog.module.css";

interface TeamChangesLogProps {
  changes: TeamChange[];
}

function formatChange(change: TeamChange): string {
  switch (change.kind) {
    case "team_created":
      return `${change.username} drafted lineup (${change.to})`;
    case "captain_change":
      return `${change.username} changed Captain to ${change.to}`;
    case "vc_change":
      return `${change.username} changed Vice-Captain to ${change.to}`;
    case "player_swap":
      return `${change.username} switched preset to ${change.to}`;
    default:
      return `${change.username} updated lineup`;
  }
}

export const TeamChangesLog: React.FC<TeamChangesLogProps> = ({ changes }) => {
  const [isOpen, setIsOpen] = useState(false);

  if (changes.length === 0) return null;

  return (
    <div className={styles.container}>
      <button
        type="button"
        className={styles.toggleBtn}
        onClick={() => setIsOpen((prev) => !prev)}
      >
        <span className="flex items-center gap-1.5">
          <History size={13} /> Lineup Activity ({changes.length})
        </span>
        {isOpen ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
      </button>

      {isOpen && (
        <div className={styles.logList}>
          {changes.slice().reverse().map((c, i) => (
            <div key={`${c.user_id}-${i}`} className={styles.logItem}>
              <span className={styles.logText}>{formatChange(c)}</span>
              <span className={styles.logTime}>
                {new Date(c.at).toLocaleTimeString([], {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
