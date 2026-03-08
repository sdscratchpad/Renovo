import React, { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  LuClock, LuTimer, LuCircleCheckBig, LuTrendingUp,
  LuServer, LuArrowRight, LuActivity,
} from '../icons';
import { getIncidents, getKPIs, getRemediationResult, getIncidentStatus } from '../api/client';
import { useWS } from '../context/WSContext';
import { useDrawer } from '../context/DrawerContext';
import { IncidentEvent, KPISummary, RemediationResult, IncidentStatusUpdate } from '../api/types';
import Spinner from '../components/Spinner';
import ErrorBanner from '../components/ErrorBanner';
import FaultDial from '../components/FaultDial';
import styles from './Dashboard.module.css';

function severityClass(severity: string): string {
  switch (severity) {
    case 'critical': return styles.severityCritical;
    case 'high':     return styles.severityHigh;
    case 'medium':   return styles.severityMedium;
    default:         return styles.severityLow;
  }
}

function scenarioLabel(scenario: string): string {
  switch (scenario) {
    case 'bad-rollout':          return 'Bad Rollout';
    case 'resource-saturation':  return 'Resource Saturation';
    case 'batch-timeout':        return 'Batch Timeout';
    default:                     return scenario;
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case 'detected':          return 'Detected';
    case 'analyzing':         return 'Analyzing…';
    case 'awaiting_approval': return 'Awaiting Approval';
    case 'remediating':       return 'Remediating…';
    case 'resolved':          return 'Resolved';
    case 'failed':            return 'Failed';
    default:                  return status;
  }
}

function statusClass(status: string, styles: Record<string, string>): string {
  switch (status) {
    case 'detected':          return styles.statusDetected;
    case 'analyzing':         return styles.statusAnalyzing;
    case 'awaiting_approval': return styles.statusAwaitingApproval;
    case 'remediating':       return styles.statusRemediating;
    case 'resolved':          return styles.statusResolved;
    case 'failed':            return styles.statusFailed;
    default:                  return styles.statusDetected;
  }
}

const MONITORED_SERVICES = ['sample-app', 'batch-worker'];
const PRODUCT_SERVICES   = ['fault-injector', 'diagnosis', 'orchestrator', 'event-store'];

