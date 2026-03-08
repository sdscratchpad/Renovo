import React, { useState } from 'react';
import { LuZap, LuCheck, LuRotateCcw, LuLoader } from '../icons';
import { injectFault } from '../api/client';
import styles from './FaultInjector.module.css';

interface Scenario {
  id: string;
  label: string;
  description: string;
  icon: string;
  severity: 'critical' | 'high' | 'medium';
}

const SCENARIOS: Scenario[] = [
  {
    id: 'bad-rollout',
    label: 'Bad Rollout',
    description:
      'Deploys a misconfigured version of sample-app with an invalid environment variable, causing CrashLoopBackOff and an HTTP error rate spike.',
    icon: '🔴',
    severity: 'critical',
  },
  {
    id: 'resource-saturation',
    label: 'Resource Saturation',
    description:
      'Injects CPU stress into sample-app pods, driving CPU usage above 95% and triggering a latency degradation alert.',
    icon: '🟠',
    severity: 'high',
  },
  {
    id: 'batch-timeout',
    label: 'Batch Timeout',
    description:
      'Blocks the batch-worker dependency so the periodic job stalls and fails to complete within its SLA window.',
    icon: '🟡',
    severity: 'medium',
  },
];

type InjectionState = 'idle' | 'loading' | 'success' | 'error';

const FaultInjector: React.FC = () => {
  const [states, setStates] = useState<Record<string, InjectionState>>(
    Object.fromEntries(SCENARIOS.map(s => [s.id, 'idle']))
  );
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleInject = async (scenarioId: string) => {
    setStates(prev => ({ ...prev, [scenarioId]: 'loading' }));
    setErrors(prev => { const next = { ...prev }; delete next[scenarioId]; return next; });
    try {
      await injectFault(scenarioId);
      setStates(prev => ({ ...prev, [scenarioId]: 'success' }));
      // Reset back to idle after 3s
      setTimeout(() => {
        setStates(prev => ({ ...prev, [scenarioId]: 'idle' }));
      }, 3000);
    } catch {
      setStates(prev => ({ ...prev, [scenarioId]: 'error' }));
      setErrors(prev => ({ ...prev, [scenarioId]: 'Injection failed. Is the fault-injector running?' }));
    }
  };

  return (
    <div className={styles.page}>
      <h1 className={styles.heading}>Fault Injector</h1>
      <p className={styles.intro}>
        Trigger one of the 3 MVP fault scenarios on the local kind cluster. Each scenario fires the
        corresponding fault-injector API endpoint and starts the AI incident response pipeline.
      </p>

      <div className={styles.scenarioGrid}>
        {SCENARIOS.map(scenario => {
          const state = states[scenario.id];
          return (
            <div key={scenario.id} className={`${styles.card} ${styles['sev_' + scenario.severity]}`}>
              <div className={styles.cardHeader}>
                <span className={styles.cardIcon}>{scenario.icon}</span>
                <h2 className={styles.cardTitle}>{scenario.label}</h2>
                <span className={`${styles.severityBadge} ${styles['badge_' + scenario.severity]}`}>
                  {scenario.severity}
                </span>
              </div>
              <p className={styles.cardDesc}>{scenario.description}</p>
              <div className={styles.cardFooter}>
                <button
                  className={`${styles.injectBtn} ${state === 'success' ? styles.injectBtnSuccess : ''}`}
                  onClick={() => handleInject(scenario.id)}
                  disabled={state === 'loading'}
                >
                  {state === 'loading' && <><LuLoader size={14} className={styles.spinIcon} /> Injecting…</>}
                  {state === 'success' && <><LuCheck size={14} /> Injected</>}
                  {state === 'error' && <><LuRotateCcw size={14} /> Retry</>}
                  {state === 'idle' && <><LuZap size={14} /> Inject Fault</>}
                </button>
                {state === 'success' && (
                  <span className={styles.successNote}>Fault triggered — check Dashboard for the incident.</span>
                )}
                {errors[scenario.id] && (
                  <span className={styles.errorNote}>{errors[scenario.id]}</span>
                )}
              </div>
            </div>
          );
        })}
      </div>

      <div className={styles.warningBox}>
        <span className={styles.warningIcon}>⚠️</span>
        <p>
          These scenarios modify the <strong>ravi-poc</strong> kind cluster. They are designed to be
          self-healing once the orchestrator approves the remediation action.
        </p>
      </div>
    </div>
  );
};

export default FaultInjector;
