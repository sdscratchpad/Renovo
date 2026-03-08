import React from 'react';

interface RenovoLogoProps {
  size?: number;
}

/**
 * Renovo logo — a circular refresh arc (restore/renew) with a pulse beat
 * inside, evoking both infrastructure resilience and the healing loop.
 * Palette: electric-blue (#7aa2f7) accent on dark transparent background.
 */
const RenovoLogo: React.FC<RenovoLogoProps> = ({ size = 36 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 36 36"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    aria-label="Renovo logo"
  >
    {/* Background circle */}
    <circle cx="18" cy="18" r="17" fill="url(#bg)" />

    {/* Outer arc — open loop indicating the restore/renew cycle */}
    <path
      d="M18 4 C27.389 4 35 11.611 35 21 C35 30.389 27.389 38 18 38"
      stroke="url(#arcGrad)"
      strokeWidth="2.4"
      strokeLinecap="round"
      fill="none"
    />

    {/* Shorter counter arc for the other side — creates asymmetric motion feel */}
    <path
      d="M18 4 C8.611 4 1 11.611 1 21"
      stroke="#2d3f6e"
      strokeWidth="2.4"
      strokeLinecap="round"
      fill="none"
    />

    {/* Arrow on the open tail of the arc — clockwise, signals renewal */}
    <polyline
      points="15.5,35.5 18,38 20.5,35.5"
      stroke="url(#arcGrad)"
      strokeWidth="2.2"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />

    {/* Heartbeat / pulse line — infrastructure health monitor metaphor */}
    <polyline
      points="7,19 10,19 12,13 14.5,25 16.5,16 18.5,22 20.5,19 29,19"
      stroke="url(#pulseGrad)"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />

    <defs>
      <radialGradient id="bg" cx="50%" cy="40%" r="55%">
        <stop offset="0%" stopColor="#1e2340" />
        <stop offset="100%" stopColor="#0d1021" />
      </radialGradient>
      <linearGradient id="arcGrad" x1="18" y1="4" x2="32" y2="35" gradientUnits="userSpaceOnUse">
        <stop offset="0%" stopColor="#7aa2f7" />
        <stop offset="100%" stopColor="#2ac3de" />
      </linearGradient>
      <linearGradient id="pulseGrad" x1="7" y1="19" x2="29" y2="19" gradientUnits="userSpaceOnUse">
        <stop offset="0%" stopColor="#3d59a1" />
        <stop offset="45%" stopColor="#7aa2f7" />
        <stop offset="100%" stopColor="#2ac3de" />
      </linearGradient>
    </defs>
  </svg>
);

export default RenovoLogo;
