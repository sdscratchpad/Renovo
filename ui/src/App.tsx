import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { WSProvider } from './context/WSContext';
import { DrawerProvider, useDrawer } from './context/DrawerContext';
import NavBar from './components/NavBar';
import SideNav from './components/SideNav';
import IncidentDrawer from './components/IncidentDrawer';
import Dashboard from './pages/Dashboard';
import IncidentDetail from './pages/IncidentDetail';
import FaultInjector from './pages/FaultInjector';
import AuditLog from './pages/AuditLog';
import AILog from './pages/AILog';
import styles from './App.module.css';

// Inner component so it can consume DrawerContext (providers must wrap consumers).
const AppBody: React.FC = () => {
  const { drawerIncidentId, closeDrawer } = useDrawer();
  return (
    <div className={styles.body}>
      <SideNav />
      <main className={styles.main}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/incidents/:id" element={<IncidentDetail />} />
          <Route path="/inject" element={<FaultInjector />} />
          <Route path="/audit" element={<AuditLog />} />
          <Route path="/ai-log" element={<AILog />} />
        </Routes>
      </main>
      <IncidentDrawer incidentId={drawerIncidentId} onClose={closeDrawer} />
    </div>
  );
};

const App: React.FC = () => {
  return (
    <WSProvider>
      <DrawerProvider>
        <BrowserRouter>
          <div className={styles.app}>
            <NavBar />
            <AppBody />
          </div>
        </BrowserRouter>
      </DrawerProvider>
    </WSProvider>
  );
};

export default App;
