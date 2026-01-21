import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { useChatStore } from '@/store/chatStore';
import { MessageSquare, ArrowRight, Loader2 } from 'lucide-react';
import type { User } from '@/types/chat';

export function LoginPage() {
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const { setCurrentUser, setToken, setRooms, setMessages } = useChatStore();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) return;

    setIsLoading(true);
    setError('');

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000));

      // Mock user data
      const mockUser: User = {
        id: crypto.randomUUID(),
        application_id: 'demo-app',
        username: email.split('@')[0],
        email: email,
        avatar_url: `https://api.dicebear.com/7.x/avataaars/svg?seed=${email}`,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        is_online: true,
      };

      // Mock token
      const mockToken = 'mock-jwt-token-' + crypto.randomUUID();

      // Set user and token
      setCurrentUser(mockUser);
      setToken(mockToken);

      // Load mock rooms
      const mockRooms = [
        {
          id: 'room-1',
          application_id: 'demo-app',
          name: 'General',
          type: 'channel' as const,
          description: 'General discussions',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          unread_count: 3,
          last_message: {
            id: 'msg-1',
            room_id: 'room-1',
            user_id: 'user-2',
            content: 'Hey everyone! Welcome to the channel 👋',
            message_type: 'text' as const,
            created_at: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
            updated_at: new Date().toISOString(),
            user: {
              id: 'user-2',
              application_id: 'demo-app',
              username: 'Sarah',
              avatar_url: 'https://api.dicebear.com/7.x/avataaars/svg?seed=sarah',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
          },
        },
        {
          id: 'room-2',
          application_id: 'demo-app',
          name: 'Design Team',
          type: 'group' as const,
          description: 'Design team discussions',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          unread_count: 0,
          last_message: {
            id: 'msg-2',
            room_id: 'room-2',
            user_id: 'user-3',
            content: 'The new mockups look great!',
            message_type: 'text' as const,
            created_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
            updated_at: new Date().toISOString(),
            user: {
              id: 'user-3',
              application_id: 'demo-app',
              username: 'Mike',
              avatar_url: 'https://api.dicebear.com/7.x/avataaars/svg?seed=mike',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
          },
        },
        {
          id: 'room-3',
          application_id: 'demo-app',
          name: 'John Doe',
          type: 'direct' as const,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          unread_count: 1,
          last_message: {
            id: 'msg-3',
            room_id: 'room-3',
            user_id: 'user-4',
            content: 'Can we discuss the project tomorrow?',
            message_type: 'text' as const,
            created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
            updated_at: new Date().toISOString(),
            user: {
              id: 'user-4',
              application_id: 'demo-app',
              username: 'John',
              avatar_url: 'https://api.dicebear.com/7.x/avataaars/svg?seed=john',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
          },
        },
      ];

      setRooms(mockRooms);

      // Load mock messages for first room
      setMessages('room-1', [
        {
          id: 'msg-welcome',
          room_id: 'room-1',
          user_id: 'system',
          content: 'Welcome to General! This is where the team hangs out.',
          message_type: 'system',
          created_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 'msg-1',
          room_id: 'room-1',
          user_id: 'user-2',
          content: 'Hey everyone! Welcome to the channel 👋',
          message_type: 'text',
          created_at: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
          updated_at: new Date().toISOString(),
          status: 'read',
          user: {
            id: 'user-2',
            application_id: 'demo-app',
            username: 'Sarah',
            avatar_url: 'https://api.dicebear.com/7.x/avataaars/svg?seed=sarah',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
          reactions: [
            {
              id: 'react-1',
              message_id: 'msg-1',
              user_id: 'user-3',
              reaction: '👋',
              created_at: new Date().toISOString(),
            },
            {
              id: 'react-2',
              message_id: 'msg-1',
              user_id: 'user-4',
              reaction: '👋',
              created_at: new Date().toISOString(),
            },
          ],
        },
        {
          id: 'msg-1b',
          room_id: 'room-1',
          user_id: 'user-3',
          content: 'Excited to be here! Looking forward to collaborating with everyone.',
          message_type: 'text',
          created_at: new Date(Date.now() - 4 * 60 * 1000).toISOString(),
          updated_at: new Date().toISOString(),
          status: 'read',
          user: {
            id: 'user-3',
            application_id: 'demo-app',
            username: 'Mike',
            avatar_url: 'https://api.dicebear.com/7.x/avataaars/svg?seed=mike',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
        },
      ]);
    } catch {
      setError('Failed to login. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="flex flex-col items-center mb-8">
          <div className="flex items-center justify-center h-16 w-16 rounded-2xl bg-primary mb-4">
            <MessageSquare className="h-8 w-8 text-primary-foreground" />
          </div>
          <h1 className="text-2xl font-bold">ChatSDK</h1>
          <p className="text-muted-foreground mt-1">Multi-tenant chat platform</p>
        </div>

        <Card className="border-border/50 bg-card/50 backdrop-blur-sm">
          <CardHeader className="text-center">
            <CardTitle>Welcome back</CardTitle>
            <CardDescription>
              Enter your email to sign in to your account
            </CardDescription>
          </CardHeader>
          <form onSubmit={handleLogin}>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="bg-background/50"
                />
              </div>
              {error && (
                <p className="text-sm text-destructive">{error}</p>
              )}
            </CardContent>
            <CardFooter>
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Signing in...
                  </>
                ) : (
                  <>
                    Continue
                    <ArrowRight className="h-4 w-4 ml-2" />
                  </>
                )}
              </Button>
            </CardFooter>
          </form>
        </Card>

        <p className="text-center text-sm text-muted-foreground mt-6">
          Don't have an account?{' '}
          <button className="text-primary hover:underline">Sign up</button>
        </p>
      </div>
    </div>
  );
}
