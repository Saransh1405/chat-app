import type { WSEvent } from "@/types/chat";

type EventHandler = (payload: unknown) => void;

const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:8080/api/v1/ws";

class WebSocketClient {
  private ws: WebSocket | null = null;
  private handlers: Map<string, Set<EventHandler>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private shouldReconnect = true;

  connect(token: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    try {
      this.ws = new WebSocket(`${WS_URL}?token=${token}`);

      this.ws.onopen = () => {
        console.log("[WS] Connected");
        this.reconnectAttempts = 0;
      };

      this.ws.onmessage = (event) => {
        try {
          console.log("[WS] Raw message received:", event.data);
          const data: any = JSON.parse(event.data);
          console.log("[WS] Parsed message:", data);
          console.log("[WS] Message type:", data.type);
          console.log("[WS] Message payload:", data.payload);

          if (data.type && data.payload) {
            console.log(`[WS] Emitting event '${data.type}' with payload:`, data.payload);
            this.emit(data.type, data.payload);
          } else {
            console.warn("[WS] Message missing type or payload:", data);
          }
        } catch (err) {
          console.error("[WS] Failed to parse message:", err, "Raw data:", event.data);
        }
      };

      this.ws.onclose = () => {
        console.log("[WS] Disconnected");
        if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.reconnectAttempts++;
          const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
          console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
          setTimeout(() => this.connect(token), delay);
        }
      };

      this.ws.onerror = (error) => {
        console.error("[WS] Error:", error);
      };
    } catch (err) {
      console.error("[WS] Connection failed:", err);
    }
  }

  disconnect(): void {
    this.shouldReconnect = false;
    this.ws?.close();
    this.ws = null;
  }

  send(type: string, payload: unknown, roomId?: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message: any = { type, payload };
      if (roomId) message.room_id = roomId;
      this.ws.send(JSON.stringify(message));
    }
  }

  on(event: string, handler: EventHandler): () => void {
    console.log(`[WS] Registering handler for event: '${event}'`);
    if (!this.handlers.has(event)) {
      this.handlers.set(event, new Set());
    }
    this.handlers.get(event)!.add(handler);
    console.log(`[WS] Total handlers for '${event}':`, this.handlers.get(event)!.size);

    return () => {
      console.log(`[WS] Unregistering handler for event: '${event}'`);
      this.handlers.get(event)?.delete(handler);
    };
  }

  private emit(event: string, payload: unknown): void {
    const handlers = this.handlers.get(event);
    console.log(`[WS] Emitting event '${event}' to ${handlers?.size || 0} handlers`);
    handlers?.forEach((handler) => {
      try {
        handler(payload);
      } catch (err) {
        console.error(`[WS] Error in handler for event '${event}':`, err);
      }
    });
  }
}

export const wsClient = new WebSocketClient();
