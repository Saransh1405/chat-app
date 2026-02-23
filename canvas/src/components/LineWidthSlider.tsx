import React from 'react';
import { Slider } from '@/components/ui/slider';
import { PenTool } from 'lucide-react';

interface LineWidthSliderProps {
  lineWidth: number;
  onLineWidthChange: (width: number) => void;
}

export const LineWidthSlider: React.FC<LineWidthSliderProps> = ({
  lineWidth,
  onLineWidthChange
}) => {
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <PenTool className="w-4 h-4 text-muted-foreground" />
        <h3 className="text-sm font-medium text-foreground">
          Line Width ({lineWidth}px)
        </h3>
      </div>
      
      <div className="space-y-3">
        <Slider
          value={[lineWidth]}
          onValueChange={(value) => onLineWidthChange(value[0])}
          max={20}
          min={1}
          step={1}
          className="w-full"
        />
        
        {/* Visual preview */}
        <div className="flex justify-center py-2">
          <div
            className="bg-foreground rounded-full"
            style={{
              width: `${lineWidth}px`,
              height: `${lineWidth}px`,
              minWidth: '2px',
              minHeight: '2px'
            }}
          />
        </div>
      </div>
    </div>
  );
};