const Dashboard: React.FC = () => {
  const [incidents, setIncidents] = useState<IncidentEvent[]>([]);
  const [kpis, setKpis] = useState<KPISummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [healthTab, setHealthTab] = useState<'monitored' | 'product'>('monitored');
  const [incidentTab, setIncidentTab] = useState<'active' | 'resolved' | 'failed'>('active');
  const [remediationResults, setRemediationResults] = useState<Map<string, RemediationResult>>(new Map());
  const [incidentStatuses, setIncidentStatuses] = useState<Map<string, IncidentStatusUpdate>>(new Map());
  const { openDrawer } = useDrawer();

  const fetchData = useCallback(async () => {
    try {
      const [inc, kpi] = await Promise.all([getIncidents(), getKPIs()]);
      setIncidents(inc);
      setKpis(kpi);
      const [results, statuses] = await Promise.all([
        Promise.all(inc.map(i => getRemediationResult(i.id))),
        Promise.all(inc.map(i => getIncidentStatus(i.id))),
      ]);
      const resultMap = new Map<string, RemediationResult>();
      inc.forEach((i, idx) => { if (results[idx]) resultMap.set(i.id, results[idx]!); });
      setRemediationResults(resultMap);
      const statusMap = new Map<string, IncidentStatusUpdate>();
      inc.forEach((i, idx) => { if (statuses[idx]) statusMap.set(i.id, statuses[idx]!); });
      setIncidentStatuses(statusMap);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load dashboard data.');
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial load + 30 s fallback poll (catches any events missed by WS).
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30_000);
    return () => clearInterval(interval);
  }, [fetchData]);

  // Immediate refresh driven by WebSocket push events from the event-store.
  const { lastEventAt } = useWS();
  useEffect(() => {
    if (lastEventAt === 0) return; // skip the initial value before first event
    fetchData();
  }, [lastEventAt, fetchData]);

  // Only incidents that are not yet resolved or failed affect the service health map.
  const affectedServices = new Set(
    incidents
      .filter(i => {
        const st = incidentStatuses.get(i.id)?.status;
        return st !== 'resolved' && st !== 'failed';
      })
      .map(i => i.service)
  );

  return (
    <>
    <div className={styles.page}>
      <div className={styles.pageHeader}>
        <h1 className={styles.heading}><LuActivity size={22} /> Dashboard</h1>
        <p className={styles.subheading}>Live view of infrastructure health, AI-detected incidents and SLO metrics.</p>
      </div>
      {error && <ErrorBanner message={error} />}
      {/* Health Map */}
      <section className={styles.section}>
        <div className={styles.sectionTitleRow}>
          <h2 className={styles.sectionTitle}>Service Health</h2>
          <div className={styles.tabBar}>
            <button
              className={`${styles.tab} ${healthTab === 'monitored' ? styles.tabActive : ''}`}
              onClick={() => setHealthTab('monitored')}
            >
              Monitored
            </button>
            <button
              className={`${styles.tab} ${healthTab === 'product' ? styles.tabActive : ''}`}
              onClick={() => setHealthTab('product')}
            >
              Product Services
            </button>
          </div>
        </div>
        <div className={styles.healthGrid}>
          {(healthTab === 'monitored' ? MONITORED_SERVICES : PRODUCT_SERVICES).map(svc => {
            const incident = incidents.find(i => {
              if (i.service !== svc) return false;
              const st = incidentStatuses.get(i.id)?.status;
              return st !== 'resolved' && st !== 'failed';
            });
            const isAffected = incident !== undefined;
            return (
              <div
                key={svc}
                className={`${styles.healthCard} ${isAffected ? severityClass(incident!.severity) : styles.healthOk}`}
              >
                <LuServer size={13} className={styles.serverIcon} />
                <span className={styles.healthDot} />
                <span className={styles.healthName}>{svc}</span>
                {isAffected && (
                  <span className={styles.healthBadge}>{incident!.severity.toUpperCase()}</span>
                )}
              </div>
            );
          })}
        </div>
      </section>

      {/* KPI Panel */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>KPIs <span className={styles.refreshNote}>(refreshes every 10s)</span></h2>
        {kpis ? (          <div className={styles.kpiGrid}>
            <div className={`${styles.kpiCard} ${styles.kpiMttd}`}>
              <LuClock size={16} className={styles.kpiIcon} />
              <div className={styles.kpiText}>
                <span className={styles.kpiValue}>{kpis.mttd_minutes.toFixed(1)} min</span>
                <span className={styles.kpiLabel}>MTTD</span>
              </div>
            </div>
            <div className={`${styles.kpiCard} ${styles.kpiMttr}`}>
              <LuTimer size={16} className={styles.kpiIcon} />
              <div className={styles.kpiText}>
                <span className={styles.kpiValue}>{kpis.mttr_minutes.toFixed(1)} min</span>
                <span className={styles.kpiLabel}>MTTR</span>
              </div>
            </div>
            <div className={`${styles.kpiCard} ${styles.kpiResolved}`}>
              <LuCircleCheckBig size={16} className={styles.kpiIcon} />
              <div className={styles.kpiText}>
                <span className={styles.kpiValue}>{kpis.resolved_today}</span>
                <span className={styles.kpiLabel}>Resolved Today</span>
              </div>
            </div>
            <div className={`${styles.kpiCard} ${styles.kpiAutoResolve}`}>
              <LuTrendingUp size={16} className={styles.kpiIcon} />
              <div className={styles.kpiText}>
                <span className={styles.kpiValue}>{kpis.auto_resolve_rate}%</span>
                <span className={styles.kpiLabel}>Auto-Resolve Rate</span>
              </div>
            </div>
          </div>
        ) : (
          <Spinner label="Loading KPIs…" />
        )}
      </section>

      {/* Incidents List */}
      <section className={styles.section}>
        {(() => {
          const activeIncidents   = incidents.filter(i => !remediationResults.has(i.id));
          const resolvedIncidents = incidents.filter(i => remediationResults.get(i.id)?.success === true);
          const failedIncidents   = incidents.filter(i => remediationResults.get(i.id)?.success === false);
          const tabIncidents = incidentTab === 'active' ? activeIncidents
            : incidentTab === 'resolved' ? resolvedIncidents
            : failedIncidents;
          const emptyMessages: Record<string, string> = {
            active:   'No active incidents.',
            resolved: 'No resolved incidents.',
            failed:   'No failed remediations.',
          };
          return (
            <>
              <div className={styles.sectionTitleRow}>
                <h2 className={styles.sectionTitle}>Incidents</h2>
                <div className={styles.tabBar}>
                  <button
                    className={`${styles.tab} ${incidentTab === 'active' ? styles.tabActive : ''}`}
                    onClick={() => setIncidentTab('active')}
                  >
                    Active ({activeIncidents.length})
                  </button>
                  <button
                    className={`${styles.tab} ${incidentTab === 'resolved' ? styles.tabActive : ''}`}
                    onClick={() => setIncidentTab('resolved')}
                  >
                    Resolved ({resolvedIncidents.length})
                  </button>
                  <button
                    className={`${styles.tab} ${incidentTab === 'failed' ? styles.tabActive : ''}`}
                    onClick={() => setIncidentTab('failed')}
                  >
                    Failed Remediation ({failedIncidents.length})
                  </button>
                </div>
              </div>
              {loading ? (
                <Spinner label="Loading incidents…" />
              ) : tabIncidents.length === 0 ? (
                <p className={styles.emptyText}>{emptyMessages[incidentTab]}</p>
              ) : (
                <table className={styles.table}>
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>Scenario</th>
                      <th>Service</th>
                      <th>Severity</th>
                      <th>Detected At</th>
                      <th>Status</th>
                      {incidentTab !== 'active' && <th>Result</th>}
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {tabIncidents.map(inc => {
                      const st = incidentStatuses.get(inc.id)?.status ?? 'detected';
                      return (
                      <tr key={inc.id}>
                        <td className={styles.monoCell}>{inc.id}</td>
                        <td>{scenarioLabel(inc.scenario)}</td>
                        <td>{inc.service}</td>
                        <td>
                          <span className={`${styles.severityBadge} ${severityClass(inc.severity)}`}>
                            {inc.severity}
                          </span>
                        </td>
                        <td>{new Date(inc.detected_at).toLocaleString()}</td>
                        <td>
                            <span className={`${styles.statusBadge} ${statusClass(st, styles)}`}>
                              {statusLabel(st)}
                            </span>
                          </td>
                        {incidentTab !== 'active' && (
                          <td className={styles.resultCell}>{remediationResults.get(inc.id)?.message ?? '—'}</td>
                        )}
                        <td>
                          <button
                            className={styles.detailLink}
                            onClick={() => openDrawer(inc.id)}
                          >
                            View <LuArrowRight size={12} />
                          </button>
                        </td>
                      </tr>
                      );
                    })}
                  </tbody>
                </table>
              )}
            </>
          );
        })()}
      </section>
    </div>
    <FaultDial />
    </>
  );
};

export default Dashboard;
