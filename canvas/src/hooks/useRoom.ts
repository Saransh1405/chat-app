import { useState, useEffect, useCallback } from 'react';
import { toast } from 'sonner';
import roomService, { Room, RoomStats } from '@/services/roomService';

export const useRoom = (roomId?: string) => {
  const [room, setRoom] = useState<Room | null>(null);
  const [stats, setStats] = useState<RoomStats | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load room data
  const loadRoom = useCallback(async (id: string) => {
    setIsLoading(true);
    setError(null);
    
    try {
      const [roomData] = await Promise.all([
        roomService.getRoomInfo(id),
      ]);
      
      setRoom(roomData);
      setStats(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load room data';
      setError(message);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Refresh room data
  const refreshRoom = useCallback(async () => {
    if (roomId) {
      await loadRoom(roomId);
      toast.success('Room information refreshed');
    }
  }, [roomId, loadRoom]);

  // Load room data when roomId changes
  useEffect(() => {
    if (roomId) {
      loadRoom(roomId);
    }
  }, [roomId, loadRoom]);

  return {
    room,
    stats,
    isLoading,
    error,
    loadRoom,
    refreshRoom,
    updateRoom: null,
    deleteRoom: null,
  };
};

export const useRoomsList = () => {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 20,
    total: 0,
    hasMore: false
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load rooms list
  const loadRooms = useCallback(async (page: number = 1, size: number = 20) => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await roomService.listRooms(page, size);
      setRooms(response.rooms);
      setPagination({
        page: response.page,
        pageSize: response.page_size,
        total: response.total,
        hasMore: response.has_more
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load rooms';
      setError(message);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Go to specific page
  const goToPage = useCallback((page: number) => {
    if (page >= 1 && page <= Math.ceil(pagination.total / pagination.pageSize)) {
      setPagination(prev => ({ ...prev, page }));
    }
  }, [pagination.total, pagination.pageSize]);

  // Refresh rooms list
  const refreshRooms = useCallback(async () => {
    await loadRooms(pagination.page, pagination.pageSize);
    toast.success('Rooms list refreshed');
  }, [pagination.page, pagination.pageSize, loadRooms]);

  return {
    rooms,
    pagination,
    isLoading,
    error,
    loadRooms,
    goToPage,
    refreshRooms,
  };
};
