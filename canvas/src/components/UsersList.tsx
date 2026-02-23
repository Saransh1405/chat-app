import React from 'react';
import { User } from './Whiteboard';
import { Users } from 'lucide-react';

interface UsersListProps {
  users: User[];
}

export const UsersList: React.FC<UsersListProps> = ({ users }) => {
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <Users className="w-4 h-4 text-muted-foreground" />
        <h3 className="text-sm font-medium text-foreground">
          Connected Users ({users.length})
        </h3>
      </div>
      
      {users.length === 0 ? (
        <p className="text-xs text-muted-foreground">No other users connected</p>
      ) : (
        <div className="space-y-2">
          {users.map((user) => (
            <div
              key={user.id}
              className="flex items-center gap-2 p-2 bg-muted rounded-md"
            >
              <div
                className="w-3 h-3 rounded-full border border-white"
                style={{ backgroundColor: user.color }}
              />
              <span className="text-xs text-foreground font-mono">
                {user.id.slice(0, 8)}...
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};