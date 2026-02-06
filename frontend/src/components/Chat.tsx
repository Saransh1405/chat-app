import { useState, useEffect, useRef } from 'react';
import { User, Room, Message, MessageReaction, TypingIndicator, RoomMember } from '../types';
import { roomsAPI } from '../api/rooms';
import { messagesAPI } from '../api/messages';
import { reactionsAPI } from '../api/reactions';
import { typingAPI } from '../api/typing';
import { usersAPI } from '../api/users';
import { wsClient } from '../api/websocket';

interface ChatProps {
  user: User;
  onLogout: () => void;
}

export default function Chat({ user, onLogout }: ChatProps) {
  // Validate user object
  if (!user || !user.id) {
    console.error('Invalid user object:', user);
    return (
      <div style={{ padding: '20px', textAlign: 'center' }}>
        <p>Invalid user data. Please log in again.</p>
        <button onClick={onLogout}>Go to Login</button>
      </div>
    );
  }

  const [rooms, setRooms] = useState<Room[]>([]);
  const [selectedRoom, setSelectedRoom] = useState<Room | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [messageContent, setMessageContent] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [typingUsers, setTypingUsers] = useState<Set<string>>(new Set());
  const [reactions, setReactions] = useState<Map<string, MessageReaction[]>>(new Map());
  const [roomMembers, setRoomMembers] = useState<RoomMember[]>([]);
  const [isMember, setIsMember] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [checkingMembership, setCheckingMembership] = useState(false);
  const [showUserList, setShowUserList] = useState(false);
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [fetchingUsers, setFetchingUsers] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const typingTimeoutRef = useRef<Map<string, any>>(new Map());
  const selectedRoomRef = useRef<Room | null>(null);
  const messagesRef = useRef<Message[]>([]);

  // Keep refs in sync with state
  useEffect(() => {
    selectedRoomRef.current = selectedRoom;
  }, [selectedRoom]);

  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  useEffect(() => {
    let mounted = true;

    const initialize = async () => {
      try {
        await loadRooms();
        if (mounted) {
          await connectWebSocket();
        }
      } catch (error) {
        console.error('Initialization error:', error);
        setError('Failed to initialize. Please refresh the page.');
      }
    };

    initialize();

    return () => {
      mounted = false;
      // Cleanup WebSocket on unmount
      try {
        wsClient.disconnect();
      } catch (error) {
        console.error('Error disconnecting WebSocket:', error);
      }
    };
  }, []);

  useEffect(() => {
    if (selectedRoom) {
      checkMembership(selectedRoom.id);
      loadRoomMembers(selectedRoom.id);
    }
    return () => {
      if (selectedRoom) {
        wsClient.unsubscribe(selectedRoom.id);
      }
    };
  }, [selectedRoom]);

  useEffect(() => {
    if (selectedRoom && isMember) {
      console.log('User is member, loading messages and subscribing to room:', selectedRoom.id);
      loadMessages(selectedRoom.id).catch(console.error);

      // Subscribe to room for real-time updates
      const subscribeToRoom = () => {
        if (wsClient.isConnected()) {
          console.log('📡 Subscribing to selected room:', selectedRoom.id);
          wsClient.subscribe(selectedRoom.id);
          console.log('✅ Subscribe call completed for room:', selectedRoom.id);
          return true;
        }
        return false;
      };

      // Try to subscribe immediately
      if (!subscribeToRoom()) {
        console.warn('⚠️ WebSocket not connected, will retry subscription in 1 second...');
        // Retry subscription after a short delay if WebSocket is not connected
        const retryTimeout = setTimeout(() => {
          if (!subscribeToRoom()) {
            console.warn('⚠️ WebSocket still not connected after retry, will keep trying...');
            // Keep retrying every 2 seconds until connected
            const intervalId = setInterval(() => {
              if (subscribeToRoom()) {
                clearInterval(intervalId);
              }
            }, 2000);

            // Clean up interval after 30 seconds
            setTimeout(() => clearInterval(intervalId), 30000);
          }
        }, 1000);

        // Cleanup timeout on unmount
        return () => clearTimeout(retryTimeout);
      }
    } else if (selectedRoom && !isMember) {
      // Unsubscribe if user is not a member
      console.log('User is not a member, unsubscribing from room:', selectedRoom.id);
      wsClient.unsubscribe(selectedRoom.id);
    }
  }, [selectedRoom, isMember]);

  useEffect(() => {
    if (selectedRoom && messages.length > 0) {
      loadReactions().catch(console.error);
    }
  }, [selectedRoom, messages.length]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const connectWebSocket = async () => {
    console.log('🔌 connectWebSocket called');
    const token = localStorage.getItem('access_token');
    if (!token) {
      console.warn('⚠️ No access token found for WebSocket connection');
      return;
    }
    console.log('✅ Access token found, length:', token.length);

    try {
      console.log('🚀 Calling wsClient.connect...');
      await wsClient.connect(token);
      console.log('✅ WebSocket connection established successfully');

      // Message handlers - use refs to get current values
      const handleMessage = (payload: Message) => {
        console.log('🔵 WebSocket message received:', payload);
        const currentRoom = selectedRoomRef.current;
        if (currentRoom && payload.room_id === currentRoom.id) {
          console.log('✅ Adding message to current room:', payload.id);
          setMessages((prev) => {
            // Check if message already exists (avoid duplicates)
            const exists = prev.some(msg => msg.id === payload.id);
            if (exists) {
              console.log('⚠️ Message already exists, skipping:', payload.id);
              return prev;
            }
            // Add new message at the END (oldest first, newest at bottom)
            // This matches the UI display order
            return [...prev, payload];
          });
        } else {
          console.log('❌ Message for different room. Current:', currentRoom?.id, 'Message room:', payload.room_id);
        }
      };

      const handleMessageUpdated = (payload: Message) => {
        const currentRoom = selectedRoomRef.current;
        if (currentRoom && payload.room_id === currentRoom.id) {
          setMessages((prev) =>
            prev.map((msg) => (msg.id === payload.id ? payload : msg))
          );
        }
      };

      const handleReactionAdded = () => {
        const currentRoom = selectedRoomRef.current;
        const currentMessages = messagesRef.current;
        if (currentRoom && currentMessages.length > 0) {
          loadReactions().catch(console.error);
        }
      };

      const handleReactionRemoved = () => {
        const currentRoom = selectedRoomRef.current;
        const currentMessages = messagesRef.current;
        if (currentRoom && currentMessages.length > 0) {
          loadReactions().catch(console.error);
        }
      };

      const handleRoomCreated = () => {
        loadRooms().catch(console.error);
      };

      const handleRoomUpdated = (payload: Room) => {
        console.log('🔄 Room updated event received:', payload);
        loadRooms().catch(console.error);
        const currentRoom = selectedRoomRef.current;
        if (currentRoom && payload.id === currentRoom.id) {
          // Check if room was deleted
          if (payload.deleted_at) {
            console.log('🗑️ Room was deleted, clearing selection');
            setSelectedRoom(null);
            setMessages([]);
            setRoomMembers([]);
            setIsMember(false);
            setIsAdmin(false);
          } else {
            setSelectedRoom(payload);
          }
        }
      };

      const handleRoomMemberAdded = (payload: any) => {
        console.log('➕ Room member added event received:', payload);
        const currentRoom = selectedRoomRef.current;
        if (currentRoom && payload.room && payload.room.id === currentRoom.id) {
          // Reload room members if viewing the affected room
          loadRoomMembers(currentRoom.id).catch(console.error);
          // If the added member is the current user, update membership status
          if (payload.member && payload.member.user_id === user.id) {
            checkMembership(currentRoom.id).catch(console.error);
          }
        }
        // Always reload rooms list to reflect member count changes
        loadRooms().catch(console.error);
      };

      const handleRoomMemberRemoved = (payload: any) => {
        console.log('➖ Room member removed event received:', payload);
        const currentRoom = selectedRoomRef.current;
        if (currentRoom && payload.room && payload.room.id === currentRoom.id) {
          // Reload room members if viewing the affected room
          loadRoomMembers(currentRoom.id).catch(console.error);
          // If the removed member is the current user, update membership status
          if (payload.memberRemoved && payload.memberRemoved.id === user.id) {
            setIsMember(false);
            setIsAdmin(false);
            setMessages([]);
            // Unsubscribe from room if user was removed
            if (wsClient.isConnected()) {
              wsClient.unsubscribe(currentRoom.id);
            }
          }
        }
        // Always reload rooms list to reflect member count changes
        loadRooms().catch(console.error);
      };

      const handleWebSocketError = (payload: any) => {
        console.error('❌ WebSocket error received:', payload);
        const errorMessage = payload?.message || 'WebSocket error occurred';
        setError(errorMessage);
        // Show error toast or notification
        console.warn('⚠️ WebSocket Error:', errorMessage);
      };

      // Register event listeners
      console.log('👂 Registering WebSocket event listeners...');
      wsClient.on('message', handleMessage);
      wsClient.on('message_updated', handleMessageUpdated);
      wsClient.on('reaction_added', handleReactionAdded);
      wsClient.on('reaction_removed', handleReactionRemoved);
      wsClient.on('typing', handleTypingIncoming);
      wsClient.on('room_created', handleRoomCreated);
      wsClient.on('room_updated', handleRoomUpdated);
      wsClient.on('room_member_added', handleRoomMemberAdded);
      wsClient.on('room_member_removed', handleRoomMemberRemoved);
      wsClient.on('error', handleWebSocketError);
      console.log('✅ All WebSocket event listeners registered');
    } catch (error) {
      console.error('❌ WebSocket connection failed:', error);
      console.error('📊 Error details:', {
        message: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined,
      });
      // Don't crash the app if WebSocket fails
    }
  };

  const loadRooms = async () => {
    try {
      const response = await roomsAPI.list();
      const roomsList = response?.rooms;
      // Ensure it's always an array
      setRooms(Array.isArray(roomsList) ? roomsList : []);
    } catch (err: any) {
      console.error('Error loading rooms:', err);
      setError(err.response?.data?.error?.message || 'Failed to load rooms');
      setRooms([]); // Ensure rooms is always an array
    }
  };

  const loadMessages = async (roomId: string) => {
    try {
      const response = await messagesAPI.list(roomId);
      const messagesList = response?.messages;
      // Ensure it's always an array
      const validMessages = Array.isArray(messagesList) ? messagesList : [];
      setMessages([...validMessages].reverse());
    } catch (err: any) {
      console.error('Error loading messages:', err);
      setError(err.response?.data?.error?.message || 'Failed to load messages');
      setMessages([]); // Ensure messages is always an array
    }
  };

  const loadReactions = async () => {
    if (!selectedRoom || messages.length === 0) return;
    const newReactions = new Map<string, MessageReaction[]>();
    for (const message of messages) {
      try {
        const response = await reactionsAPI.list(message.id);
        if (response.reactions && response.reactions.length > 0) {
          newReactions.set(message.id, response.reactions);
        }
      } catch (err) {
        // Silently fail for reactions - not critical
        console.error('Failed to load reactions for message:', message.id);
      }
    }
    setReactions(newReactions);
  };

  const checkMembership = async (roomId: string) => {
    setCheckingMembership(true);
    try {
      const response = await roomsAPI.listMembers(roomId);
      const members = response.members || [];
      setRoomMembers(members);
      const userMember = members.find(member => member.user_id === user.id);
      const userIsMember = !!userMember;
      setIsMember(userIsMember);
      // Check if user is admin or creator
      setIsAdmin(userMember?.role === 'admin' || userMember?.role === 'owner' || false);
    } catch (err: any) {
      console.error('Error checking membership:', err);
      setIsMember(false);
      setIsAdmin(false);
      setRoomMembers([]);
    } finally {
      setCheckingMembership(false);
    }
  };

  const loadAllUsers = async () => {
    setFetchingUsers(true);
    try {
      const response = await usersAPI.list(user.application_id);
      setAllUsers(response.users || []);
    } catch (err) {
      console.error('Failed to load users:', err);
    } finally {
      setFetchingUsers(false);
    }
  };

  useEffect(() => {
    if (showUserList && isAdmin) {
      loadAllUsers();
    }
  }, [showUserList, isAdmin]);

  const handleInviteUser = async (userId: string) => {
    if (!selectedRoom) return;

    setLoading(true);
    try {
      await roomsAPI.addMember({
        room_id: selectedRoom.id,
        user_id: userId,
      });
      await loadRoomMembers(selectedRoom.id);
      await checkMembership(selectedRoom.id);
      setError(''); // Clear any previous errors
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to invite user');
    } finally {
      setLoading(false);
    }
  };

  const loadRoomMembers = async (roomId: string) => {
    try {
      const response = await roomsAPI.listMembers(roomId);
      setRoomMembers(response.members || []);
    } catch (err: any) {
      console.error('Error loading room members:', err);
      setRoomMembers([]);
    }
  };

  const handleJoinRoom = async () => {
    if (!selectedRoom) return;

    setLoading(true);
    try {
      await roomsAPI.addMember({
        room_id: selectedRoom.id,
        user_id: user.id,
      });
      await checkMembership(selectedRoom.id);
      await loadRoomMembers(selectedRoom.id);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to join room');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRoom = async () => {
    const name = prompt('Enter room name:');
    if (!name) return;

    setLoading(true);
    try {
      const response = await roomsAPI.create({
        name,
        type: 'group',
        created_by: user.id
      });

      // Automatically add creator as a member
      try {
        await roomsAPI.addMember({
          room_id: response.room.id,
          user_id: user.id,
          role: 'admin'
        });
      } catch (memberErr) {
        console.error('Failed to add creator as member:', memberErr);
        // Continue anyway - room was created
      }

      await loadRooms();
      // Select the newly created room
      setSelectedRoom(response.room);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to create room');
    } finally {
      setLoading(false);
    }
  };

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedRoom || !messageContent.trim()) return;

    setLoading(true);
    try {
      await messagesAPI.create({
        room_id: selectedRoom.id,
        content: messageContent,
      });
      setMessageContent('');
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to send message');
    } finally {
      setLoading(false);
    }
  };

  const handleTypingIncoming = (payload: TypingIndicator) => {
    console.log('⌨️ handleTypingIncoming received:', payload);
    const currentRoom = selectedRoomRef.current;
    if (!currentRoom) {
      console.log('⚠️ No room currently selected, ignoring typing event');
      return;
    }

    console.log('📍 Current room:', currentRoom.id, 'Payload room:', payload.room_id);
    console.log('👤 Current user:', user.id, 'Payload user:', payload.user_id);

    if (payload.room_id === currentRoom.id && payload.user_id !== user.id) {
      console.log('✅ Matches! Adding user to typing list:', payload.user_id);
      // Clear existing timeout for this user if any
      const existingTimeout = typingTimeoutRef.current.get(payload.user_id);
      if (existingTimeout) {
        clearTimeout(existingTimeout);
      }

      setTypingUsers((prev) => {
        const next = new Set(prev);
        next.add(payload.user_id);
        return next;
      });

      const timeout = setTimeout(() => {
        console.log('⏲️ Removing typing user after timeout:', payload.user_id);
        setTypingUsers((prev) => {
          const next = new Set(prev);
          next.delete(payload.user_id);
          return next;
        });
        typingTimeoutRef.current.delete(payload.user_id);
      }, 5000);

      typingTimeoutRef.current.set(payload.user_id, timeout);
    } else {
      console.log('❌ Does not match criteria for showing indicator');
    }
  };

  const lastTypingSentRef = useRef<number>(0);
  const handleTypingOutgoing = async () => {
    if (!selectedRoom) return;

    const now = Date.now();
    if (now - lastTypingSentRef.current < 3000) return;

    lastTypingSentRef.current = now;

    try {
      if (user.id === '00000000-0000-0000-0000-000000000000') return;
      await typingAPI.create({
        room_id: selectedRoom.id,
        user_id: user.id,
        application_id: user.application_id,
      });
    } catch (err) {
      console.error('Failed to send typing indicator:', err);
    }
  };

  const handleAddReaction = async (messageId: string, reaction: string) => {
    try {
      await reactionsAPI.create({ message_id: messageId, reaction });
      loadReactions();
    } catch (err: any) {
      console.error('Failed to add reaction:', err);
    }
  };

  const handleRemoveReaction = async (messageId: string, reaction: string) => {
    try {
      await reactionsAPI.delete({ message_id: messageId, reaction });
      loadReactions();
    } catch (err: any) {
      console.error('Failed to remove reaction:', err);
    }
  };

  return (
    <div style={{ display: 'flex', height: '100vh' }}>
      {/* Sidebar */}
      <div style={{
        width: '250px',
        borderRight: '1px solid #ddd',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: '#f9f9f9'
      }}>
        <div style={{ padding: '15px', borderBottom: '1px solid #ddd' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h3 style={{ margin: 0 }}>Chat App</h3>
            <button
              onClick={onLogout}
              style={{
                padding: '5px 10px',
                backgroundColor: '#dc3545',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '12px'
              }}
            >
              Logout
            </button>
          </div>
          <div style={{ marginTop: '10px', fontSize: '12px', color: '#666' }}>
            {user.username} {user.email && `(${user.email})`}
          </div>
        </div>

        <div style={{ padding: '10px', borderBottom: '1px solid #ddd' }}>
          <button
            onClick={handleCreateRoom}
            style={{
              width: '100%',
              padding: '8px',
              backgroundColor: '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer'
            }}
          >
            + Create Room
          </button>
        </div>

        <div style={{ flex: 1, overflowY: 'auto' }}>
          {Array.isArray(rooms) && rooms.length > 0 ? rooms.map((room) => (
            <div
              key={room.id}
              onClick={() => setSelectedRoom(room)}
              style={{
                padding: '15px',
                cursor: 'pointer',
                backgroundColor: selectedRoom?.id === room.id ? '#e3f2fd' : 'transparent',
                borderBottom: '1px solid #eee'
              }}
            >
              <div style={{ fontWeight: 'bold' }}>{room.name}</div>
              <div style={{ fontSize: '12px', color: '#666' }}>{room.type}</div>
            </div>
          )) : (
            <div style={{ padding: '20px', textAlign: 'center', color: '#666' }}>
              No rooms yet. Create one to get started!
            </div>
          )}
        </div>
      </div>

      {/* Main Chat Area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        {selectedRoom ? (
          <>
            {/* Chat Header */}
            <div style={{
              padding: '15px',
              borderBottom: '1px solid #ddd',
              backgroundColor: 'white'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                <div>
                  <h3 style={{ margin: 0 }}>{selectedRoom.name}</h3>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    {selectedRoom.type} • {roomMembers.length} member{roomMembers.length !== 1 ? 's' : ''}
                    {isAdmin && ' • You are admin'}
                  </div>
                </div>
                <div style={{ display: 'flex', gap: '10px' }}>
                  {isMember && isAdmin && (
                    <button
                      onClick={() => setShowUserList(!showUserList)}
                      style={{
                        padding: '8px 16px',
                        backgroundColor: '#6c757d',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer',
                        fontSize: '12px'
                      }}
                    >
                      {showUserList ? 'Hide' : 'Invite Users'}
                    </button>
                  )}
                  {!isMember && !checkingMembership && (
                    <button
                      onClick={handleJoinRoom}
                      disabled={loading}
                      style={{
                        padding: '8px 16px',
                        backgroundColor: '#28a745',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: loading ? 'not-allowed' : 'pointer',
                        fontWeight: 'bold'
                      }}
                    >
                      Join Room
                    </button>
                  )}
                  {checkingMembership && (
                    <div style={{ fontSize: '12px', color: '#666' }}>Checking...</div>
                  )}
                </div>
              </div>

              {/* Room Members List */}
              {isMember && roomMembers.length > 0 && (
                <div style={{
                  marginTop: '10px',
                  paddingTop: '10px',
                  borderTop: '1px solid #eee',
                  fontSize: '12px'
                }}>
                  <div style={{ fontWeight: 'bold', marginBottom: '5px' }}>Members:</div>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                    {roomMembers.map((member) => (
                      <div
                        key={member.id}
                        style={{
                          padding: '4px 8px',
                          backgroundColor: member.user_id === user.id ? '#e3f2fd' : '#f0f0f0',
                          borderRadius: '12px',
                          fontSize: '11px'
                        }}
                      >
                        {member.user_id === user.id ? 'You' : (member.username || `User ${member.user_id.slice(0, 8)}`)}
                        {member.role && member.role !== 'member' && ` (${member.role})`}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Invite Users Panel (for admins) */}
              {showUserList && isAdmin && (
                <div style={{
                  marginTop: '15px',
                  padding: '15px',
                  backgroundColor: '#f9f9f9',
                  borderRadius: '4px',
                  border: '1px solid #ddd'
                }}>
                  <div style={{ fontWeight: 'bold', marginBottom: '10px', fontSize: '14px' }}>
                    Invite Users to Room
                  </div>
                  <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
                    {fetchingUsers ? (
                      <div style={{ textAlign: 'center', padding: '10px', fontSize: '12px' }}>Loading users...</div>
                    ) : allUsers.length > 0 ? (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
                        {allUsers
                          .filter(u => !roomMembers.some(m => m.user_id === u.id))
                          .map(u => (
                            <div key={u.id} style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              alignItems: 'center',
                              padding: '8px',
                              backgroundColor: 'white',
                              borderRadius: '4px',
                              border: '1px solid #eee'
                            }}>
                              <div>
                                <div style={{ fontWeight: 'bold', fontSize: '12px' }}>{u.username}</div>
                                <div style={{ fontSize: '10px', color: '#999' }}>{u.id.slice(0, 8)}...</div>
                              </div>
                              <button
                                onClick={() => handleInviteUser(u.id)}
                                disabled={loading}
                                style={{
                                  padding: '5px 10px',
                                  backgroundColor: '#007bff',
                                  color: 'white',
                                  border: 'none',
                                  borderRadius: '4px',
                                  cursor: loading ? 'not-allowed' : 'pointer',
                                  fontSize: '11px'
                                }}
                              >
                                Invite
                              </button>
                            </div>
                          ))}
                      </div>
                    ) : (
                      <div style={{ textAlign: 'center', padding: '10px', fontSize: '12px', color: '#999' }}>No other users found.</div>
                    )}
                  </div>
                  <div style={{ marginTop: '15px', paddingTop: '10px', borderTop: '1px solid #eee' }}>
                    <div style={{ fontSize: '12px', fontWeight: 'bold', marginBottom: '5px' }}>Invite by ID:</div>
                    <div style={{ display: 'flex', gap: '10px' }}>
                      <input
                        type="text"
                        placeholder="Enter User ID to invite"
                        id="invite-user-id"
                        style={{
                          flex: 1,
                          padding: '8px',
                          border: '1px solid #ddd',
                          borderRadius: '4px',
                          fontSize: '12px'
                        }}
                      />
                      <button
                        onClick={async () => {
                          const input = document.getElementById('invite-user-id') as HTMLInputElement;
                          const userId = input?.value.trim();
                          if (userId) {
                            await handleInviteUser(userId);
                            input.value = '';
                          }
                        }}
                        disabled={loading}
                        style={{
                          padding: '8px 16px',
                          backgroundColor: '#6c757d',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: loading ? 'not-allowed' : 'pointer',
                          fontSize: '12px'
                        }}
                      >
                        Invite ID
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Messages */}
            <div style={{
              flex: 1,
              overflowY: 'auto',
              padding: '20px',
              backgroundColor: '#f5f5f5'
            }}>
              {!isMember ? (
                <div style={{
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'center',
                  alignItems: 'center',
                  height: '100%',
                  textAlign: 'center',
                  color: '#666'
                }}>
                  <div style={{ fontSize: '18px', marginBottom: '10px' }}>
                    You're not a member of this room
                  </div>
                  <div style={{ fontSize: '14px', marginBottom: '20px' }}>
                    Join the room to view and send messages
                  </div>
                  <button
                    onClick={handleJoinRoom}
                    disabled={loading}
                    style={{
                      padding: '10px 20px',
                      backgroundColor: '#28a745',
                      color: 'white',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: loading ? 'not-allowed' : 'pointer',
                      fontSize: '16px',
                      fontWeight: 'bold'
                    }}
                  >
                    {loading ? 'Joining...' : 'Join Room'}
                  </button>
                </div>
              ) : Array.isArray(messages) && messages.length > 0 ? messages.map((message) => {
                const messageReactions = reactions.get(message.id);
                const hasReactions = messageReactions && Array.isArray(messageReactions) && messageReactions.length > 0;

                return (
                  <div
                    key={message.id}
                    style={{
                      marginBottom: '15px',
                      padding: '10px',
                      backgroundColor: message.user_id === user.id ? '#007bff' : 'white',
                      color: message.user_id === user.id ? 'white' : 'black',
                      borderRadius: '8px',
                      maxWidth: '70%',
                      marginLeft: message.user_id === user.id ? 'auto' : '0',
                      marginRight: message.user_id === user.id ? '0' : 'auto'
                    }}
                  >
                    <div style={{ fontSize: '12px', opacity: 0.8, marginBottom: '5px' }}>
                      {(() => {
                        const member = roomMembers.find(m => m.user_id === message.user_id);
                        return message.user_id === user.id ? 'You' : (member?.username || `User ${message.user_id.slice(0, 8)}`);
                      })()}
                    </div>
                    <div>{message.content}</div>
                    <div style={{ fontSize: '10px', opacity: 0.7, marginTop: '5px' }}>
                      {new Date(message.created_at).toLocaleTimeString()}
                    </div>

                    {/* Reactions */}
                    {hasReactions && (
                      <div style={{ marginTop: '5px', display: 'flex', gap: '5px', flexWrap: 'wrap' }}>
                        {messageReactions!.map((reaction) => (
                          <button
                            key={reaction.id}
                            onClick={() => {
                              if (reaction.user_id === user.id) {
                                handleRemoveReaction(message.id, reaction.reaction);
                              } else {
                                handleAddReaction(message.id, reaction.reaction);
                              }
                            }}
                            style={{
                              padding: '2px 6px',
                              backgroundColor: reaction.user_id === user.id ? 'rgba(255,255,255,0.3)' : 'rgba(0,0,0,0.1)',
                              border: 'none',
                              borderRadius: '4px',
                              cursor: 'pointer',
                              fontSize: '12px'
                            }}
                          >
                            {reaction.reaction} {messageReactions!.filter(r => r.reaction === reaction.reaction).length}
                          </button>
                        ))}
                      </div>
                    )}

                    {/* Add Reaction Button */}
                    <div style={{ marginTop: '5px' }}>
                      <button
                        onClick={() => {
                          const reaction = prompt('Enter reaction (emoji or text, max 10 chars):');
                          if (reaction && reaction.length <= 10) {
                            handleAddReaction(message.id, reaction);
                          }
                        }}
                        style={{
                          padding: '2px 6px',
                          backgroundColor: 'transparent',
                          border: '1px solid currentColor',
                          borderRadius: '4px',
                          cursor: 'pointer',
                          fontSize: '10px',
                          opacity: 0.7
                        }}
                      >
                        + React
                      </button>
                    </div>
                  </div>
                );
              }) : (
                <div style={{ padding: '20px', textAlign: 'center', color: '#666' }}>
                  No messages yet. Start the conversation!
                </div>
              )}



              <div ref={messagesEndRef} />
            </div>

            {/* Typing Indicator */}
            {typingUsers.size > 0 && (
              <div style={{
                padding: '5px 20px',
                fontStyle: 'italic',
                color: '#666',
                fontSize: '12px',
                backgroundColor: 'rgba(255, 255, 255, 0.8)',
                borderTop: '1px solid #eee'
              }}>
                {Array.from(typingUsers).map((userId, index, array) => {
                  const member = roomMembers.find(m => m.user_id === userId);
                  const displayName = userId === user.id ? 'You' : (member?.username || `User ${userId.slice(0, 8)}`);

                  if (index === 0) return displayName;
                  if (index === array.length - 1) return ` and ${displayName}`;
                  return `, ${displayName}`;
                })}
                {typingUsers.size === 1 ? ' is typing...' : ' are typing...'}
              </div>
            )}

            {/* Message Input */}
            {isMember && (
              <form onSubmit={handleSendMessage} style={{
                padding: '15px',
                borderTop: '1px solid #ddd',
                backgroundColor: 'white',
                display: 'flex',
                gap: '10px'
              }}>
                <input
                  type="text"
                  value={messageContent}
                  onChange={(e) => {
                    setMessageContent(e.target.value);
                    handleTypingOutgoing();
                  }}
                  placeholder="Type a message..."
                  style={{
                    flex: 1,
                    padding: '10px',
                    border: '1px solid #ddd',
                    borderRadius: '4px'
                  }}
                />
                <button
                  type="submit"
                  disabled={loading || !messageContent.trim()}
                  style={{
                    padding: '10px 20px',
                    backgroundColor: '#007bff',
                    color: 'white',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: loading ? 'not-allowed' : 'pointer'
                  }}
                >
                  Send
                </button>
              </form>
            )}
          </>
        ) : (
          <div style={{
            flex: 1,
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            color: '#666'
          }}>
            Select a room to start chatting
          </div>
        )}

        {error && (
          <div style={{
            position: 'fixed',
            bottom: '20px',
            right: '20px',
            padding: '15px',
            backgroundColor: '#fee',
            color: '#c33',
            borderRadius: '4px',
            boxShadow: '0 2px 10px rgba(0,0,0,0.1)'
          }}>
            {error}
            <button
              onClick={() => setError('')}
              style={{
                marginLeft: '10px',
                padding: '5px 10px',
                backgroundColor: '#c33',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer'
              }}
            >
              ×
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

