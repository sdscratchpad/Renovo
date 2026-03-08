import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { LuBrainCircuit, LuClock, LuChevronDown, LuChevronRight } from '../icons';
import { getLLMInteractions } from '../api/client';
import { LLMInteraction } from '../api/types';
import Spinner from '../components/Spinner';
import ErrorBanner from '../components/ErrorBanner';
import styles from './AILog.module.css';

// ── JSON syntax highlighter ──────────────────────────────────────────────────

function highlightJSON(raw: string): string {
  let json: string;
  try {
    json = JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    // not valid JSON — render as plain text
    return escapeHtml(raw);
  }

  return json
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(
      /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
      (match) => {
        if (/^"/.test(match)) {
          if (/:$/.test(match)) {
            return `<span class="json-key">${match}</span>`;
          }
          return `<span class="json-string">${match}</span>`;
        }
        if (/true|false/.test(match)) return `<span class="json-bool">${match}</span>`;
        if (/null/.test(match)) return `<span class="json-null">${match}</span>`;
        return `<span class="json-number">${match}</span>`;
      }
    );
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

// ── Sub-components ───────────────────────────────────────────────────────────

interface SectionProps {
  title: string;
  content: string;
  mono?: boolean;
  highlight?: boolean;
}

function Section({ title, content, mono, highlight }: SectionProps) {
  const [open, setOpen] = useState(false);
  return (
    <div className={styles.section}>
      <button className={styles.sectionToggle} onClick={() => setOpen(o => !o)}>
        {open ? <LuChevronDown size={13} /> : <LuChevronRight size={13} />}
        <span>{title}</span>
      </button>
      {open && (
        highlight
          ? (
            <pre
              className={`${styles.codeBlock} ${styles.jsonBlock}`}
              dangerouslySetInnerHTML={{ __html: highlightJSON(content) }}
            />
          )
          : (
            <pre className={`${styles.codeBlock} ${mono ? styles.monoBlock : ''}`}>
              {content}
            </pre>
          )
      )}
    </div>
  );
}

// ── Main component ───────────────────────────────────────────────────────────

const AILog: React.FC = () => {
  const [interactions, setInteractions] = useState<LLMInteraction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getLLMInteractions()
      .then(setInteractions)
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load AI log.'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className={styles.page}><Spinner /></div>;
  if (error)   return <div className={styles.page}><ErrorBanner message={error} /></div>;

  return (
    <div className={styles.page}>
      <div className={styles.pageHeader}>
        <h1 className={styles.heading}>
          <LuBrainCircuit size={22} /> AI Log
        </h1>
        <p className={styles.intro}>
          Every GPT-4o prompt and raw response recorded during incident analysis.
        </p>
      </div>

      {interactions.length === 0 ? (
        <div className={styles.empty}>No AI interactions logged yet.</div>
      ) : (
        <div className={styles.list}>
          {interactions.map(item => (
            <div key={item.incident_id} className={styles.card}>
              <div className={styles.cardHeader}>
                <div className={styles.cardMeta}>
                  <span className={styles.modelBadge}>{item.model}</span>
                  <Link to={`/incidents/${item.incident_id}`} className={styles.incidentLink}>
                    {item.incident_id}
                  </Link>
                </div>
                <span className={styles.timestamp}>
                  <LuClock size={11} />
                  {new Date(item.created_at).toLocaleString()}
                </span>
              </div>

              <div className={styles.sections}>
                <Section title="System Prompt"   content={item.system_prompt} mono />
                <Section title="User Prompt"     content={item.user_prompt}   mono />
                <Section title="Model Response"  content={item.raw_response}  highlight />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default AILog;
