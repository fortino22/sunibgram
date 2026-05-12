type Handler = (data: unknown) => void;

class SocketManager {
  private ws: WebSocket | null = null;
  private listeners = new Map<string, Set<Handler>>();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private currentUrl = '';
  private active = false;

  connect(token: string, baseUrl: string) {
    const wsBase = baseUrl.replace(/\/$/, '').replace(/^http/, 'ws');
    this.currentUrl = `${wsBase}/ws?token=${encodeURIComponent(token)}`;
    this.active = true;
    this.open();
  }

  disconnect() {
    this.active = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.ws?.close();
    this.ws = null;
  }

  emit(event: string, roomId?: number) {
    if (this.ws?.readyState !== WebSocket.OPEN) return;
    const msg: Record<string, unknown> = { event };
    if (roomId !== undefined) msg.room_id = roomId;
    this.ws.send(JSON.stringify(msg));
  }

  on(event: string, handler: Handler): () => void {
    if (!this.listeners.has(event)) this.listeners.set(event, new Set());
    this.listeners.get(event)!.add(handler);
    return () => this.off(event, handler);
  }

  off(event: string, handler: Handler) {
    this.listeners.get(event)?.delete(handler);
  }

  get connected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private open() {
    if (!this.active) return;
    if (
      this.ws?.readyState === WebSocket.OPEN ||
      this.ws?.readyState === WebSocket.CONNECTING
    ) {
      return;
    }

    this.ws = new WebSocket(this.currentUrl);

    this.ws.onmessage = (e) => {
      try {
        const { event, data } = JSON.parse(e.data as string) as {
          event: string;
          data: unknown;
        };
        this.listeners.get(event)?.forEach((fn) => fn(data));
      } catch {
        // ignore malformed frames
      }
    };

    this.ws.onclose = () => {
      if (this.active) {
        this.reconnectTimer = setTimeout(() => this.open(), 3000);
      }
    };

    this.ws.onerror = () => this.ws?.close();
  }
}

export const socketManager = new SocketManager();
