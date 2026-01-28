import { WebSocketMessage } from '../types';

const WS_BASE_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/api/v1/ws';

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners: Map<string, Set<(data: any) => void>> = new Map();
  private isConnecting = false;

  connect(token: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
        console.log('🔌 WebSocket already connected or connecting, skipping...');
        resolve();
        return;
      }

      console.log('🚀 Starting WebSocket connection...');
      console.log('📍 WebSocket URL:', WS_BASE_URL);
      this.isConnecting = true;
      // Note: WebSocket API in browsers doesn't support custom headers
      // The backend needs to accept token in query parameter for WebSocket connections
      // For now, we'll pass it in the URL. If backend requires Authorization header,
      // you may need to modify the backend to accept token in query params for WS connections
      const url = `${WS_BASE_URL}?token=${encodeURIComponent(token)}`;
      console.log('🔗 Connecting to:', url.replace(/token=[^&]+/, 'token=***'));
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        console.log('✅ WebSocket connected successfully');
        console.log('📊 WebSocket readyState:', this.ws?.readyState, '(OPEN = 1)');
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        resolve();
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          console.log('Raw WebSocket message received:', message);
          this.handleMessage(message);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error, event.data);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.isConnecting = false;
        reject(error);
      };

      this.ws.onclose = (event) => {
        console.log('🔌 WebSocket disconnected');
        console.log('📊 Close code:', event.code, 'Reason:', event.reason || 'No reason provided');
        console.log('🔄 Was clean close:', event.wasClean);
        this.isConnecting = false;
        this.handleReconnect(token);
      };
    });
  }

  private handleReconnect(token: string) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      console.log(`Reconnecting in ${delay}ms... (attempt ${this.reconnectAttempts})`);
      setTimeout(() => {
        this.connect(token).catch(console.error);
      }, delay);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  private handleMessage(message: WebSocketMessage) {
    console.log('🔔 Handling WebSocket message - Type:', message.type, 'Payload:', message.payload);
    const listeners = this.listeners.get(message.type);
    if (listeners) {
      console.log(`📢 Found ${listeners.size} listeners for event: ${message.type}`);
      listeners.forEach((listener) => {
        try {
          listener(message.payload);
        } catch (error) {
          console.error('❌ Error in WebSocket listener:', error);
        }
      });
    } else {
      console.warn(`⚠️ No listeners found for event type: ${message.type}`);
      console.log('Available listeners:', Array.from(this.listeners.keys()));
    }

    // Also trigger wildcard listeners
    const wildcardListeners = this.listeners.get('*');
    if (wildcardListeners) {
      wildcardListeners.forEach((listener) => {
        try {
          listener(message);
        } catch (error) {
          console.error('❌ Error in wildcard WebSocket listener:', error);
        }
      });
    }
  }

  subscribe(roomId: string) {
    console.log('📡 Attempting to subscribe to room:', roomId);
    console.log('📊 WebSocket state:', this.ws?.readyState, '(OPEN = 1, CONNECTING = 0, CLOSING = 2, CLOSED = 3)');
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const subscribeMsg = {
        type: 'subscribe',
        room_id: roomId,
      };
      console.log('📤 Sending subscribe message:', subscribeMsg);
      this.ws.send(JSON.stringify(subscribeMsg));
      console.log('✅ Subscribe message sent successfully');
    } else {
      console.warn('⚠️ Cannot subscribe: WebSocket not open. State:', this.ws?.readyState);
      console.warn('💡 WebSocket states: 0=CONNECTING, 1=OPEN, 2=CLOSING, 3=CLOSED');
    }
  }

  unsubscribe(roomId: string) {
    console.log('📡 Attempting to unsubscribe from room:', roomId);
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const unsubscribeMsg = {
        type: 'unsubscribe',
        room_id: roomId,
      };
      console.log('📤 Sending unsubscribe message:', unsubscribeMsg);
      this.ws.send(JSON.stringify(unsubscribeMsg));
      console.log('✅ Unsubscribe message sent successfully');
    } else {
      console.warn('⚠️ Cannot unsubscribe: WebSocket not open. State:', this.ws?.readyState);
    }
  }

  sendMessage(roomId: string, payload: any) {
    console.log('📤 Attempting to send message to room:', roomId);
    console.log('📊 Message payload:', payload);
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message = {
        type: 'message',
        room_id: roomId,
        payload,
      };
      console.log('📤 Sending message:', JSON.stringify(message));
      this.ws.send(JSON.stringify(message));
      console.log('✅ Message sent successfully');
    } else {
      console.warn('⚠️ Cannot send message: WebSocket not open. State:', this.ws?.readyState);
    }
  }

  on(event: string, callback: (data: any) => void) {
    console.log('👂 Registering listener for event:', event);
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);
    console.log('✅ Listener registered. Total listeners for', event + ':', this.listeners.get(event)!.size);
    console.log('📋 All registered events:', Array.from(this.listeners.keys()));
  }

  off(event: string, callback: (data: any) => void) {
    const listeners = this.listeners.get(event);
    if (listeners) {
      listeners.delete(callback);
    }
  }

  disconnect() {
    console.log('🔌 Disconnecting WebSocket...');
    if (this.ws) {
      console.log('📊 WebSocket state before close:', this.ws.readyState);
      this.ws.close();
      this.ws = null;
      console.log('✅ WebSocket closed and cleared');
    }
    this.listeners.clear();
    this.reconnectAttempts = 0;
    console.log('🧹 All listeners cleared');
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

export const wsClient = new WebSocketClient();

