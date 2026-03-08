import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  LuArrowLeft, LuBrainCircuit, LuCircleCheck, LuLoader, LuShieldAlert,
} from '../icons';
import { getIncidents, getRCA, getRemediation, approveRemediation, triggerDiagnosis, getRemediationResult } from '../api/client';
import { IncidentEvent, RCAPayload, RemediationRequest, RemediationResult } from '../api/types';
import Spinner from '../components/Spinner';
import ErrorBanner from '../components/ErrorBanner';
import StatusBadge, { deriveStatus } from '../components/StatusBadge';
import styles from './IncidentDetail.module.css';

function riskClass(risk: string): string {
  switch (risk) {
    case 'high':   return styles.riskHigh;
    case 'medium': return styles.riskMedium;
    default:       return styles.riskLow;
  }
}

const IncidentDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [incident, setIncident] = useState<IncidentEvent | null>(null);
  const [rca, setRca] = useState<RCAPayload | null>(null);
  const [remediation, setRemediation] = useState<RemediationRequest | null>(null);
  const [approving, setApproving] = useState(false);
  const [approveError, setApproveError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [diagnosing, setDiagnosing] = useState(false);
  const [diagnosisError, setDiagnosisError] = useState<string | null>(null);
  const [rcaTimedOut, setRcaTimedOut] = useState(false);
  const [executionResult, setExecutionResult] = useState<RemediationResult | null>(null);

  useEffect(() => {
    if (!id) return;
    Promise.all([getIncidents(), getRCA(id), getRemediation(id), getRemediationResult(id)])
      .then(([incidents, r, rem, execResult]) => {
        setIncident(incidents.find(i => i.id === id) ?? null);
        setRca(r);
        setRemediation(rem);
        if (execResult) setExecutionResult(execResult);
        if (!r) setRcaTimedOut(false); // will poll
      })
      .catch(e => setLoadError(e instanceof Error ? e.message : 'Failed to load incident.'))
      .finally(() => setLoading(false));
  }, [id]);

  // Poll for RCA every 3 s for up to 60 s after initial load if none exists yet.
  useEffect(() => {
    if (!id || rca || loading || rcaTimedOut) return;
    let attempts = 0;
    const MAX = 20; // 20 × 3 s = 60 s
    const timer = setInterval(async () => {
      attempts++;
      const r = await getRCA(id).catch(() => null);
      if (r) {
        setRca(r);
        setDiagnosing(false);
        clearInterval(timer);
        // Also refresh remediation now that RCA is ready.
        getRemediation(id).then(rem => { if (rem) setRemediation(rem); }).catch(() => {});
        return;
      }
      if (attempts >= MAX) {
        setRcaTimedOut(true);
        setDiagnosing(false);
        clearInterval(timer);
      }
    }, 3000);
    return () => clearInterval(timer);
  }, [id, rca, loading, rcaTimedOut]);

  // Poll for execution result while remediation is approved but result not yet loaded.
  useEffect(() => {
    if (!id || executionResult || remediation?.approval !== 'approved' || loading) return;
    const timer = setInterval(async () => {
      const result = await getRemediationResult(id).catch(() => null);
      if (result) {
        setExecutionResult(result);
        clearInterval(timer);
      }
    }, 2000);
    return () => clearInterval(timer);
  }, [id, executionResult, remediation?.approval, loading]);

  const handleDiagnose = async () => {
    if (!incident) return;
    setDiagnosing(true);
    setDiagnosisError(null);
    setRcaTimedOut(false);
    try {
      await triggerDiagnosis(incident);
      // Polling effect will pick up the result automatically.
    } catch (e) {
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
      const result = await approveRemediation(remediation.action.id);
      setExecutionResult(result);
      setRemediation(prev => prev ? { ...prev, approval: 'approved', updated_at: new Date().toISOString() } : prev);
      if (!result.success) {
        setApproveError(`Runbook failed: ${result.message}`);
      }
    } catch (e) {
      setApproveError('Failed to approve. Please try again.');
    } finally {
      setApproving(false);
    }
  };

  if (loading) {
    return <div className={styles.page}><Spinner label="Loading incident…" /></div>;
  }
  if (loadError) {
    return (
      <div className={styles.page}>
        <ErrorBanner message={loadError} />
        <Link to="/" className={styles.backLink}>← Back to Dashboard</Link>
      </div>
    );
  }
  if (!incident) {
    return (
      <div className={styles.page}>
        <p className={styles.errorText}>Incident not found.</p>
        <Link to="/" className={styles.backLink}>← Back to Dashboard</Link>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.pageHeader}>
        <Link to="/" className={styles.backLink}><LuArrowLeft size={14}/> Dashboard</Link>
        <h1 className={styles.heading}>Incident <span className={styles.headingId}>{incident.id}</span></h1>
        <StatusBadge status={deriveStatus(rca, remediation, executionResult)} />
      </div>

      <section className={styles.sectionCard}>
        <h2 className={styles.sectionTitle}>Overview</h2>
        <dl className={styles.metaGrid}>
          <div className={styles.metaItem}>
            <dt>Service</dt>
            <dd>{incident.service}</dd>
          </div>
          <div className={styles.metaItem}>
            <dt>Namespace</dt>
            <dd>{incident.namespace}</dd>
          </div>
          <div className={styles.metaItem}>
            <dt>Scenario</dt>
            <dd>{incident.scenario}</dd>
          </div>
          <div className={styles.metaItem}>
            <dt>Severity</dt>
            <dd><span className={`${styles.severityBadge} ${styles['severity_' + incident.severity]}`}>{incident.severity}</span></dd>
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

      {/* AI RCA Panel */}
      <section className={styles.sectionCard}>
        <h2 className={styles.sectionTitle}>AI Root Cause Analysis</h2>
        {rca ? (          <div className={styles.rcaPanel}>
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

      {/* Remediation Proposal */}
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
    </div>
  );
};

export default IncidentDetail;
