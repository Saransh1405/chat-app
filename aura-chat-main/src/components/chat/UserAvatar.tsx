import { cn } from "@/lib/utils";

const GRADIENTS = [
  "from-purple-500 to-pink-500",
  "from-blue-500 to-purple-500",
  "from-emerald-500 to-cyan-500",
  "from-orange-500 to-rose-500",
  "from-violet-500 to-indigo-500",
  "from-pink-500 to-rose-400",
  "from-cyan-500 to-blue-500",
  "from-amber-500 to-orange-500",
];

function getGradient(id: string): string {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash);
  }
  return GRADIENTS[Math.abs(hash) % GRADIENTS.length];
}

interface UserAvatarProps {
  userId: string;
  username: string;
  avatarUrl?: string | null;
  size?: "sm" | "md" | "lg";
  isOnline?: boolean;
  className?: string;
}

export function UserAvatar({ userId, username, avatarUrl, size = "md", isOnline, className }: UserAvatarProps) {
  const sizeClasses = {
    sm: "h-8 w-8 text-xs",
    md: "h-10 w-10 text-sm",
    lg: "h-14 w-14 text-lg",
  };

  const onlineDotSize = {
    sm: "h-2.5 w-2.5",
    md: "h-3 w-3",
    lg: "h-4 w-4",
  };

  const initials = username
    .split(/[._-]/)
    .map((p) => p[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);

  return (
    <div className={cn("relative flex-shrink-0", className)}>
      {avatarUrl ? (
        <img
          src={avatarUrl}
          alt={username}
          className={cn("rounded-full object-cover ring-2 ring-border", sizeClasses[size])}
        />
      ) : (
        <div
          className={cn(
            "rounded-full bg-gradient-to-br flex items-center justify-center font-semibold text-white shadow-lg",
            getGradient(userId),
            sizeClasses[size]
          )}
        >
          {initials}
        </div>
      )}
      {isOnline !== undefined && (
        <span
          className={cn(
            "absolute bottom-0 right-0 rounded-full border-2 border-background",
            onlineDotSize[size],
            isOnline ? "bg-online" : "bg-muted-foreground"
          )}
        />
      )}
    </div>
  );
}
