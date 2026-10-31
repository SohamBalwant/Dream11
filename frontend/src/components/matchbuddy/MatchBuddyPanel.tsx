import React, { useState } from "react";
import { useMatchBuddy } from "../../hooks/useMatchBuddy";
import { useAuth } from "../../hooks/useAuth";
import { MatchScoreboard } from "./MatchScoreboard";
import { Leaderboard } from "./Leaderboard";
import { MatchControls } from "./MatchControls";
import { WhatIfChips } from "./WhatIfChips";
import { AskBuddyInput } from "./AskBuddyInput";
import { FeedCardItem } from "./FeedCardItem";
import { TeamChangesLog } from "./TeamChangesLog";
import { TeamBuilderModal } from "./TeamBuilderModal";
import { Activity } from "lucide-react";
import styles from "./MatchBuddyPanel.module.css";

interface MatchBuddyPanelProps {
  roomId: string | undefined;
}

export const MatchBuddyPanel: React.FC<MatchBuddyPanelProps> = ({ roomId }) => {
  const { user } = useAuth();
  const [isTeamModalOpen, setIsTeamModalOpen] = useState(false);

  const {
    matchState,
    feedCards,
    teamChanges,
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
  } = useMatchBuddy(roomId);

  const currentTeam = user?.id && matchState?.teams ? matchState.teams[user.id] : undefined;
  const usernames: Record<string, string> = {};
  if (matchState?.teams) {
    Object.values(matchState.teams).forEach((t) => {
      usernames[t.user_id] = t.username;
    });
  }

  return (
    <div className={styles.panel} role="region" aria-label="Match Buddy live fantasy companion">
      {/* 1. Scoreboard */}
      <MatchScoreboard state={matchState} />

      {/* 2. Replay Controls */}
      <MatchControls
        status={matchState?.status || "lobby"}
        speed={matchState?.speed_multiplier || 1.0}
        hasBot={matchState?.has_bot || false}
        teamsLocked={matchState?.teams_locked || false}
        onStart={startMatch}
        onPause={pauseMatch}
        onResume={resumeMatch}
        onSetSpeed={setSpeed}
        onSkipOver={skipOver}
        onReset={resetMatch}
        onToggleSolo={toggleSolo}
        onOpenTeamBuilder={() => setIsTeamModalOpen(true)}
      />

      {/* 3. Fantasy Leaderboard */}
      <Leaderboard
        entries={matchState?.leaderboard || []}
        recentDeltas={recentDeltas}
      />

      {/* 4. What-If Simulator */}
      <WhatIfChips
        onSelectScenario={whatIf}
        disabled={matchState?.status === "completed"}
      />

      {/* 5. Ask Match Buddy / AI Companion */}
      <AskBuddyInput
        onAsk={askBuddy}
        onCatchUp={catchUp}
        disabled={matchState?.status === "lobby" && !matchState?.teams_locked}
      />

      {/* 6. Pre-Match Team Changes Audit Log */}
      <TeamChangesLog changes={teamChanges} />

      {/* 7. Live Companion Feed */}
      <div className={styles.feedSection}>
        <div className={styles.feedHeader}>
          <span className={styles.feedTitle}>
            <Activity size={12} className="inline mr-1 text-[#e11d2a]" /> Companion Feed
          </span>
          <span className="text-xs text-[var(--color-text-muted)]">
            {feedCards.length} update{feedCards.length === 1 ? "" : "s"}
          </span>
        </div>

        <div className={styles.feedStream}>
          {feedCards.length === 0 ? (
            <div className={styles.emptyFeed}>
              Start the match to see live point swings, milestones, and strategic commentary here!
            </div>
          ) : (
            feedCards.map((card) => (
              <FeedCardItem
                key={card.id}
                card={card}
                usernames={usernames}
              />
            ))
          )}
        </div>
      </div>

      {/* Team Builder Modal */}
      {matchState && (
        <TeamBuilderModal
          isOpen={isTeamModalOpen}
          onClose={() => setIsTeamModalOpen(false)}
          squad={matchState.squad}
          presets={matchState.presets}
          currentTeam={currentTeam}
          isLocked={matchState.teams_locked}
          onSaveTeam={(pIds, cId, vcId, presetId) => {
            setTeam(pIds, cId, vcId, presetId);
          }}
        />
      )}
    </div>
  );
};
