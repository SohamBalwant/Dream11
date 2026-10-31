import React, { useState } from "react";
import type { Player, PresetDefinition, FantasyTeam } from "../../types/matchbuddy";
import { X, Sparkles, CheckCircle2 } from "lucide-react";
import styles from "./TeamBuilderModal.module.css";

interface TeamBuilderModalProps {
  isOpen: boolean;
  onClose: () => void;
  squad: Player[];
  presets: PresetDefinition[];
  currentTeam?: FantasyTeam;
  isLocked?: boolean;
  onSaveTeam: (
    playerIds: string[],
    captainId: string,
    viceCaptainId: string,
    presetId?: string,
  ) => void;
}

export const TeamBuilderModal: React.FC<TeamBuilderModalProps> = ({
  isOpen,
  onClose,
  squad,
  presets,
  currentTeam,
  isLocked = false,
  onSaveTeam,
}) => {
  const [selectedPreset, setSelectedPreset] = useState<string>(
    currentTeam?.preset_id || "",
  );
  const [selectedIds, setSelectedIds] = useState<string[]>(
    currentTeam?.player_ids || presets[0]?.player_ids || [],
  );
  const [captainId, setCaptainId] = useState<string>(
    currentTeam?.captain_id || presets[0]?.captain_id || "",
  );
  const [viceCaptainId, setViceCaptainId] = useState<string>(
    currentTeam?.vice_captain_id || presets[0]?.vice_captain_id || "",
  );

  if (!isOpen) return null;

  const handleApplyPreset = (p: PresetDefinition) => {
    setSelectedPreset(p.id);
    setSelectedIds([...p.player_ids]);
    setCaptainId(p.captain_id);
    setViceCaptainId(p.vice_captain_id);
  };

  const togglePlayer = (id: string) => {
    if (isLocked) return;
    setSelectedPreset(""); // custom edit
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter((p) => p !== id));
      if (captainId === id) setCaptainId("");
      if (viceCaptainId === id) setViceCaptainId("");
    } else {
      if (selectedIds.length < 11) {
        setSelectedIds([...selectedIds, id]);
      }
    }
  };

  const handleSetCaptain = (id: string) => {
    if (isLocked) return;
    setSelectedPreset("");
    if (viceCaptainId === id) setViceCaptainId("");
    setCaptainId(id);
  };

  const handleSetViceCaptain = (id: string) => {
    if (isLocked) return;
    setSelectedPreset("");
    if (captainId === id) setCaptainId("");
    setViceCaptainId(id);
  };

  // Validation
  const mbCount = squad.filter(
    (p) => selectedIds.includes(p.id) && p.team === "MB",
  ).length;
  const dtCount = squad.filter(
    (p) => selectedIds.includes(p.id) && p.team === "DT",
  ).length;

  const isValidCount = selectedIds.length === 11;
  const isValidSplit = mbCount <= 7 && dtCount <= 7;
  const hasCaptain = captainId !== "" && selectedIds.includes(captainId);
  const hasVC =
    viceCaptainId !== "" &&
    selectedIds.includes(viceCaptainId) &&
    viceCaptainId !== captainId;
  const canSave = isValidCount && isValidSplit && hasCaptain && hasVC && !isLocked;

  const handleSave = () => {
    if (!canSave) return;
    onSaveTeam(selectedIds, captainId, viceCaptainId, selectedPreset || undefined);
    onClose();
  };

  return (
    <div className={styles.overlay} onClick={onClose} role="dialog" aria-modal="true">
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <div className={styles.titleCluster}>
            <h2 className={styles.title}>Pick Your Dream11</h2>
            <p className={styles.subtitle}>
              Select 11 players (max 7 per side), 1 Captain (2x), 1 Vice-Captain (1.5x)
            </p>
          </div>
          <button className={styles.closeBtn} onClick={onClose} aria-label="Close modal">
            <X size={20} />
          </button>
        </div>

        <div className={styles.content}>
          {/* Quick Presets */}
          <div className={styles.presetsSection}>
            <span className={styles.sectionTitle}>
              <Sparkles size={12} className="inline mr-1" /> Quick Presets (1-Click)
            </span>
            <div className={styles.presetsGrid}>
              {presets.map((p) => {
                const isSelected = selectedPreset === p.id;
                return (
                  <button
                    key={p.id}
                    type="button"
                    className={`${styles.presetCard} ${isSelected ? styles.presetSelected : ""}`}
                    onClick={() => handleApplyPreset(p)}
                    disabled={isLocked}
                  >
                    <div className="flex items-center justify-between">
                      <span className={styles.presetName}>{p.name}</span>
                      {isSelected && <CheckCircle2 size={14} className="text-[#e11d2a]" />}
                    </div>
                    <span className={styles.presetDesc}>{p.description}</span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Status Bar */}
          <div className={styles.statusBar}>
            <div className={styles.counters}>
              <span>
                Selected:{" "}
                <span className={isValidCount ? styles.countValid : styles.countPending}>
                  {selectedIds.length}/11
                </span>
              </span>
              <span>·</span>
              <span>
                MB: <span className={mbCount <= 7 ? "" : styles.countPending}>{mbCount}/7</span>
              </span>
              <span>
                DT: <span className={dtCount <= 7 ? "" : styles.countPending}>{dtCount}/7</span>
              </span>
            </div>

            <div className="flex items-center gap-2 text-xs">
              <span className={hasCaptain ? "text-emerald-400 font-semibold" : "text-zinc-500"}>
                {hasCaptain ? "✓ C chosen" : "Pick C (2x)"}
              </span>
              <span>·</span>
              <span className={hasVC ? "text-emerald-400 font-semibold" : "text-zinc-500"}>
                {hasVC ? "✓ VC chosen" : "Pick VC (1.5x)"}
              </span>
            </div>
          </div>

          {/* Squad List */}
          <div className={styles.playerList}>
            {squad.map((player) => {
              const isSelected = selectedIds.includes(player.id);
              const isC = captainId === player.id;
              const isVC = viceCaptainId === player.id;

              return (
                <div
                  key={player.id}
                  className={`${styles.playerRow} ${isSelected ? styles.playerRowSelected : ""}`}
                >
                  <div className={styles.playerInfo}>
                    <input
                      type="checkbox"
                      checked={isSelected}
                      disabled={isLocked || (!isSelected && selectedIds.length >= 11)}
                      onChange={() => togglePlayer(player.id)}
                      className="accent-[#e11d2a] w-4 h-4 cursor-pointer"
                    />
                    <span
                      className={`${styles.teamBadge} ${
                        player.team === "MB" ? styles.teamMB : styles.teamDT
                      }`}
                    >
                      {player.team}
                    </span>
                    <div className={styles.playerText}>
                      <span className={styles.pName}>{player.name}</span>
                      <span className={styles.pMeta}>
                        {player.role} · {player.credits} Cr
                      </span>
                    </div>
                  </div>

                  <div className={styles.actionCluster}>
                    <button
                      type="button"
                      title="Set Captain (2x points)"
                      className={`${styles.roleBtn} ${isC ? styles.roleBtnActiveC : ""}`}
                      onClick={() => handleSetCaptain(player.id)}
                      disabled={!isSelected || isLocked}
                    >
                      C
                    </button>
                    <button
                      type="button"
                      title="Set Vice-Captain (1.5x points)"
                      className={`${styles.roleBtn} ${isVC ? styles.roleBtnActiveVC : ""}`}
                      onClick={() => handleSetViceCaptain(player.id)}
                      disabled={!isSelected || isLocked}
                    >
                      VC
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        <div className={styles.footer}>
          <span className="text-xs text-[var(--color-text-muted)]">
            {isLocked ? "Teams locked — match in progress" : "Changes sync to all room members"}
          </span>
          <button
            type="button"
            className={styles.saveBtn}
            onClick={handleSave}
            disabled={!canSave}
          >
            Confirm 11 Lineup
          </button>
        </div>
      </div>
    </div>
  );
};
