import { useChatStore } from '@/store/chatStore';
import { LoginPage } from '@/components/auth/LoginPage';
import { ChatLayout } from '@/components/chat/ChatLayout';

const Index = () => {
  const { isAuthenticated } = useChatStore();

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  return <ChatLayout />;
};

export default Index;
