import React from "react";
import type { FeedCard } from "../../types/matchbuddy";
import { Zap, Award, BarChart3, HelpCircle, MessageSquareText, Bot } from "lucide-react";
import styles from "./FeedCardItem.module.css";

interface FeedCardItemProps {
  card: FeedCard;
  /** Maps user_id to username for clean deltas display */
  usernames?: Record<string, string>;
}

export const FeedCardItem: React.FC<FeedCardItemProps> = ({
  card,
  usernames = {},
}) => {
  let cardClass = styles.card;
  let Icon = Zap;
  let tagLabel = "Live Update";

  switch (card.kind) {
    case "swing":
      cardClass = `${styles.card} ${styles.cardSwing}`;
      Icon = Zap;
      tagLabel = "Points Swing";
      break;
    case "milestone":
      cardClass = `${styles.card} ${styles.cardMilestone}`;
      Icon = Award;
      tagLabel = "Milestone";
      break;
    case "over_summary":
      cardClass = `${styles.card}`;
      Icon = BarChart3;
      tagLabel = "Over Summary";
      break;
    case "whatif":
      cardClass = `${styles.card} ${styles.cardWhatIf}`;
      Icon = HelpCircle;
      tagLabel = "What If";
      break;
    case "answer":
      cardClass = `${styles.card} ${styles.cardAnswer}`;
      Icon = MessageSquareText;
      tagLabel = "Match Buddy";
      break;
    case "bot":
      cardClass = `${styles.card} ${styles.cardBot}`;
      Icon = Bot;
      tagLabel = "Rahul (bot)";
      break;
  }

  const timeStr = card.timestamp
    ? new Date(card.timestamp).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      })
    : "";

  return (
    <div className={cardClass}>
      <div className={styles.topRow}>
        <div className={styles.tagCluster}>
          <Icon size={12} className="text-zinc-400" />
          <span className={styles.tag}>{tagLabel}</span>
          {card.over !== undefined && card.ball !== undefined && (
            <span className="text-[11px] font-mono text-[var(--color-text-muted)]">
              {card.over}.{card.ball} ov
            </span>
          )}
        </div>
        <span className={styles.time}>{timeStr}</span>
      </div>

      <div className={styles.headline}>{card.headline}</div>
      <div className={styles.bodyText}>{card.body_text}</div>

      {card.deltas && Object.keys(card.deltas).length > 0 && (
        <div className={styles.deltasRow}>
          {Object.entries(card.deltas).map(([uid, delta]) => {
            const name = usernames[uid] || uid;
            return (
              <span key={uid} className={styles.deltaItem}>
                {name}:{" "}
                <span className={delta >= 0 ? styles.deltaPlus : "text-rose-400"}>
                  {delta >= 0 ? `+${delta.toFixed(1)}` : delta.toFixed(1)}
                </span>
              </span>
            );
          })}
        </div>
      )}
    </div>
  );
};
