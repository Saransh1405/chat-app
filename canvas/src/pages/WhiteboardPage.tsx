import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Whiteboard } from '@/components/Whiteboard';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Users, Settings, Trash2, ArrowLeft, Info, BarChart3 } from 'lucide-react';
import { useRoom } from '@/hooks/useRoom';

export const WhiteboardPage: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const [showRoomInfo, setShowRoomInfo] = useState(false);
  
  const { 
    room: roomInfo, 
    stats: roomStats, 
    isLoading, 
    error,
    refreshRoom, 
    updateRoom, 
    deleteRoom 
  } = useRoom(roomId);

  const handleDeleteRoom = async () => {
    const success = await deleteRoom();
    if (success) {
      navigate('/');
    }
  };

  const handleUpdateRoom = async () => {
    if (!roomInfo) return;
    
    try {
      await updateRoom({
        name: roomInfo.name,
        description: roomInfo.description || '',
        is_private: false
      });
    } catch (error) {
      // Error is already handled by the hook
    }
  };

  if (!roomId) {
    return <div>Room ID not found</div>;
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle className="text-destructive">Error Loading Room</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Button onClick={() => refreshRoom()} className="w-full">
              Retry
            </Button>
            <Button onClick={() => navigate('/')} variant="outline" className="w-full">
              Back to Home
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Room Header */}
      <div className="bg-card border-b p-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate('/')}
              className="flex items-center gap-2"
            >
              <ArrowLeft className="w-4 h-4" />
              Back to Home
            </Button>
            
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-semibold">
                {roomInfo?.name || `Room ${roomId}`}
              </h1>
              <Badge variant={roomInfo?.status === 'active' ? 'default' : 'secondary'}>
                {roomInfo?.status || 'Unknown'}
              </Badge>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowRoomInfo(!showRoomInfo)}
              className="flex items-center gap-2"
            >
              <Info className="w-4 h-4" />
              Room Info
            </Button>
            
            <Button
              variant="outline"
              size="sm"
              onClick={refreshRoom}
              disabled={isLoading}
              className="flex items-center gap-2"
            >
              <BarChart3 className="w-4 h-4" />
              Refresh
            </Button>
            
            <Button
              variant="outline"
              size="sm"
              onClick={handleUpdateRoom}
              disabled={isLoading}
              className="flex items-center gap-2"
            >
              <Settings className="w-4 h-4" />
              Update
            </Button>
            
            <Button
              variant="destructive"
              size="sm"
              onClick={handleDeleteRoom}
              disabled={isLoading}
              className="flex items-center gap-2"
            >
              <Trash2 className="w-4 h-4" />
              Delete
            </Button>
          </div>
        </div>
      </div>

      {/* Room Information Panel */}
      {showRoomInfo && (
        <div className="bg-muted/50 border-b p-4">
          <div className="max-w-7xl mx-auto">
            <div className="grid md:grid-cols-3 gap-4">
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm flex items-center gap-2">
                    <Users className="w-4 h-4" />
                    Active Users
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold">{roomInfo?.active_users || 0}</p>
                  <p className="text-xs text-muted-foreground">
                    Currently in the room
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm flex items-center gap-2">
                    <BarChart3 className="w-4 h-4" />
                    Room Statistics
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold">{roomStats?.total_users || 0}</p>
                  <p className="text-xs text-muted-foreground">
                    Total users joined
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm flex items-center gap-2">
                    <Info className="w-4 h-4" />
                    Room Details
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm font-medium">{roomInfo?.id}</p>
                  <p className="text-xs text-muted-foreground">
                    Room ID
                  </p>
                </CardContent>
              </Card>
            </div>

            {roomInfo?.users && roomInfo.users.length > 0 && (
              <Card className="mt-4">
                <CardHeader>
                  <CardTitle className="text-sm">Active Users</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex flex-wrap gap-2">
                    {roomInfo.users.map((user) => (
                      <Badge key={user.id} variant={user.is_online ? 'default' : 'secondary'}>
                        {user.username || user.id}
                        {user.is_online && <span className="ml-1">●</span>}
                      </Badge>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      )}

      {/* Main Whiteboard */}
      <div className="flex-1">
        <Whiteboard />
      </div>
    </div>
  );
};