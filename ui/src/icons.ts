/**
 * Central icon re-exports typed as React.FC<IconBaseProps>.
 *
 * react-icons v5 changed IconType to `(props) => ReactNode`, which TypeScript 4.9.5
 * rejects in JSX positions (ReactNode includes undefined, incompatible with JSX.Element).
 * Re-casting to React.FC restores the ReactElement | null return type that TS 4.9.5 expects.
 */
import type { FC, SVGAttributes } from 'react';
import {
  LuActivity as _LuActivity,
  LuArrowLeft as _LuArrowLeft,
  LuArrowRight as _LuArrowRight,
  LuBot as _LuBot,
  LuBrainCircuit as _LuBrainCircuit,
  LuCheck as _LuCheck,
  LuCircleAlert as _LuCircleAlert,
  LuCircleCheck as _LuCircleCheck,
  LuCircleCheckBig as _LuCircleCheckBig,
  LuCircleX as _LuCircleX,
  LuClipboardList as _LuClipboardList,
  LuClock as _LuClock,
  LuCpu as _LuCpu,
  LuLayoutDashboard as _LuLayoutDashboard,
  LuLoader as _LuLoader,
  LuPlay as _LuPlay,
  LuRefreshCw as _LuRefreshCw,
  LuRotateCcw as _LuRotateCcw,
  LuServer as _LuServer,
  LuShieldAlert as _LuShieldAlert,
  LuTimer as _LuTimer,
  LuTrendingUp as _LuTrendingUp,
  LuUser as _LuUser,
  LuZap as _LuZap,
} from 'react-icons/lu';

type IconBaseProps = SVGAttributes<SVGElement> & {
  size?: string | number;
  color?: string;
  title?: string;
};

type Icon = FC<IconBaseProps>;

export const LuActivity       = _LuActivity       as Icon;
export const LuArrowLeft      = _LuArrowLeft      as Icon;
export const LuArrowRight     = _LuArrowRight     as Icon;
export const LuBot            = _LuBot            as Icon;
export const LuBrainCircuit   = _LuBrainCircuit   as Icon;
export const LuCheck          = _LuCheck          as Icon;
export const LuCircleAlert    = _LuCircleAlert    as Icon;
export const LuCircleCheck    = _LuCircleCheck    as Icon;
export const LuCircleCheckBig = _LuCircleCheckBig as Icon;
export const LuCircleX        = _LuCircleX        as Icon;
export const LuClipboardList  = _LuClipboardList  as Icon;
export const LuClock          = _LuClock          as Icon;
export const LuCpu            = _LuCpu            as Icon;
export const LuLayoutDashboard = _LuLayoutDashboard as Icon;
export const LuLoader         = _LuLoader         as Icon;
export const LuPlay           = _LuPlay           as Icon;
export const LuRefreshCw      = _LuRefreshCw      as Icon;
export const LuRotateCcw      = _LuRotateCcw      as Icon;
export const LuServer         = _LuServer         as Icon;
export const LuShieldAlert    = _LuShieldAlert    as Icon;
export const LuTimer          = _LuTimer          as Icon;
export const LuTrendingUp     = _LuTrendingUp     as Icon;
export const LuUser           = _LuUser           as Icon;
export const LuZap            = _LuZap            as Icon;
