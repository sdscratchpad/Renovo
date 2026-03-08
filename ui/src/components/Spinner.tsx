import React from 'react';
import styles from './Spinner.module.css';

interface SpinnerProps {
  label?: string;
}

const Spinner: React.FC<SpinnerProps> = ({ label = 'Loading…' }) => (
  <div className={styles.wrapper}>
    <span className={styles.spinner} aria-hidden="true" />
    <span>{label}</span>
  </div>
);

export default Spinner;
