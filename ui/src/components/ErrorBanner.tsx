import React, { useState } from 'react';
import styles from './ErrorBanner.module.css';

interface ErrorBannerProps {
  message: string;
}

const ErrorBanner: React.FC<ErrorBannerProps> = ({ message }) => {
  const [dismissed, setDismissed] = useState(false);
  if (dismissed) return null;
  return (
    <div className={styles.banner} role="alert">
      <span className={styles.icon}>⚠</span>
      <span className={styles.message}>{message}</span>
      <button
        className={styles.dismiss}
        onClick={() => setDismissed(true)}
        aria-label="Dismiss error"
      >
        ✕
      </button>
    </div>
  );
};

export default ErrorBanner;
