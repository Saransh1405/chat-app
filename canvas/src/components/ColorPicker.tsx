import React from 'react';
import { Button } from '@/components/ui/button';
import { Palette } from 'lucide-react';

interface ColorPickerProps {
  activeColor: string;
  onColorChange: (color: string) => void;
}

export const ColorPicker: React.FC<ColorPickerProps> = ({
  activeColor,
  onColorChange
}) => {
  const colors = [
    { name: 'Light Gray', value: '#cbd5e1', cssVar: 'draw-light-gray' },
    { name: 'Medium Gray', value: '#94a3b8', cssVar: 'draw-medium-gray' },
    { name: 'Dark Gray', value: '#64748b', cssVar: 'draw-dark-gray' },
    { name: 'Darker Gray', value: '#475569', cssVar: 'draw-darker-gray' },
    { name: 'Charcoal', value: '#334155', cssVar: 'draw-charcoal' },
    { name: 'Black', value: '#0f172a', cssVar: 'draw-black' }
  ];

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <Palette className="w-4 h-4 text-muted-foreground" />
        <h3 className="text-sm font-medium text-foreground">Colors</h3>
      </div>
      
      <div className="grid grid-cols-3 gap-2">
        {colors.map(({ name, value }) => (
          <Button
            key={value}
            variant="outline"
            size="sm"
            onClick={() => onColorChange(value)}
            className={`p-2 h-10 border-2 hover:scale-105 transition-transform ${
              activeColor === value 
                ? 'border-tool-active ring-2 ring-tool-active ring-opacity-20' 
                : 'border-tool-border hover:border-tool-active'
            }`}
            style={{ backgroundColor: value }}
            title={name}
          >
            <span className="sr-only">{name}</span>
          </Button>
        ))}
      </div>
      
      <div className="flex items-center gap-2">
        <label 
          htmlFor="custom-color" 
          className="text-xs text-muted-foreground cursor-pointer"
        >
          Custom:
        </label>
        <input
          id="custom-color"
          type="color"
          value={activeColor}
          onChange={(e) => onColorChange(e.target.value)}
          className="w-8 h-8 rounded border border-tool-border cursor-pointer"
        />
      </div>
    </div>
  );
};