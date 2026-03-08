import {
  IncidentEvent,
  RCAPayload,
  RemediationRequest,
  AuditEntry,
  KPISummary,
} from './types';

// ---- Incidents ----

export const mockIncidents: IncidentEvent[] = [
  {
    id: 'inc-001',
    scenario: 'bad-rollout',
    service: 'sample-app',
    namespace: 'default',
    severity: 'critical',
    detected_at: '2026-03-07T09:05:00Z',
    evidence: [
      {
        source: 'metrics',
        signal: 'http_error_rate',
        value: '42%',
        timestamp: '2026-03-07T09:04:50Z',
      },
      {
        source: 'logs',
        signal: 'pod_crash_loop',
        value: 'CrashLoopBackOff x5',
        timestamp: '2026-03-07T09:04:55Z',
      },
    ],
    labels: { env: 'prod', team: 'platform' },
  },
  {
    id: 'inc-002',
    scenario: 'resource-saturation',
    service: 'sample-app',
    namespace: 'default',
    severity: 'high',
    detected_at: '2026-03-07T10:30:00Z',
    evidence: [
      {
        source: 'metrics',
        signal: 'cpu_usage_percent',
        value: '97%',
        timestamp: '2026-03-07T10:29:45Z',
      },
      {
        source: 'metrics',
        signal: 'memory_usage_percent',
        value: '89%',
        timestamp: '2026-03-07T10:29:50Z',
      },
    ],
    labels: { env: 'prod', team: 'platform' },
  },
];

// ---- RCA ----

export const mockRCAs: Record<string, RCAPayload> = {
  'inc-001': {
    incident_id: 'inc-001',
    summary:
      'A misconfigured deployment of sample-app v2.1.0 introduced an invalid environment variable reference causing all pods to crash on startup. The 42% HTTP error rate is a direct consequence of the unavailable pods.',
    root_cause: 'invalid-env-config-bad-rollout',
    confidence_score: 0.94,
    supporting_evidence: [
      {
        source: 'logs',
        signal: 'container_exit_code',
        value: 'exit code 1: missing required env var DATABASE_URL',
        timestamp: '2026-03-07T09:04:55Z',
      },
      {
        source: 'events',
        signal: 'deployment_revision',
        value: 'sample-app rolled out revision 12 at 09:03:10Z',
        timestamp: '2026-03-07T09:03:10Z',
      },
    ],
    generated_at: '2026-03-07T09:05:45Z',
  },
  'inc-002': {
    incident_id: 'inc-002',
    summary:
      'CPU saturation reached 97% on all sample-app pods due to a sudden traffic spike (3x baseline). Current replica count of 2 is insufficient; scaling to 4 replicas should restore normal latency.',
    root_cause: 'insufficient-replicas-traffic-spike',
    confidence_score: 0.88,
    supporting_evidence: [
      {
        source: 'metrics',
        signal: 'cpu_usage_percent',
        value: '97%',
        timestamp: '2026-03-07T10:29:45Z',
      },
      {
        source: 'metrics',
        signal: 'request_rate_rps',
        value: '1850 rps (baseline: 620 rps)',
        timestamp: '2026-03-07T10:29:50Z',
      },
    ],
    generated_at: '2026-03-07T10:31:00Z',
  },
};

// ---- Remediations ----

export const mockRemediations: Record<string, RemediationRequest> = {
  'inc-001': {
    action: {
      id: 'rem-001',
      incident_id: 'inc-001',
      runbook_name: 'rollback-deployment',
      description: 'Roll back sample-app to the previous stable revision (rev 11).',
      risk: 'low',
      params: {
        namespace: 'default',
        deployment: 'sample-app',
        target_revision: '11',
      },
      created_at: '2026-03-07T09:05:50Z',
    },
    approval: 'pending',
    updated_at: '2026-03-07T09:05:50Z',
  },
  'inc-002': {
    action: {
      id: 'rem-002',
      incident_id: 'inc-002',
      runbook_name: 'scale-deployment',
      description: 'Scale sample-app deployment from 2 to 4 replicas to handle the traffic spike.',
      risk: 'low',
      params: {
        namespace: 'default',
        deployment: 'sample-app',
        replicas: '4',
      },
      created_at: '2026-03-07T10:31:05Z',
    },
    approval: 'pending',
    updated_at: '2026-03-07T10:31:05Z',
  },
};

// ---- Audit entries ----

export const mockAuditEntries: AuditEntry[] = [
  {
    id: 'aud-001',
    incident_id: 'inc-001',
    actor: 'system',
    event: 'incident_detected',
    detail: 'Anomaly detected: http_error_rate=42% on sample-app',
    timestamp: '2026-03-07T09:05:00Z',
  },
  {
    id: 'aud-002',
    incident_id: 'inc-001',
    actor: 'system',
    event: 'rca_generated',
    detail: 'RCA completed with confidence 0.94: invalid-env-config-bad-rollout',
    timestamp: '2026-03-07T09:05:45Z',
  },
  {
    id: 'aud-003',
    incident_id: 'inc-001',
    actor: 'operator',
    event: 'action_approved',
    detail: 'Operator approved rollback-deployment for sample-app rev 11',
    timestamp: '2026-03-07T09:07:10Z',
  },
  {
    id: 'aud-004',
    incident_id: 'inc-001',
    actor: 'system',
    event: 'action_executed',
    detail: 'rollback-deployment executed successfully; sample-app restored',
    timestamp: '2026-03-07T09:07:35Z',
  },
  {
    id: 'aud-005',
    incident_id: 'inc-002',
    actor: 'system',
    event: 'incident_detected',
    detail: 'Anomaly detected: cpu_usage_percent=97% on sample-app',
    timestamp: '2026-03-07T10:30:00Z',
  },
  {
    id: 'aud-006',
    incident_id: 'inc-002',
    actor: 'system',
    event: 'rca_generated',
    detail: 'RCA completed with confidence 0.88: insufficient-replicas-traffic-spike',
    timestamp: '2026-03-07T10:31:00Z',
  },
];

// ---- KPI ----

export const mockKPIs: KPISummary = {
  mttd_minutes: 1.5,
  mttr_minutes: 2.6,
  resolved_today: 1,
  auto_resolve_rate: 85,
};

// ---- Typed async API functions (mock implementations) ----

export async function getIncidents(): Promise<IncidentEvent[]> {
  return Promise.resolve(mockIncidents);
}

export async function getRCA(incidentId: string): Promise<RCAPayload | null> {
  return Promise.resolve(mockRCAs[incidentId] ?? null);
}

export async function getRemediation(incidentId: string): Promise<RemediationRequest | null> {
  return Promise.resolve(mockRemediations[incidentId] ?? null);
}

export async function approveRemediation(remediationId: string): Promise<void> {
  // Mock: find and update the in-memory record
  for (const key of Object.keys(mockRemediations)) {
    if (mockRemediations[key].action.id === remediationId) {
      mockRemediations[key] = {
        ...mockRemediations[key],
        approval: 'approved',
        updated_at: new Date().toISOString(),
      };
    }
  }
  return Promise.resolve();
}

export async function injectFault(scenario: string): Promise<void> {
  return Promise.resolve();
}

export async function getAuditEntries(): Promise<AuditEntry[]> {
  return Promise.resolve(mockAuditEntries);
}

export async function getKPIs(): Promise<KPISummary> {
  return Promise.resolve(mockKPIs);
}
