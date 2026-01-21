import { Check, CheckCheck } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ReadReceiptProps {
  status: 'sending' | 'sent' | 'delivered' | 'read';
  className?: string;
}

export function ReadReceipt({ status, className }: ReadReceiptProps) {
  if (status === 'sending') {
    return (
      <div className={cn('h-4 w-4 rounded-full border-2 border-muted-foreground/40 border-t-transparent animate-spin', className)} />
    );
  }

  if (status === 'sent') {
    return <Check className={cn('h-4 w-4 text-muted-foreground/60', className)} />;
  }

  if (status === 'delivered') {
    return <CheckCheck className={cn('h-4 w-4 text-muted-foreground/60', className)} />;
  }

  if (status === 'read') {
    return <CheckCheck className={cn('h-4 w-4 text-primary', className)} />;
  }

  return null;
}
