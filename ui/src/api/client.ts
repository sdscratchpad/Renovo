// client.ts — typed async fetch functions for all live service endpoints.
// All inter-service calls go through apiFetch so errors surface uniformly.

import { IncidentEvent, RCAPayload, RemediationRequest, RemediationResult, IncidentStatusUpdate, AuditEntry, KPISummary, KPISnapshot, LLMInteraction } from './types';

const EVENT_STORE   = 'http://localhost:8085';
const DIAGNOSIS     = 'http://localhost:8083';
const ORCHESTRATOR  = 'http://localhost:8084';
const FAULT_INJECTOR = 'http://localhost:8082';

// Nanoseconds → minutes (Go time.Duration is int64 nanoseconds)
const NS_PER_MINUTE = 60_000_000_000;

async function apiFetch<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, options);
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`API error ${res.status}: ${text}`);
  }
  return res.json() as Promise<T>;
}

// ---- Incidents ----

export async function getIncidents(): Promise<IncidentEvent[]> {
  return apiFetch<IncidentEvent[]>(`${EVENT_STORE}/incidents`);
}

// ---- Remediation results ----

export async function getRemediationResult(incidentId: string): Promise<RemediationResult | null> {
  try {
    return await apiFetch<RemediationResult>(`${EVENT_STORE}/remediation-results/${incidentId}`);
  } catch {
    return null;
  }
}

// ---- Incident pipeline status ----

export async function getIncidentStatus(incidentId: string): Promise<IncidentStatusUpdate | null> {
  try {
    return await apiFetch<IncidentStatusUpdate>(`${EVENT_STORE}/status/${incidentId}`);
  } catch {
    return null;
  }
}

// ---- RCA ----

export async function getRCA(incidentId: string): Promise<RCAPayload | null> {
  try {
    return await apiFetch<RCAPayload>(`${EVENT_STORE}/rca/${incidentId}`);
  } catch {
    return null;
  }
}

// Trigger AI diagnosis for an incident. Safe to call even if no RCA exists yet.
export async function triggerDiagnosis(incident: IncidentEvent): Promise<void> {
  await apiFetch<unknown>(`${DIAGNOSIS}/diagnose`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(incident),
  });
}

// ---- Remediations ----

// GET /remediations/{incident_id} returns an array; return the first (newest).
export async function getRemediation(incidentId: string): Promise<RemediationRequest | null> {
  const list = await apiFetch<RemediationRequest[]>(`${EVENT_STORE}/remediations/${incidentId}`);
  return list.length > 0 ? list[0] : null;
}

// PATCH /remediations/{id}/approve — triggers orchestrator to execute the runbook
// asynchronously. The orchestrator immediately returns 202 Accepted once the
// incident status is set to "remediating". The caller should poll
// /remediation-results/{incidentId} to detect when the runbook finishes.
export async function approveRemediation(remediationId: string): Promise<void> {
  const res = await fetch(`${ORCHESTRATOR}/remediations/${remediationId}/approve`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ approval: 'approved' }),
  });
  if (res.status === 202) return;
  const text = await res.text().catch(() => res.statusText);
  throw new Error(`API error ${res.status}: ${text}`);
}

// ---- Fault injection ----

export async function injectFault(scenario: string): Promise<void> {
  await apiFetch<unknown>(`${FAULT_INJECTOR}/inject/${scenario}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({}),
  });
}

// ---- Audit ----

// No global audit endpoint exists; fetch per-incident and flatten.
export async function getAuditEntries(): Promise<AuditEntry[]> {
  const incidents = await apiFetch<IncidentEvent[]>(`${EVENT_STORE}/incidents`);
  const batches = await Promise.all(
    incidents.map(inc =>
      apiFetch<AuditEntry[]>(`${EVENT_STORE}/audit/${inc.id}`).catch((): AuditEntry[] => [])
    )
  );
  return batches.flat();
}

// ---- LLM interaction log ----

export async function getLLMInteraction(incidentId: string): Promise<LLMInteraction | null> {
  try {
    return await apiFetch<LLMInteraction>(`${EVENT_STORE}/llm-log/${incidentId}`);
  } catch {
    return null;
  }
}

export async function getLLMInteractions(): Promise<LLMInteraction[]> {
  try {
    return await apiFetch<LLMInteraction[]>(`${EVENT_STORE}/llm-log`);
  } catch {
    return [];
  }
}

// ---- KPI ----

// Aggregate KPI snapshots across all incidents.
export async function getKPIs(): Promise<KPISummary> {
  const incidents = await apiFetch<IncidentEvent[]>(`${EVENT_STORE}/incidents`);
  const snapshots = await Promise.all(
    incidents.map(inc =>
      apiFetch<KPISnapshot>(`${EVENT_STORE}/kpi/${inc.id}`).catch((): null => null)
    )
  );
  const valid = snapshots.filter((s): s is KPISnapshot => s !== null);
  if (valid.length === 0) {
    return { mttd_minutes: 0, mttr_minutes: 0, resolved_today: 0, auto_resolve_rate: 0 };
  }

  const mttd_minutes =
    valid.reduce((sum, s) => sum + s.mttd / NS_PER_MINUTE, 0) / valid.length;

  const withMttr = valid.filter(s => s.mttr > 0);
  const mttr_minutes =
    withMttr.length > 0
      ? withMttr.reduce((sum, s) => sum + s.mttr / NS_PER_MINUTE, 0) / withMttr.length
      : 0;

  const resolved_today = valid.filter(s => s.auto_resolved).length;
  const auto_resolve_rate =
    valid.length > 0 ? Math.round((resolved_today / valid.length) * 100) : 0;

  return { mttd_minutes, mttr_minutes, resolved_today, auto_resolve_rate };
}
