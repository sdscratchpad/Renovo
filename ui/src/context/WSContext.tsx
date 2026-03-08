import React, { createContext, useContext, useEffect, useRef, useState } from 'react';

export type WSStatus = 'connecting' | 'connected' | 'disconnected';

interface WSContextValue {
  /** Current WebSocket connection state. */
  status: WSStatus;
  /**
   * Epoch-ms timestamp bumped on every server push.
   * Consumers can use this in a useEffect dependency array to trigger
   * an immediate data refresh whenever the backend writes new data.
   */
  lastEventAt: number;
}

const WSContext = createContext<WSContextValue>({ status: 'connecting', lastEventAt: 0 });

const WS_URL = 'ws://localhost:8085/ws';
const RECONNECT_DELAY_MS = 2000;

export const WSProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [status, setStatus] = useState<WSStatus>('connecting');
  const [lastEventAt, setLastEventAt] = useState(0);
  const wsRef = useRef<WebSocket | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const destroyedRef = useRef(false);

  useEffect(() => {
    destroyedRef.current = false;

    function connect() {
      if (destroyedRef.current) return;
      setStatus('connecting');

      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        if (!destroyedRef.current) setStatus('connected');
      };

      ws.onmessage = () => {
        if (!destroyedRef.current) setLastEventAt(Date.now());
      };

      ws.onclose = () => {
        if (!destroyedRef.current) {
          setStatus('disconnected');
          timerRef.current = setTimeout(connect, RECONNECT_DELAY_MS);
        }
      };

      ws.onerror = () => {
        // onclose fires immediately after onerror, which handles the reconnect.
        ws.close();
      };
    }

    connect();

    return () => {
      destroyedRef.current = true;
      if (timerRef.current) clearTimeout(timerRef.current);
      wsRef.current?.close();
    };
  }, []);

  return (
    <WSContext.Provider value={{ status, lastEventAt }}>
      {children}
    </WSContext.Provider>
  );
};

/** Hook to read the shared WebSocket state from any component. */
export function useWS(): WSContextValue {
  return useContext(WSContext);
}
