import { PostComposer } from '@/components/shared/PostComposer';

interface ThreadComposerProps {
  placeholder: string;
  parentId: number;
}

export function ThreadComposer({ placeholder, parentId }: ThreadComposerProps) {
  return <PostComposer placeholder={placeholder} parentId={parentId} />;
}
