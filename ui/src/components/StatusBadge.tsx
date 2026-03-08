import React from 'react';
import {
  LuCircleAlert,
  LuClock,
  LuCpu,
  LuRefreshCw,
  LuCircleCheck,
  LuCircleX,
  LuScanSearch,
} from '../icons';
import { RCAPayload, RemediationRequest } from '../api/types';
import styles from './StatusBadge.module.css';

export type IncidentStatus = 'detected' | 'diagnosing' | 'awaiting_approval' | 'remediating' | 'verifying' | 'resolved' | 'failed';

export function deriveStatus(
  rca: RCAPayload | null,
  remediation: RemediationRequest | null,
  executionResult?: { success: boolean } | null,
  pipelineStatus?: string | null,
): IncidentStatus {
  if (!rca) return 'detected';
  if (!remediation) return 'diagnosing';
  if (executionResult != null) return executionResult.success ? 'resolved' : 'failed';
  if (remediation.approval === 'approved') {
    if (pipelineStatus === 'verifying') return 'verifying';
    return 'remediating';
  }
  if (remediation.approval === 'rejected') return 'resolved';
  return 'awaiting_approval';
}

interface StatusBadgeProps {
  status: IncidentStatus;
}

const LABELS: Record<IncidentStatus, string> = {
  detected:          'Detected',
  diagnosing:        'Diagnosing',
  awaiting_approval: 'Awaiting Approval',
  remediating:       'Remediating',
  verifying:         'Verifying',
  resolved:          'Resolved',
  failed:            'Failed',
};

const STATUS_ICONS: Record<IncidentStatus, React.ReactNode> = {
  detected:          <LuCircleAlert size={11} />,
  diagnosing:        <LuCpu size={11} />,
  awaiting_approval: <LuClock size={11} />,
  remediating:       <LuRefreshCw size={11} className={styles.spinIcon} />,
  verifying:         <LuScanSearch size={11} className={styles.pulseIcon} />,
  resolved:          <LuCircleCheck size={11} />,
  failed:            <LuCircleX size={11} />,
};

const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => (
  <span className={`${styles.badge} ${styles[status]}`}>
    {STATUS_ICONS[status]}
    {LABELS[status]}
  </span>
);

export default StatusBadge;
