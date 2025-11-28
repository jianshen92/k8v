export function createResourceSocket(state, handlers) {
  let socket = null;

  function connect() {
    const myConnectionId = ++state.ws.connectionId;
    const url = handlers.buildUrl();
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    socket = new WebSocket(`${protocol}//${window.location.host}${url}`);

    socket.onopen = () => {
      if (myConnectionId !== state.ws.connectionId) {
        socket.close();
        return;
      }
      clearTimeout(window.snapshotTimer);
      state.snapshotComplete = false;
      state.snapshotCount = 0;
      state.ws.manual = false;
      handlers.onOpen?.();
    };

    socket.onmessage = (event) => {
      if (myConnectionId !== state.ws.connectionId) return;
      const msg = JSON.parse(event.data);
      if (!state.snapshotComplete && msg.type === 'ADDED') {
        state.snapshotCount++;
        clearTimeout(window.snapshotTimer);
        window.snapshotTimer = setTimeout(() => {
          if (!state.snapshotComplete) {
            state.snapshotComplete = true;
            console.log(`[WS] Snapshot complete: ${state.snapshotCount} resources loaded`);
          }
        }, 900);
      }
      handlers.onMessage?.(msg);
    };

    socket.onerror = () => {
      if (myConnectionId !== state.ws.connectionId) return;
      handlers.onError?.();
    };

    socket.onclose = () => {
      if (myConnectionId !== state.ws.connectionId) return;
      handlers.onClose?.();
      if (!state.ws.manual) {
        state.ws.reconnectTimeout = setTimeout(connect, 2000);
      }
    };
  }

  function disconnect(manual = false) {
    state.ws.manual = manual;
    if (state.ws.reconnectTimeout) {
      clearTimeout(state.ws.reconnectTimeout);
      state.ws.reconnectTimeout = null;
    }
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      socket.close();
    }
  }

  return { connect, disconnect };
}
