import React from 'react';
import { NavLink } from 'react-router-dom';
import { LuLayoutDashboard, LuBrainCircuit, LuClipboardList } from '../icons';
import styles from './SideNav.module.css';

const NAV_ITEMS = [
  { to: '/',        end: true,  icon: <LuLayoutDashboard size={16} />, label: 'Dashboard'         },
  { to: '/ai-log',  end: false, icon: <LuBrainCircuit size={16} />,    label: 'AI Log'            },
  { to: '/audit',   end: false, icon: <LuClipboardList size={16} />,   label: 'Incident History'  },
];

const SideNav: React.FC = () => (
  <aside className={styles.sidenav}>
    <nav>
      <ul className={styles.list}>
        {NAV_ITEMS.map(({ to, end, icon, label }) => (
          <li key={to}>
            <NavLink
              to={to}
              end={end}
              className={({ isActive }) => `${styles.item} ${isActive ? styles.active : ''}`}
            >
              <span className={styles.icon}>{icon}</span>
              <span className={styles.label}>{label}</span>
            </NavLink>
          </li>
        ))}
      </ul>
    </nav>
  </aside>
);

export default SideNav;
