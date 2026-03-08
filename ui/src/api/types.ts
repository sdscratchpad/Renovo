// TypeScript types mirroring contracts/types.go
// Field names match Go JSON tags (camelCase)

export type SeverityLevel = 'critical' | 'high' | 'medium' | 'low';

export interface EvidenceItem {
  source: string;    // "metrics", "logs", "events"
  signal: string;    // e.g. "cpu_usage_percent", "pod_restart_count"
  value: string;
  timestamp: string; // ISO 8601
}

export interface IncidentEvent {
  id: string;
  scenario: string;   // "bad-rollout" | "resource-saturation" | "batch-timeout"
  service: string;
  namespace: string;
  severity: SeverityLevel;
  detected_at: string; // ISO 8601
  evidence: EvidenceItem[];
  labels?: Record<string, string>;
}

export interface RCAPayload {
  incident_id: string;
  summary: string;
  root_cause: string;
  confidence_score: number; // 0.0 to 1.0
  supporting_evidence: EvidenceItem[];
  generated_at: string; // ISO 8601
}

export type RiskLevel = 'low' | 'medium' | 'high';

export interface RemediationAction {
  id: string;
  incident_id: string;
  runbook_name: string;
  description: string;
  risk: RiskLevel;
  params: Record<string, string>;
  created_at: string; // ISO 8601
}

export type ApprovalStatus = 'pending' | 'approved' | 'rejected';

export interface RemediationRequest {
  action: RemediationAction;
  approval: ApprovalStatus;
  updated_at: string; // ISO 8601
}

export interface RemediationResult {
  action_id: string;
  incident_id: string;
  success: boolean;
  message: string;
  executed_at: string;  // ISO 8601
  recovered_at?: string; // ISO 8601
}

export type IncidentStatus =
  | 'detected'
  | 'analyzing'
  | 'awaiting_approval'
  | 'remediating'
  | 'resolved'
  | 'failed';

export interface IncidentStatusUpdate {
  incident_id: string;
  status: IncidentStatus;
  updated_at: string; // ISO 8601
}

export interface AuditEntry {
  id: string;
  incident_id: string;
  actor: string;  // "system" | "operator"
  event: string;  // "incident_detected" | "rca_generated" | "action_approved" | "action_executed"
  detail: string;
  timestamp: string; // ISO 8601
}

export interface KPISnapshot {
  incident_id: string;
  mttd: number;  // nanoseconds (Go time.Duration)
  mttr: number;  // nanoseconds (Go time.Duration)
  auto_resolved: boolean;
  estimated_cost_saved_usd: number;
}

export interface KPISummary {
  mttd_minutes: number;
  mttr_minutes: number;
  resolved_today: number;
  auto_resolve_rate: number; // percentage 0-100
}

export interface LLMInteraction {
  incident_id: string;
  model: string;
  system_prompt: string;
  user_prompt: string;
  raw_response: string;
  created_at: string; // ISO 8601
}
