import React, { useRef, useEffect, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { Toolbar } from './Toolbar';
import { UsersList } from './UsersList';
import { ColorPicker } from './ColorPicker';
import { LineWidthSlider } from './LineWidthSlider';
import { toast } from 'sonner';
import roomService from '@/services/roomService';

export type DrawingTool = 'pen' | 'eraser' | 'rectangle' | 'circle' | 'line';

export interface DrawMessage {
  type: 'draw_start' | 'draw_move' | 'draw_end' | 'shape' | 'cursor_move' | 'clear' | 'erase';
  x?: number;
  y?: number;
  lastX?: number;
  lastY?: number;
  width?: number;
  height?: number;
  color?: string;
  lineWidth?: number;
  tool?: DrawingTool;
  cursorX?: number;
  cursorY?: number;
  userId?: string;
}

export interface User {
  id: string;
  cursorX?: number;
  cursorY?: number;
  color: string;
}

export const Whiteboard: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [isDrawing, setIsDrawing] = useState(false);
  const [activeTool, setActiveTool] = useState<DrawingTool>('pen');
  const [activeColor, setActiveColor] = useState('#0f172a'); // draw-black
  const [lineWidth, setLineWidth] = useState(2);
  const [users, setUsers] = useState<User[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const lastCursorUpdate = useRef<number>(0);
  const startPos = useRef<{ x: number; y: number }>({ x: 0, y: 0 });

  // Initialize canvas
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Set canvas size
    const resizeCanvas = () => {
      const rect = canvas.getBoundingClientRect();
      canvas.width = rect.width * window.devicePixelRatio;
      canvas.height = rect.height * window.devicePixelRatio;
      ctx.scale(window.devicePixelRatio, window.devicePixelRatio);
      canvas.style.width = rect.width + 'px';
      canvas.style.height = rect.height + 'px';
    };

    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);

    // Set default canvas styles
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    return () => {
      window.removeEventListener('resize', resizeCanvas);
    };
  }, []);

  // WebSocket connection
  useEffect(() => {
    if (!roomId) return;

    const connectWebSocket = () => {
      const ws = new WebSocket(roomService.getWhiteBoardWsUrl(roomId));
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        toast.success('Connected to whiteboard');
      };

      ws.onclose = () => {
        setIsConnected(false);
        toast.error('Disconnected from whiteboard');
        // Attempt to reconnect after 3 seconds
        setTimeout(connectWebSocket, 3000);
      };

      ws.onerror = () => {
        toast.error('Connection error');
      };

      ws.onmessage = (event) => {
        try {
          const message: DrawMessage = JSON.parse(event.data);
          handleIncomingMessage(message);
        } catch (error) {
          console.error('Failed to parse message:', error);
        }
      };
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [roomId]);

  const sendMessage = useCallback((message: DrawMessage) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      // Transform the message to match backend format
      const structuredMessage = {
        type: message.type,
        data: {
          cursorX: message.cursorX || message.x,
          cursorY: message.cursorY || message.y,
          color: message.color || '#3b82f6',
          lineWidth: message.lineWidth,
          tool: message.tool,
          width: message.width,
          height: message.height
        }
      };
      
      wsRef.current.send(JSON.stringify(structuredMessage));
    }
  }, []);

  const handleIncomingMessage = useCallback((message: any) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Handle the backend response format
    const messageData = message.data || message;
    const userId = message.userId || message.id;

    switch (message.type) {
      case 'draw_start':
        if (messageData.cursorX !== undefined && messageData.cursorY !== undefined) {
          ctx.strokeStyle = messageData.color || '#000000';
          ctx.lineWidth = messageData.lineWidth || 2;
          ctx.beginPath();
          ctx.moveTo(messageData.cursorX, messageData.cursorY);
        }
        break;

      case 'draw_move':
        if (messageData.cursorX !== undefined && messageData.cursorY !== undefined) {
          ctx.lineTo(messageData.cursorX, messageData.cursorY);
          ctx.stroke();
        }
        break;

      case 'draw_end':
        ctx.closePath();
        break;

      case 'shape':
        if (messageData.cursorX !== undefined && messageData.cursorY !== undefined && 
            messageData.width !== undefined && messageData.height !== undefined) {
          ctx.strokeStyle = messageData.color || '#000000';
          ctx.lineWidth = messageData.lineWidth || 2;
          ctx.beginPath();
          
          if (messageData.tool === 'rectangle') {
            ctx.rect(messageData.cursorX, messageData.cursorY, messageData.width, messageData.height);
          } else if (messageData.tool === 'circle') {
            const centerX = messageData.cursorX + messageData.width / 2;
            const centerY = messageData.cursorY + messageData.height / 2;
            const radius = Math.min(Math.abs(messageData.width), Math.abs(messageData.height)) / 2;
            ctx.arc(centerX, centerY, radius, 0, 2 * Math.PI);
          }
          ctx.stroke();
        }
        break;

      case 'clear':
        ctx.fillStyle = '#ffffff';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        break;

      case 'erase':
        if (messageData.cursorX !== undefined && messageData.cursorY !== undefined && messageData.lineWidth) {
          ctx.globalCompositeOperation = 'destination-out';
          ctx.beginPath();
          ctx.arc(messageData.cursorX, messageData.cursorY, messageData.lineWidth / 2, 0, 2 * Math.PI);
          ctx.fill();
          ctx.globalCompositeOperation = 'source-over';
        }
        break;

      case 'cursor_move':
        if (userId && messageData.cursorX !== undefined && messageData.cursorY !== undefined) {
          setUsers(prev => {
            const existing = prev.find(u => u.id === userId);
            if (existing) {
              return prev.map(u => 
                u.id === userId 
                  ? { ...u, cursorX: messageData.cursorX, cursorY: messageData.cursorY, color: messageData.color }
                  : u
              );
            } else {
              return [...prev, {
                id: userId,
                cursorX: messageData.cursorX,
                cursorY: messageData.cursorY,
                color: messageData.color || '#3b82f6'
              }];
            }
          });
        }
        break;
    }
  }, []);

  const getCanvasCoordinates = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas) return { x: 0, y: 0 };

    const rect = canvas.getBoundingClientRect();
    return {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top
    };
  };

  const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const { x, y } = getCanvasCoordinates(e);
    setIsDrawing(true);

    if (activeTool === 'pen' || activeTool === 'eraser') {
      const canvas = canvasRef.current;
      if (!canvas) return;

      const ctx = canvas.getContext('2d');
      if (!ctx) return;

      if (activeTool === 'pen') {
        ctx.globalCompositeOperation = 'source-over';
        ctx.strokeStyle = activeColor;
        ctx.lineWidth = lineWidth;
        ctx.beginPath();
        ctx.moveTo(x, y);

        sendMessage({
          type: 'draw_start',
          x,
          y,
          color: activeColor,
          lineWidth,
          tool: activeTool
        });
      } else if (activeTool === 'eraser') {
        sendMessage({
          type: 'erase',
          x,
          y,
          lineWidth
        });
      }
    } else {
      // For shapes, just store the start position
      startPos.current = { x, y };
    }
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const { x, y } = getCanvasCoordinates(e);

    // Send cursor position (debounced)
    const now = Date.now();
    if (now - lastCursorUpdate.current > 50) {
      sendMessage({
        type: 'cursor_move',
        cursorX: x,
        cursorY: y
      });
      lastCursorUpdate.current = now;
    }

    if (!isDrawing) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    if (activeTool === 'pen') {
      ctx.lineTo(x, y);
      ctx.stroke();

      sendMessage({
        type: 'draw_move',
        x,
        y,
        color: activeColor,
        lineWidth,
        tool: activeTool
      });
    } else if (activeTool === 'eraser') {
      ctx.globalCompositeOperation = 'destination-out';
      ctx.beginPath();
      ctx.arc(x, y, lineWidth / 2, 0, 2 * Math.PI);
      ctx.fill();
      ctx.globalCompositeOperation = 'source-over';

      sendMessage({
        type: 'erase',
        x,
        y,
        lineWidth
      });
    }
  };

  const handleMouseUp = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDrawing) return;
    setIsDrawing(false);

    const { x, y } = getCanvasCoordinates(e);

    if (activeTool === 'pen') {
      sendMessage({ type: 'draw_end', x, y });
    } else if (activeTool === 'rectangle' || activeTool === 'circle') {
      const width = x - startPos.current.x;
      const height = y - startPos.current.y;

      const canvas = canvasRef.current;
      if (!canvas) return;

      const ctx = canvas.getContext('2d');
      if (!ctx) return;

      ctx.strokeStyle = activeColor;
      ctx.lineWidth = lineWidth;
      ctx.beginPath();

      if (activeTool === 'rectangle') {
        ctx.rect(startPos.current.x, startPos.current.y, width, height);
      } else if (activeTool === 'circle') {
        const centerX = startPos.current.x + width / 2;
        const centerY = startPos.current.y + height / 2;
        const radius = Math.min(Math.abs(width), Math.abs(height)) / 2;
        ctx.arc(centerX, centerY, radius, 0, 2 * Math.PI);
      }
      ctx.stroke();

      sendMessage({
        type: 'shape',
        x: startPos.current.x,
        y: startPos.current.y,
        width,
        height,
        color: activeColor,
        lineWidth,
        tool: activeTool
      });
    }
  };

  const handleClearCanvas = () => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    sendMessage({ type: 'clear' });
    toast.success('Canvas cleared');
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Header */}
      <header className="bg-toolbar-bg border-b border-toolbar-border px-4 py-3 shadow-soft">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-xl font-semibold text-foreground">Collaborative Whiteboard</h1>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">Room:</span>
              <code className="px-2 py-1 bg-muted rounded text-sm font-mono">{roomId}</code>
              <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex">
        {/* Sidebar */}
        <aside className="w-64 bg-toolbar-bg border-r border-toolbar-border p-4 shadow-soft">
          <div className="space-y-6">
            <Toolbar
              activeTool={activeTool}
              onToolChange={setActiveTool}
              onClear={handleClearCanvas}
            />
            
            <ColorPicker
              activeColor={activeColor}
              onColorChange={setActiveColor}
            />
            
            <LineWidthSlider
              lineWidth={lineWidth}
              onLineWidthChange={setLineWidth}
            />
            
            <UsersList users={users} />
          </div>
        </aside>

        {/* Canvas Area */}
        <main className="flex-1 p-4">
          <div className="w-full h-full bg-canvas-bg border border-canvas-border rounded-lg shadow-medium overflow-hidden">
            <canvas
              ref={canvasRef}
              className="w-full h-full cursor-crosshair"
              onMouseDown={handleMouseDown}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onMouseLeave={() => setIsDrawing(false)}
            />
          </div>
        </main>

        {/* User Cursors */}
        {users.map(user => (
          user.cursorX !== undefined && user.cursorY !== undefined && (
            <div
              key={user.id}
              className="absolute pointer-events-none z-50"
              style={{
                left: user.cursorX + 256 + 16, // Adjust for sidebar width + padding
                top: user.cursorY + 64 + 16, // Adjust for header height + padding
                transform: 'translate(-50%, -50%)'
              }}
            >
              <div 
                className="w-3 h-3 rounded-full border-2 border-white shadow-md"
                style={{ backgroundColor: user.color }}
              />
            </div>
          )
        ))}
      </div>
    </div>
  );
};