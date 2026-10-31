import React, { type ReactNode } from "react";
import { Logo } from "../common/Logo";
import { KeyRound } from "lucide-react";
import styles from "./AuthLayout.module.css";

interface AuthLayoutProps {
  title: string;
  subtitle?: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
  onQuickFill?: () => void;
}

/** Shared frame for sign in, sign up and password recovery. */
export const AuthLayout: React.FC<AuthLayoutProps> = ({
  title,
  subtitle,
  children,
  footer,
  onQuickFill,
}) => (
  <main className={styles.page}>
    <aside className={styles.demoCredentialsBox} aria-label="Demo login credentials">
      <div className={styles.demoCredHeader}>
        <KeyRound size={13} className={styles.demoCredIcon} />
        <span>Use credentials to login</span>
      </div>
      <div className={styles.demoCredRow}>
        <span className={styles.demoCredLabel}>Email:</span>
        <code className={styles.demoCredValue}>cricketer@dream11.com</code>
      </div>
      <div className={styles.demoCredRow}>
        <span className={styles.demoCredLabel}>Password:</span>
        <code className={styles.demoCredValue}>Password@123</code>
      </div>
      {onQuickFill && (
        <button
          type="button"
          className={styles.quickFillBtn}
          onClick={onQuickFill}
        >
          Click to auto-fill
        </button>
      )}
    </aside>

    <div className={styles.brand}>
      <Logo size={36} withWordmark />
    </div>
    <div className={styles.card}>
      <header className={styles.header}>
        <h1 className={styles.title}>{title}</h1>
        {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
      </header>
      {children}
      {footer && <div className={styles.footer}>{footer}</div>}
    </div>
  </main>
);
