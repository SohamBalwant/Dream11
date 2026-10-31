import React, { useState } from "react";
import { MessageSquareText, Send, Zap } from "lucide-react";
import styles from "./AskBuddyInput.module.css";

interface AskBuddyInputProps {
  onAsk: (question: string) => void;
  onCatchUp: () => void;
  disabled?: boolean;
}

const QUICK_PROMPTS = [
  "Who is winning right now?",
  "How is the match going?",
  "Who is the top fantasy performer?",
];

export const AskBuddyInput: React.FC<AskBuddyInputProps> = ({
  onAsk,
  onCatchUp,
  disabled = false,
}) => {
  const [question, setQuestion] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!question.trim() || disabled) return;
    onAsk(question.trim());
    setQuestion("");
  };

  const handlePromptClick = (p: string) => {
    if (disabled) return;
    onAsk(p);
  };

  return (
    <div className={styles.container}>
      <div className={styles.headerRow}>
        <span className={styles.title}>
          <MessageSquareText size={13} className="inline mr-1" /> Ask Match Buddy
        </span>
        <button
          type="button"
          className={styles.catchUpBtn}
          onClick={onCatchUp}
          disabled={disabled}
        >
          <Zap size={10} /> Catch me up
        </button>
      </div>

      <form onSubmit={handleSubmit} className={styles.inputForm}>
        <input
          type="text"
          className={styles.inputField}
          placeholder="Ask anything about points, leaders or players..."
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          disabled={disabled}
        />
        <button
          type="submit"
          className={styles.sendBtn}
          disabled={!question.trim() || disabled}
          aria-label="Send question"
        >
          <Send size={15} />
        </button>
      </form>

      <div className={styles.quickPrompts}>
        {QUICK_PROMPTS.map((p) => (
          <button
            key={p}
            type="button"
            className={styles.promptChip}
            onClick={() => handlePromptClick(p)}
            disabled={disabled}
          >
            {p}
          </button>
        ))}
      </div>
    </div>
  );
};
