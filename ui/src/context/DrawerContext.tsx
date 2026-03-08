import React, { createContext, useContext, useState } from 'react';

interface DrawerCtx {
  drawerIncidentId: string | null;
  openDrawer: (id: string) => void;
  closeDrawer: () => void;
}

const DrawerContext = createContext<DrawerCtx>({
  drawerIncidentId: null,
  openDrawer: () => {},
  closeDrawer: () => {},
});

export const DrawerProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [drawerIncidentId, setDrawerIncidentId] = useState<string | null>(null);
  return (
    <DrawerContext.Provider value={{
      drawerIncidentId,
      openDrawer: (id) => setDrawerIncidentId(id),
      closeDrawer: () => setDrawerIncidentId(null),
    }}>
      {children}
    </DrawerContext.Provider>
  );
};

export const useDrawer = () => useContext(DrawerContext);
