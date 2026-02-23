import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { 
  ArrowLeft, 
  Plus, 
  Search, 
  Users, 
  Calendar, 
  ArrowRight, 
  ArrowLeft as ArrowLeftIcon,
  RefreshCw,
  Eye
} from 'lucide-react';
import { useRoomsList } from '@/hooks/useRoom';

export const RoomsList: React.FC = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');
  
  const { 
    rooms, 
    pagination, 
    isLoading, 
    error,
    loadRooms, 
    goToPage, 
    refreshRooms 
  } = useRoomsList();

  useEffect(() => {
    loadRooms();
  }, [loadRooms]);

  const handleSearch = () => {
    // Reset to first page when searching
    goToPage(1);
    // You could implement server-side search here
    // For now, we'll just filter client-side
  };

  const filteredRooms = rooms.filter(room => 
    room.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    room.id.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const joinRoom = (roomId: string) => {
    navigate(`/room/${roomId}`);
  };

  const createNewRoom = () => {
    navigate('/');
  };

  const formatDate = (timestamp: number) => {
    if (!timestamp) return 'Unknown';
    return new Date(timestamp * 1000).toLocaleDateString();
  };

  if (error) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle className="text-destructive">Error Loading Rooms</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Button onClick={() => refreshRooms()} className="w-full">
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
    <div className="min-h-screen bg-background p-4">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              onClick={() => navigate('/')}
              className="flex items-center gap-2"
            >
              <ArrowLeft className="w-4 h-4" />
              Back to Home
            </Button>
            
            <div>
              <h1 className="text-3xl font-bold">Available Rooms</h1>
              <p className="text-muted-foreground">
                Browse and join collaborative whiteboard sessions
              </p>
            </div>
          </div>

          <Button onClick={createNewRoom} className="flex items-center gap-2">
            <Plus className="w-4 h-4" />
            Create Room
          </Button>
        </div>

        {/* Search and Filters */}
        <Card className="mb-6">
          <CardContent className="pt-6">
            <div className="flex gap-4">
              <div className="flex-1">
                <Input
                  placeholder="Search rooms by name or ID..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                />
              </div>
              <Button onClick={handleSearch} variant="outline">
                <Search className="w-4 h-4 mr-2" />
                Search
              </Button>
              <Button onClick={refreshRooms} variant="outline" disabled={isLoading}>
                <RefreshCw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                Refresh
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Rooms Grid */}
        {isLoading ? (
          <div className="text-center py-12">
            <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-4" />
            <p>Loading rooms...</p>
          </div>
        ) : filteredRooms.length === 0 ? (
          <Card className="text-center py-12">
            <CardContent>
              <p className="text-muted-foreground mb-4">
                {searchTerm ? 'No rooms found matching your search.' : 'No rooms available.'}
              </p>
              <Button onClick={createNewRoom} variant="outline">
                <Plus className="w-4 h-4 mr-2" />
                Create First Room
              </Button>
            </CardContent>
          </Card>
        ) : (
          <>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
              {filteredRooms.map((room) => (
                <Card key={room.id} className="hover:shadow-lg transition-shadow">
                  <CardHeader>
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <CardTitle className="text-lg mb-2">{room.name}</CardTitle>
                        <CardDescription className="text-sm font-mono text-muted-foreground">
                          {room.id}
                        </CardDescription>
                      </div>
                      <Badge variant={room.status === 'active' ? 'default' : 'secondary'}>
                        {room.status}
                      </Badge>
                    </div>
                  </CardHeader>
                  
                  <CardContent>
                    <div className="space-y-3">
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <Users className="w-4 h-4" />
                        <span>{room.active_users} active users</span>
                      </div>
                      
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <Calendar className="w-4 h-4" />
                        <span>Created {formatDate(room.created_at)}</span>
                      </div>

                      <Button 
                        onClick={() => joinRoom(room.id)}
                        className="w-full flex items-center gap-2"
                      >
                        <Eye className="w-4 h-4" />
                        Join Room
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>

            {/* Pagination */}
            {pagination.total > pagination.pageSize && (
              <Card className="mb-6">
                <CardContent className="pt-6">
                  <div className="flex items-center justify-between">
                    <div className="text-sm text-muted-foreground">
                      Showing {((pagination.page - 1) * pagination.pageSize) + 1} to{' '}
                      {Math.min(pagination.page * pagination.pageSize, pagination.total)} of{' '}
                      {pagination.total} rooms
                    </div>
                    
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => goToPage(pagination.page - 1)}
                        disabled={pagination.page <= 1}
                      >
                        <ArrowLeftIcon className="w-4 h-4" />
                        Previous
                      </Button>
                      
                      <span className="text-sm">
                        Page {pagination.page} of {Math.ceil(pagination.total / pagination.pageSize)}
                      </span>
                      
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => goToPage(pagination.page + 1)}
                        disabled={!pagination.hasMore}
                      >
                        Next
                        <ArrowRight className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </>
        )}

        {/* API Info */}
        <Card className="mt-8">
          <CardHeader>
            <CardTitle className="text-sm">API Endpoints Used</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground space-y-1">
            <p><code>GET /api/v1/white-board/rooms</code> - List rooms with pagination</p>
            <p><code>GET /api/v1/white-board/rooms/:id</code> - Get room details</p>
            <p><code>GET /api/v1/white-board/ws/:roomId</code> - WebSocket for real-time canvas</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};
