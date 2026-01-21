import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Users, User, Hash } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useChatStore } from '@/store/chatStore';
import type { Room } from '@/types/chat';

interface CreateRoomDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const roomTypes = [
  {
    value: 'group',
    label: 'Group',
    description: 'Private group for selected members',
    icon: Users,
  },
  {
    value: 'direct',
    label: 'Direct Message',
    description: 'One-on-one conversation',
    icon: User,
  },
  {
    value: 'channel',
    label: 'Channel',
    description: 'Public channel anyone can join',
    icon: Hash,
  },
] as const;

export function CreateRoomDialog({ open, onOpenChange }: CreateRoomDialogProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [type, setType] = useState<'group' | 'direct' | 'channel'>('group');
  const [isLoading, setIsLoading] = useState(false);

  const { setRooms, rooms, setCurrentRoom, currentUser, currentApplication } = useChatStore();

  const handleCreate = async () => {
    if (!name.trim() || !currentUser) return;

    setIsLoading(true);

    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 500));

    const newRoom: Room = {
      id: crypto.randomUUID(),
      application_id: currentApplication?.id || '',
      name: name.trim(),
      type,
      description: description.trim() || undefined,
      created_by: currentUser.id,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      unread_count: 0,
      members: [
        {
          id: crypto.randomUUID(),
          room_id: '',
          user_id: currentUser.id,
          role: 'admin',
          user: currentUser,
          joined_at: new Date().toISOString(),
        },
      ],
    };

    setRooms([newRoom, ...rooms]);
    setCurrentRoom(newRoom);
    setIsLoading(false);
    onOpenChange(false);

    // Reset form
    setName('');
    setDescription('');
    setType('group');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create Conversation</DialogTitle>
          <DialogDescription>
            Start a new group, direct message, or channel
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Room Type */}
          <div className="space-y-3">
            <Label>Conversation Type</Label>
            <RadioGroup
              value={type}
              onValueChange={(v) => setType(v as typeof type)}
              className="grid grid-cols-3 gap-3"
            >
              {roomTypes.map((roomType) => {
                const Icon = roomType.icon;
                return (
                  <div key={roomType.value}>
                    <RadioGroupItem
                      value={roomType.value}
                      id={roomType.value}
                      className="peer sr-only"
                    />
                    <Label
                      htmlFor={roomType.value}
                      className={cn(
                        'flex flex-col items-center gap-2 p-4 rounded-xl border-2 cursor-pointer transition-all',
                        'hover:bg-muted/50',
                        'peer-data-[state=checked]:border-primary peer-data-[state=checked]:bg-primary/5'
                      )}
                    >
                      <Icon className="h-6 w-6" />
                      <span className="font-medium text-sm">{roomType.label}</span>
                    </Label>
                  </div>
                );
              })}
            </RadioGroup>
          </div>

          {/* Name */}
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={type === 'channel' ? 'e.g., announcements' : 'e.g., Project Team'}
            />
          </div>

          {/* Description */}
          {type !== 'direct' && (
            <div className="space-y-2">
              <Label htmlFor="description">Description (optional)</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="What's this conversation about?"
                rows={3}
              />
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleCreate} disabled={!name.trim() || isLoading}>
            {isLoading ? 'Creating...' : 'Create'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
