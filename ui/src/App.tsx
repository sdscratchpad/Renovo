import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import NavBar from './components/NavBar';
import Dashboard from './pages/Dashboard';
import IncidentDetail from './pages/IncidentDetail';
import FaultInjector from './pages/FaultInjector';
import AuditLog from './pages/AuditLog';
import styles from './App.module.css';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <div className={styles.app}>
        <NavBar />
        <main className={styles.main}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/incidents/:id" element={<IncidentDetail />} />
            <Route path="/inject" element={<FaultInjector />} />
            <Route path="/audit" element={<AuditLog />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
};

export default App;
