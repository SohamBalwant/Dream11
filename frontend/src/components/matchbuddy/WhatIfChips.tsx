import React from "react";
import styles from "./WhatIfChips.module.css";

interface WhatIfChipsProps {
  onSelectScenario: (scenario: "six" | "four" | "wicket" | "dot") => void;
  disabled?: boolean;
}

export const WhatIfChips: React.FC<WhatIfChipsProps> = ({
  onSelectScenario,
  disabled = false,
}) => {
  return (
    <div className={styles.container}>
      <div className={styles.titleRow}>
        <span className={styles.title}>What-If Simulator</span>
        <span className={styles.subtitle}>Instant projection</span>
      </div>

      <div className={styles.chipsGrid}>
        <button
          type="button"
          className={styles.chipBtn}
          onClick={() => onSelectScenario("six")}
          disabled={disabled}
        >
          <span className={styles.chipIcon}>🚀</span>
          <span className={styles.chipLabel}>What if a 6?</span>
        </button>

        <button
          type="button"
          className={styles.chipBtn}
          onClick={() => onSelectScenario("four")}
          disabled={disabled}
        >
          <span className={styles.chipIcon}>⚡</span>
          <span className={styles.chipLabel}>What if a 4?</span>
        </button>

        <button
          type="button"
          className={styles.chipBtn}
          onClick={() => onSelectScenario("wicket")}
          disabled={disabled}
        >
          <span className={styles.chipIcon}>💥</span>
          <span className={styles.chipLabel}>What if Wicket?</span>
        </button>

        <button
          type="button"
          className={styles.chipBtn}
          onClick={() => onSelectScenario("dot")}
          disabled={disabled}
        >
          <span className={styles.chipIcon}>🛡️</span>
          <span className={styles.chipLabel}>What if Dot?</span>
        </button>
      </div>
    </div>
  );
};
