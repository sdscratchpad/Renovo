import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  LuCircleAlert, LuBrainCircuit, LuCircleCheck, LuPlay,
  LuActivity, LuUser, LuBot, LuClock,
} from '../icons';
import { getAuditEntries } from '../api/client';
import { AuditEntry } from '../api/types';
import Spinner from '../components/Spinner';
import ErrorBanner from '../components/ErrorBanner';
import styles from './AuditLog.module.css';

function eventDotClass(event: string): string {
  switch (event) {
    case 'incident_detected': return styles.dotDetected;
    case 'rca_generated':     return styles.dotRCA;
    case 'action_approved':   return styles.dotApproved;
    case 'action_executed':   return styles.dotExecuted;
    default:                  return styles.dotDefault;
  }
}

function eventBadgeClass(event: string): string {
  switch (event) {
    case 'incident_detected': return styles.eventDetected;
    case 'rca_generated':     return styles.eventRCA;
    case 'action_approved':   return styles.eventApproved;
    case 'action_executed':   return styles.eventExecuted;
    default:                  return styles.eventDefault;
  }
}

function eventLabel(event: string): string {
  switch (event) {
    case 'incident_detected': return 'Detected';
    case 'rca_generated':     return 'RCA Generated';
    case 'action_approved':   return 'Approved';
    case 'action_executed':   return 'Executed';
    default:                  return event;
  }
}

function EventDotIcon({ event }: { event: string }) {
  switch (event) {
    case 'incident_detected': return <LuCircleAlert size={13} />;
    case 'rca_generated':     return <LuBrainCircuit size={13} />;
    case 'action_approved':   return <LuCircleCheck size={13} />;
    case 'action_executed':   return <LuPlay size={13} />;
    default:                  return <LuActivity size={13} />;
  }
}

const AuditLog: React.FC = () => {
  const [entries, setEntries] = useState<AuditEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getAuditEntries()
      .then(data => {
        setEntries([...data].sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()));
      })
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load audit log.'))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className={styles.page}>
      <div className={styles.pageHeader}>
        <h1 className={styles.heading}><LuActivity size={22} /> Audit Log</h1>
        <p className={styles.intro}>Chronological trail of all system and operator actions across every incident.</p>
      </div>
      {error && <ErrorBanner message={error} />}
      {loading ? (
        <Spinner label="Loading audit entries…" />
      ) : entries.length === 0 ? (
        <p className={styles.emptyText}>No audit entries recorded yet.</p>
      ) : (
        <div className={styles.timeline}>
          {entries.map((entry, idx) => (
            <div
              key={entry.id}
              className={styles.timelineItem}
              style={{ animationDelay: `${idx * 40}ms` }}
            >
              <div className={`${styles.timelineDot} ${eventDotClass(entry.event)}`}>
                <EventDotIcon event={entry.event} />
              </div>
              <div className={styles.timelineContent}>
                <div className={styles.timelineHeader}>
                  <span className={`${styles.eventBadge} ${eventBadgeClass(entry.event)}`}>
                    {eventLabel(entry.event)}
                  </span>
                  <span className={`${styles.actorBadge} ${entry.actor === 'operator' ? styles.actorOperator : styles.actorSystem}`}>
                    {entry.actor === 'operator' ? <LuUser size={11} /> : <LuBot size={11} />}
                    {entry.actor}
                  </span>
                  <span className={styles.tsCell}>
                    <LuClock size={11} />
                    {new Date(entry.timestamp).toLocaleString()}
                  </span>
                </div>
                <Link to={`/incidents/${entry.incident_id}`} className={styles.incidentLink}>
                  {entry.incident_id}
                </Link>
                <p className={styles.detailCell}>{entry.detail}</p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default AuditLog;
