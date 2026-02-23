import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { PenTool, Users, Palette, Share2, Plus, Search, AlertCircle } from 'lucide-react';
import { toast } from 'sonner';
import roomService, { CreateRoomRequest } from '@/services/roomService';

const Index = () => {
  const navigate = useNavigate();
  const [roomId, setRoomId] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createForm, setCreateForm] = useState<CreateRoomRequest>({
    name: '',
    description: '',
    is_private: false,
  });

  const generateRoomId = () => {
    const id = roomService.generateRoomId();
    setRoomId(id);
  };

  const validateAndJoinRoom = async () => {
    if (!roomId.trim()) {
      toast.error('Please enter a room ID');
      return;
    }

    if (!roomService.isValidRoomId(roomId.trim())) {
      toast.error('Invalid room ID format. Use only letters, numbers, hyphens, and underscores.');
      return;
    }

    setIsLoading(true);
    try {
      // Check if room exists
      const exists = await roomService.roomExists(roomId.trim());
      if (exists) {
        navigate(`/room/${roomId.trim()}`);
        toast.success('Joining room...');
      } else {
        toast.error('Room not found. Please check the room ID or create a new room.');
      }
    } catch (error) {
      toast.error('Failed to check room. Please try again.');
      console.error('Error checking room:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const createNewRoom = async () => {
    if (!createForm.name.trim()) {
      toast.error('Please enter a room name');
      return;
    }

    setIsLoading(true);
    try {
      // Generate a room ID first
      const roomId = roomService.generateRoomId();
      
      // Create WebSocket connection to establish the room (white-board endpoint)
      const ws = new WebSocket(roomService.getWhiteBoardWsUrl(roomId));
      
      ws.onopen = () => {
        // WebSocket connected successfully - room is being created
        toast.success(`Room "${createForm.name}" created successfully!`);
        
        // Close the WebSocket connection after room creation
        ws.close();
        
        // Navigate to the whiteboard immediately
        navigate(`/room/${roomId}`);
      };
      
      ws.onerror = () => {
        toast.error('Failed to create room. Please try again.');
        setIsLoading(false);
      };
      
      ws.onclose = () => {
        // WebSocket closed, room creation complete
        setIsLoading(false);
        setShowCreateForm(false);
        setCreateForm({ name: '', description: '', is_private: false });
      };
      
      // Set a timeout in case WebSocket doesn't connect
      setTimeout(() => {
        if (ws.readyState === WebSocket.CONNECTING) {
          ws.close();
          toast.error('Room creation timed out. Please try again.');
          setIsLoading(false);
        }
      }, 5000);
      
    } catch (error) {
      toast.error('Failed to create room. Please try again.');
      console.error('Error creating room:', error);
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      if (showCreateForm) {
        createNewRoom();
      } else {
        validateAndJoinRoom();
      }
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-4xl">
        {/* Hero Section */}
        <div className="text-center mb-12">
          <div className="flex items-center justify-center mb-6">
            <div className="p-4 bg-tool-active-bg rounded-xl">
              <PenTool className="w-12 h-12 text-tool-active" />
            </div>
          </div>
          <h1 className="text-4xl md:text-5xl font-bold text-foreground mb-4">
            Collaborative Whiteboard
          </h1>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            Draw, sketch, and collaborate in real-time with your team. 
            Perfect for brainstorming, teaching, and creative sessions.
          </p>
        </div>

        {/* Features Grid */}
        <div className="grid md:grid-cols-3 gap-6 mb-12">
          <Card className="text-center">
            <CardHeader>
              <div className="mx-auto p-3 bg-draw-blue/10 rounded-lg w-fit mb-2">
                <PenTool className="w-6 h-6 text-draw-blue" />
              </div>
              <CardTitle className="text-lg">Rich Drawing Tools</CardTitle>
            </CardHeader>
            <CardContent>
              <CardDescription>
                Pen, eraser, shapes, and more with customizable colors and line widths
              </CardDescription>
            </CardContent>
          </Card>

          <Card className="text-center">
            <CardHeader>
              <div className="mx-auto p-3 bg-draw-green/10 rounded-lg w-fit mb-2">
                <Users className="w-6 h-6 text-draw-green" />
              </div>
              <CardTitle className="text-lg">Real-time Collaboration</CardTitle>
            </CardHeader>
            <CardContent>
              <CardDescription>
                See cursors and drawings from other users instantly as they work
              </CardDescription>
            </CardContent>
          </Card>

          <Card className="text-center">
            <CardHeader>
              <div className="mx-auto p-3 bg-draw-purple/10 rounded-lg w-fit mb-2">
                <Share2 className="w-6 h-6 text-draw-purple" />
              </div>
              <CardTitle className="text-lg">Easy Sharing</CardTitle>
            </CardHeader>
            <CardContent>
              <CardDescription>
                Share room codes to invite others and start collaborating immediately
              </CardDescription>
            </CardContent>
          </Card>
        </div>

        {/* Room Management Section */}
        <div className="grid md:grid-cols-2 gap-6 mb-12">
          {/* Join Room Card */}
          <Card>
            <CardHeader className="text-center">
              <div className="flex items-center justify-center mb-2">
                <Search className="w-5 h-5 mr-2" />
                <CardTitle>Join a Room</CardTitle>
              </div>
              <CardDescription>
                Enter a room ID to join an existing collaborative session
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <Input
                  placeholder="Enter room ID..."
                  value={roomId}
                  onChange={(e) => setRoomId(e.target.value)}
                  onKeyPress={handleKeyPress}
                  className="flex-1"
                  disabled={isLoading}
                />
                <Button
                  variant="outline"
                  onClick={generateRoomId}
                  className="shrink-0"
                  disabled={isLoading}
                >
                  Generate
                </Button>
              </div>
              
              <Button 
                onClick={validateAndJoinRoom}
                disabled={!roomId.trim() || isLoading}
                className="w-full"
                size="lg"
              >
                {isLoading ? 'Checking...' : (roomId.trim() ? 'Join Room' : 'Enter Room ID')}
              </Button>
            </CardContent>
          </Card>

          {/* Create Room Card */}
          <Card>
            <CardHeader className="text-center">
              <div className="flex items-center justify-center mb-2">
                <Plus className="w-5 h-5 mr-2" />
                <CardTitle>Create a Room</CardTitle>
              </div>
              <CardDescription>
                Start a new collaborative whiteboard session
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {!showCreateForm ? (
                <Button 
                  onClick={() => setShowCreateForm(true)}
                  className="w-full"
                  size="lg"
                  variant="outline"
                  disabled={isLoading}
                >
                  Create New Room
                </Button>
              ) : (
                <div className="space-y-3">
                  <Input
                    placeholder="Room name..."
                    value={createForm.name}
                    onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
                    onKeyPress={handleKeyPress}
                    disabled={isLoading}
                  />
                  <Input
                    placeholder="Description (optional)..."
                    value={createForm.description}
                    onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
                    onKeyPress={handleKeyPress}
                    disabled={isLoading}
                  />
                  <div className="flex gap-2">
                    <Button 
                      onClick={createNewRoom}
                      disabled={!createForm.name.trim() || isLoading}
                      className="flex-1"
                      size="lg"
                    >
                      {isLoading ? 'Creating...' : 'Create Room'}
                    </Button>
                    <Button 
                      onClick={() => setShowCreateForm(false)}
                      variant="outline"
                      size="lg"
                      disabled={isLoading}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Browse Rooms Section */}
        <Card className="max-w-2xl mx-auto mb-8">
          <CardHeader className="text-center">
            <CardTitle className="flex items-center justify-center">
              <Search className="w-5 h-5 mr-2" />
              Browse Available Rooms
            </CardTitle>
            <CardDescription>
              See all available collaborative sessions and join existing ones
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center">
            <Button 
              onClick={() => navigate('/rooms')}
              variant="outline"
              size="lg"
              className="flex items-center gap-2 mx-auto"
            >
              <Search className="w-4 h-4" />
              Browse All Rooms
            </Button>
          </CardContent>
        </Card>

        {/* API Status Info */}
        <Card className="max-w-2xl mx-auto mb-8">
          <CardHeader className="text-center">
            <CardTitle className="flex items-center justify-center">
              <AlertCircle className="w-5 h-5 mr-2 text-blue-500" />
              API Integration
            </CardTitle>
            <CardDescription>
              This app now integrates with the backend room management API
            </CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground space-y-2">
            <p><strong>Available Endpoints:</strong></p>
            <ul className="list-disc list-inside space-y-1 ml-4">
              <li><code>GET /api/v1/white-board/rooms</code> - List all rooms</li>
              <li><code>GET /api/v1/white-board/rooms/:id</code> - Get room info</li>
              <li><code>GET /api/v1/white-board/ws/:roomId</code> - WebSocket for canvas</li>
              <li><code>GET /api/v1/white-board/room/:roomId</code> - Room HTML page</li>
            </ul>
          </CardContent>
        </Card>

        {/* Footer */}
        <div className="text-center mt-12 text-sm text-muted-foreground">
          <p>Built with React, WebSocket, and HTML5 Canvas</p>
        </div>
      </div>
    </div>
  );
};

export default Index;
