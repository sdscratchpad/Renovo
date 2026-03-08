import React from 'react';
import RenovoLogo from './RenovoLogo';
import { useWS, WSStatus } from '../context/WSContext';
import styles from './NavBar.module.css';

const WS_LABELS: Record<WSStatus, string> = {
  connecting:   'Connecting…',
  connected:    'Live',
  disconnected: 'Disconnected',
};

const NavBar: React.FC = () => {
  const { status } = useWS();

  return (
    <nav className={styles.nav}>
      <div className={styles.brand}>
        <RenovoLogo size={34} />
        <div className={styles.brandText}>
          <span className={styles.brandName}>Renovo</span>
          <span className={styles.brandSub}>AI Infrastructure Resilience</span>
        </div>
      </div>
      <div className={`${styles.wsIndicator} ${styles['ws_' + status]}`}>
        <span className={styles.wsDot} />
        <span className={styles.wsLabel}>{WS_LABELS[status]}</span>
      </div>
    </nav>
  );
};

export default NavBar;
