/**
 * FaultDial — floating speed-dial button fixed to the bottom-right of the viewport.
 *
 * Click the main ⚡ button to expand three scenario arms. Click a scenario to inject
 * the fault. After a successful injection the dial shows a brief "✓ Injected" toast
 * on the trigger arm, then collapses back to idle automatically after ~1.2 s.
 */
import React, { useEffect, useRef, useState } from 'react';
import { LuZap, LuCheck, LuLoader } from '../icons';
import { injectFault } from '../api/client';
import { useDrawer } from '../context/DrawerContext';
import styles from './FaultDial.module.css';

interface Scenario {
  id: string;
  label: string;
  desc: string;
  emoji: string;
  severity: 'critical' | 'high' | 'medium';
}

const SCENARIOS: Scenario[] = [
  { id: 'bad-rollout',         label: 'Bad Rollout',         desc: 'Deploy a broken image — AI detects error spike and rolls back',       emoji: '🔴', severity: 'critical' },
  { id: 'resource-saturation', label: 'Resource Saturation', desc: 'Inject CPU stress — AI detects saturation and scales up replicas',     emoji: '🟠', severity: 'high'     },
  { id: 'batch-timeout',       label: 'Batch Timeout',       desc: 'Block job dependency — AI detects stall and triggers a retry',        emoji: '🟡', severity: 'medium'   },
];

type ItemState = 'idle' | 'loading' | 'success' | 'error';

const FaultDial: React.FC = () => {
  const [open, setOpen] = useState(false);
  const [states, setStates] = useState<Record<string, ItemState>>(
    Object.fromEntries(SCENARIOS.map(s => [s.id, 'idle']))
  );
  const containerRef = useRef<HTMLDivElement>(null);
  const { drawerIncidentId } = useDrawer();
  const drawerOpen = drawerIncidentId !== null;

  // Close on outside click
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  const handleInject = async (scenario: Scenario) => {
    if (states[scenario.id] !== 'idle') return;
    setStates(prev => ({ ...prev, [scenario.id]: 'loading' }));
    try {
      await injectFault(scenario.id);
      setStates(prev => ({ ...prev, [scenario.id]: 'success' }));
      // Show success briefly then collapse
      setTimeout(() => {
        setStates(prev => ({ ...prev, [scenario.id]: 'idle' }));
        setOpen(false);
      }, 1200);
    } catch {
      setStates(prev => ({ ...prev, [scenario.id]: 'error' }));
      // Keep error visible for 2.5 s then reset
      setTimeout(() => {
        setStates(prev => ({ ...prev, [scenario.id]: 'idle' }));
      }, 2500);
    }
  };

  return (
    <div className={styles.dial} ref={containerRef} style={{ right: drawerOpen ? 'calc(520px + 32px)' : '32px' }}>
      {/* Scenario arms — shown when open */}
      {SCENARIOS.map((scenario, i) => {
        const state = states[scenario.id];
        const isLoading = state === 'loading';
        const isSuccess = state === 'success';
        const isError   = state === 'error';
        return (
          <div
            key={scenario.id}
            className={`${styles.arm} ${open ? styles.armOpen : ''} ${styles[`arm${i}`]}`}
            style={{ transitionDelay: open ? `${i * 55}ms` : `${(SCENARIOS.length - 1 - i) * 40}ms` }}
          >
            {/* Label chip */}
            <span className={`${styles.armLabel} ${isSuccess ? styles.armLabelSuccess : ''} ${isError ? styles.armLabelError : ''}`}>
              {isSuccess ? (
                <><span className={styles.armLabelName}>✓ Injected</span><span className={styles.armLabelDesc}>Incident detection may take a few seconds</span></>
              ) : isError ? (
                <><span className={styles.armLabelName}>Failed</span><span className={styles.armLabelDesc}>Is the fault-injector service running?</span></>
              ) : (
                <><span className={styles.armLabelName}>{scenario.label}</span><span className={styles.armLabelDesc}>{scenario.desc}</span></>
              )}
            </span>
            {/* Round button */}
            <button
              className={`${styles.armBtn} ${styles['sev_' + scenario.severity]} ${isSuccess ? styles.armBtnSuccess : ''} ${isError ? styles.armBtnError : ''}`}
              onClick={() => handleInject(scenario)}
              disabled={isLoading}
              title={scenario.label}
            >
              {isLoading
                ? <LuLoader size={14} className={styles.spinIcon} />
                : isSuccess
                  ? <LuCheck size={14} />
                  : <span className={styles.emoji}>{scenario.emoji}</span>
              }
            </button>
          </div>
        );
      })}

      {/* Main FAB */}
      <div className={styles.fabRow}>
        <span className={`${styles.fabLabel} ${open ? styles.fabLabelHidden : ''}`}>
          Inject Fault
        </span>
        <button
          className={`${styles.fab} ${open ? styles.fabOpen : ''}`}
          onClick={() => setOpen(v => !v)}
          title="Fault Injector"
          aria-label="Open fault injector"
        >
          <LuZap size={20} className={`${styles.fabIcon} ${open ? styles.fabIconOpen : ''}`} />
        </button>
      </div>
    </div>
  );
};

export default FaultDial;
