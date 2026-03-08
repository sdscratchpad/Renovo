import React, { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  LuArrowRight, LuBrainCircuit, LuCircleCheck, LuLoader, LuShieldAlert, LuX,
} from '../icons';
import {
  getIncidents, getRCA, getRemediation, approveRemediation,
  triggerDiagnosis, getRemediationResult, getIncidentStatus,
} from '../api/client';
import { IncidentEvent, RCAPayload, RemediationRequest, RemediationResult } from '../api/types';
import Spinner from './Spinner';
import ErrorBanner from './ErrorBanner';
import StatusBadge, { deriveStatus } from './StatusBadge';
import styles from './IncidentDrawer.module.css';

function riskClass(risk: string): string {
  switch (risk) {
    case 'high':   return styles.riskHigh;
    case 'medium': return styles.riskMedium;
    default:       return styles.riskLow;
  }
}

interface Props {
  incidentId: string | null;
  onClose: () => void;
}

const IncidentDrawer: React.FC<Props> = ({ incidentId, onClose }) => {
  const [incident, setIncident]         = useState<IncidentEvent | null>(null);
  const [rca, setRca]                   = useState<RCAPayload | null>(null);
  const [remediation, setRemediation]   = useState<RemediationRequest | null>(null);
  const [executionResult, setExecResult] = useState<RemediationResult | null>(null);
  const [pipelineStatus, setPipelineStatus] = useState<string | null>(null);
  const [loading, setLoading]           = useState(false);
  const [loadError, setLoadError]       = useState<string | null>(null);
  const [diagnosing, setDiagnosing]     = useState(false);
  const [diagnosisError, setDiagnosisError] = useState<string | null>(null);
  const [rcaTimedOut, setRcaTimedOut]   = useState(false);
  const [approving, setApproving]       = useState(false);
  const [approveError, setApproveError] = useState<string | null>(null);

  const panelRef = useRef<HTMLDivElement>(null);

  // Close on Escape key
  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [onClose]);

  // Load data when incidentId changes
  useEffect(() => {
    if (!incidentId) {
      setIncident(null);
      setRca(null);
      setRemediation(null);
      setExecResult(null);
      setPipelineStatus(null);
      setLoadError(null);
      return;
    }
    setLoading(true);
    setLoadError(null);
    setRcaTimedOut(false);
    setDiagnosisError(null);
    Promise.all([
      getIncidents(),
      getRCA(incidentId),
      getRemediation(incidentId),
      getRemediationResult(incidentId),
      getIncidentStatus(incidentId),
    ])
      .then(([incidents, r, rem, execResult, statusUpdate]) => {
        setIncident(incidents.find(i => i.id === incidentId) ?? null);
        setRca(r);
        setRemediation(rem);
        if (execResult) setExecResult(execResult);
        if (statusUpdate) setPipelineStatus(statusUpdate.status);
      })
      .catch(e => setLoadError(e instanceof Error ? e.message : 'Failed to load incident.'))
      .finally(() => setLoading(false));
  }, [incidentId]);

  // Poll for RCA if none yet
  useEffect(() => {
    if (!incidentId || rca || loading || rcaTimedOut) return;
    let attempts = 0;
    const MAX = 20;
    const timer = setInterval(async () => {
      attempts++;
      const r = await getRCA(incidentId).catch(() => null);
      if (r) {
        setRca(r);
        setDiagnosing(false);
        clearInterval(timer);
        getRemediation(incidentId).then(rem => { if (rem) setRemediation(rem); }).catch(() => {});
        return;
      }
      if (attempts >= MAX) {
        setRcaTimedOut(true);
        setDiagnosing(false);
        clearInterval(timer);
      }
    }, 3000);
    return () => clearInterval(timer);
  }, [incidentId, rca, loading, rcaTimedOut]);

  // Poll for execution result
  useEffect(() => {
    if (!incidentId || executionResult || remediation?.approval !== 'approved' || loading) return;
    const timer = setInterval(async () => {
      const [result, statusUpdate] = await Promise.all([
        getRemediationResult(incidentId).catch(() => null),
        getIncidentStatus(incidentId).catch(() => null),
      ]);
      if (statusUpdate) setPipelineStatus(statusUpdate.status);
      if (result) {
        setExecResult(result);
        clearInterval(timer);
      }
    }, 2000);
    return () => clearInterval(timer);
  }, [incidentId, executionResult, remediation?.approval, loading]);

  const handleDiagnose = async () => {
    if (!incident) return;
    setDiagnosing(true);
    setDiagnosisError(null);
    setRcaTimedOut(false);
    try {
      await triggerDiagnosis(incident);
    } catch {
      setDiagnosisError('Failed to trigger diagnosis. Is the diagnosis service running?');
      setDiagnosing(false);
      setRcaTimedOut(true);
    }
  };

  const handleApprove = async () => {
    if (!remediation) return;
    setApproving(true);
    setApproveError(null);
    try {
      await approveRemediation(remediation.action.id);
      setRemediation(prev => prev ? { ...prev, approval: 'approved', updated_at: new Date().toISOString() } : prev);
    } catch {
      setApproveError('Failed to approve. Please try again.');
    } finally {
      setApproving(false);
    }
  };

  const isOpen = incidentId !== null;

  return (
    <div
      ref={panelRef}
      role="dialog"
      aria-modal="true"
      aria-label="Incident Details"
      className={`${styles.panel} ${isOpen ? styles.panelOpen : ''}`}
    >
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.headerLeft}>
            <LuShieldAlert size={15} className={styles.headerIcon} />
            <span className={styles.headerTitle}>
              {incident ? (
                <>
                  Incident <span className={styles.headerIdSpan}>{incident.id}</span>
                </>
              ) : 'Incident Details'}
            </span>
            {incident && (
              <StatusBadge status={deriveStatus(rca, remediation, executionResult, pipelineStatus)} />
            )}
          </div>
          <div className={styles.headerActions}>
            {incident && (
              <Link
                to={`/incidents/${incident.id}`}
                className={styles.fullPageLink}
                onClick={onClose}
                title="Open full page"
              >
                Full page <LuArrowRight size={12} />
              </Link>
            )}
            <button className={styles.closeBtn} onClick={onClose} aria-label="Close panel">
              <LuX size={16} />
            </button>
          </div>
        </div>

        {/* Body */}
        <div className={styles.body}>
          {loading && <Spinner label="Loading incident…" />}
          {loadError && <ErrorBanner message={loadError} />}
          {!loading && !loadError && incident && (
            <>
              {/* Overview */}
              <section className={styles.sectionCard}>
                <h2 className={styles.sectionTitle}>Overview</h2>
                <dl className={styles.metaGrid}>
                  <div className={styles.metaItem}><dt>Service</dt><dd>{incident.service}</dd></div>
                  <div className={styles.metaItem}><dt>Namespace</dt><dd>{incident.namespace}</dd></div>
                  <div className={styles.metaItem}><dt>Scenario</dt><dd>{incident.scenario}</dd></div>
                  <div className={styles.metaItem}>
                    <dt>Severity</dt>
                    <dd>
                      <span className={`${styles.severityBadge} ${styles['severity_' + incident.severity]}`}>
                        {incident.severity}
                      </span>
                    </dd>
                  </div>
                  <div className={styles.metaItem}>
                    <dt>Detected At</dt>
                    <dd>{new Date(incident.detected_at).toLocaleString()}</dd>
                  </div>
                </dl>
              </section>

              {/* Evidence */}
              <section className={styles.sectionCard}>
                <h2 className={styles.sectionTitle}>Evidence ({incident.evidence.length})</h2>
                <div className={styles.evidenceList}>
                  {incident.evidence.map((ev, i) => (
                    <div key={i} className={styles.evidenceCard}>
                      <div className={styles.evidenceHeader}>
                        <span className={styles.evidenceSource}>{ev.source}</span>
                        <span className={styles.evidenceSignal}>{ev.signal}</span>
                      </div>
                      <div className={styles.evidenceValue}>{ev.value}</div>
                      <div className={styles.evidenceTimestamp}>{new Date(ev.timestamp).toLocaleString()}</div>
                    </div>
                  ))}
                </div>
              </section>

              {/* AI RCA */}
              <section className={styles.sectionCard}>
                <h2 className={styles.sectionTitle}>AI Root Cause Analysis</h2>
                {rca ? (
                  <div className={styles.rcaPanel}>
                    <div className={styles.rcaHeader}>
                      <span className={styles.rcaRootCause}>{rca.root_cause}</span>
                      <div className={styles.confidenceBar}>
                        <span className={styles.confidenceLabel}>Confidence</span>
                        <div className={styles.confidenceTrack}>
                          <div
                            className={styles.confidenceFill}
                            style={{ width: `${Math.round(rca.confidence_score * 100)}%` }}
                          />
                        </div>
                        <span className={styles.confidenceScore}>{Math.round(rca.confidence_score * 100)}%</span>
                      </div>
                    </div>
                    <p className={styles.rcaSummary}>{rca.summary}</p>
                    <div className={styles.rcaGenerated}>Generated at {new Date(rca.generated_at).toLocaleString()}</div>
                    {rca.supporting_evidence.length > 0 && (
                      <>
                        <h3 className={styles.subHeading}>Supporting Evidence</h3>
                        <div className={styles.evidenceList}>
                          {rca.supporting_evidence.map((ev, i) => (
                            <div key={i} className={styles.evidenceCard}>
                              <div className={styles.evidenceHeader}>
                                <span className={styles.evidenceSource}>{ev.source}</span>
                                <span className={styles.evidenceSignal}>{ev.signal}</span>
                              </div>
                              <div className={styles.evidenceValue}>{ev.value}</div>
                              <div className={styles.evidenceTimestamp}>{new Date(ev.timestamp).toLocaleString()}</div>
                            </div>
                          ))}
                        </div>
                      </>
                    )}
                  </div>
                ) : diagnosing ? (
                  <Spinner label="Running AI diagnosis… this takes ~15 seconds." />
                ) : rcaTimedOut ? (
                  <div>
                    <p className={styles.errorText}>No diagnosis available for this incident.</p>
                    {diagnosisError && <ErrorBanner message={diagnosisError} />}
                    <button className={styles.approveBtn} onClick={handleDiagnose}>
                      <LuBrainCircuit size={15} /> Run Diagnosis
                    </button>
                  </div>
                ) : (
                  <Spinner label="Waiting for AI diagnosis…" />
                )}
              </section>

              {/* Remediation */}
              <section className={styles.sectionCard}>
                <h2 className={styles.sectionTitle}>Remediation Proposal</h2>
                {remediation ? (
                  <div className={styles.remediationPanel}>
                    <div className={styles.remediationHeader}>
                      <span className={styles.runbookName}>{remediation.action.runbook_name}</span>
                      <span className={`${styles.riskBadge} ${riskClass(remediation.action.risk)}`}>
                        Risk: {remediation.action.risk}
                      </span>
                    </div>
                    <p className={styles.remediationDesc}>{remediation.action.description}</p>
                    <div className={styles.paramsBlock}>
                      {Object.entries(remediation.action.params).map(([k, v]) => (
                        <div key={k} className={styles.paramRow}>
                          <span className={styles.paramKey}>{k}</span>
                          <span className={styles.paramValue}>{v}</span>
                        </div>
                      ))}
                    </div>
                    <div className={styles.approvalRow}>
                      <span className={`${styles.approvalBadge} ${styles['approval_' + remediation.approval]}`}>
                        {remediation.approval}
                      </span>
                      {remediation.approval === 'pending' && (
                        <button
                          className={styles.approveBtn}
                          onClick={handleApprove}
                          disabled={approving}
                        >
                          {approving
                            ? <><LuLoader size={15} className={styles.spinIcon} /> Executing runbook…</>
                            : <><LuCircleCheck size={15} /> Approve Remediation</>}
                        </button>
                      )}
                      {executionResult && (
                        <p className={executionResult.success ? styles.successText : styles.errorText}>
                          {executionResult.success
                            ? `✓ ${executionResult.message}`
                            : `✗ ${executionResult.message}`}
                        </p>
                      )}
                      {approveError && !executionResult && <ErrorBanner message={approveError} />}
                    </div>
                  </div>
                ) : (
                  <Spinner label="Waiting for remediation proposal…" />
                )}
              </section>
            </>
          )}
        </div>
    </div>
  );
};

export default IncidentDrawer;
