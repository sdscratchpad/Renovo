import React from 'react';
import { NavLink } from 'react-router-dom';
import { LuLayoutDashboard, LuZap, LuClipboardList } from '../icons';
import RenovoLogo from './RenovoLogo';
import styles from './NavBar.module.css';

const NavBar: React.FC = () => {
  return (
    <nav className={styles.nav}>
      <div className={styles.brand}>
        <RenovoLogo size={34} />
        <div className={styles.brandText}>
          <span className={styles.brandName}>Renovo</span>
          <span className={styles.brandSub}>AI Infrastructure Resilience</span>
        </div>
      </div>
      <ul className={styles.links}>
        <li>
          <NavLink to="/" end className={({ isActive }) => isActive ? styles.activeLink : styles.link}>
            <LuLayoutDashboard size={15} />
            Dashboard
          </NavLink>
        </li>
        <li>
          <NavLink to="/inject" className={({ isActive }) => isActive ? styles.activeLink : styles.link}>
            <LuZap size={15} />
            Fault Injector
          </NavLink>
        </li>
        <li>
          <NavLink to="/audit" className={({ isActive }) => isActive ? styles.activeLink : styles.link}>
            <LuClipboardList size={15} />
            Audit Log
          </NavLink>
        </li>
      </ul>
    </nav>
  );
};

export default NavBar;
