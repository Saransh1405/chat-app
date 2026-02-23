import React from 'react';
import { Button } from '@/components/ui/button';
import { Pen, Eraser, Square, Circle, Minus, Trash2 } from 'lucide-react';
import { DrawingTool } from './Whiteboard';

interface ToolbarProps {
  activeTool: DrawingTool;
  onToolChange: (tool: DrawingTool) => void;
  onClear: () => void;
}

export const Toolbar: React.FC<ToolbarProps> = ({
  activeTool,
  onToolChange,
  onClear
}) => {
  const tools = [
    { id: 'pen' as DrawingTool, icon: Pen, label: 'Pen' },
    { id: 'eraser' as DrawingTool, icon: Eraser, label: 'Eraser' },
    { id: 'rectangle' as DrawingTool, icon: Square, label: 'Rectangle' },
    { id: 'circle' as DrawingTool, icon: Circle, label: 'Circle' },
    { id: 'line' as DrawingTool, icon: Minus, label: 'Line' }
  ];

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium text-foreground mb-3">Drawing Tools</h3>
        <div className="grid grid-cols-2 gap-2">
          {tools.map(({ id, icon: Icon, label }) => (
            <Button
              key={id}
              variant={activeTool === id ? "default" : "outline"}
              size="sm"
              onClick={() => onToolChange(id)}
              className={`flex flex-col items-center gap-1 h-12 ${
                activeTool === id 
                  ? 'bg-tool-active-bg border-tool-active text-tool-active' 
                  : 'hover:bg-tool-hover border-tool-border'
              }`}
            >
              <Icon className="w-4 h-4" />
              <span className="text-xs">{label}</span>
            </Button>
          ))}
        </div>
      </div>

      <div className="pt-4 border-t border-border">
        <Button
          variant="destructive"
          size="sm"
          onClick={onClear}
          className="w-full flex items-center gap-2"
        >
          <Trash2 className="w-4 h-4" />
          Clear Canvas
        </Button>
      </div>
    </div>
  );
};