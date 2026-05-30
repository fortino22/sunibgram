'use client';

import { ReportModal } from '@/components/shared/ReportModal';

interface ReportPostModalProps {
  postId: number;
  onClose: () => void;
}

export function ReportPostModal({ postId, onClose }: ReportPostModalProps) {
  return <ReportModal target={{ type: 'post', postId }} onClose={onClose} />;
}
