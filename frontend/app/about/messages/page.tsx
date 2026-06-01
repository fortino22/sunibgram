import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';
import {
  MessagesDemoDMs,
  MessagesDemoGroups,
  MessagesDemoMedia,
  MessagesDemoReadReceipts,
} from '@/components/about/demos/MessagesDemos';

export const metadata: Metadata = {
  title: 'Messages — Sunibgram',
  description:
    'Real-time private messaging with verified Binus students — DMs and group rooms.',
};

export default function MessagesAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="DMs · Real-time"
      title="Private chats with classmates."
      description="Real-time DMs and group rooms — verified, private, and built for campus collaboration. No strangers, no impostors."
      sections={[
        {
          id: 'one-on-one',
          title: 'One-on-one DMs',
          description:
            'Talk to anyone with a verified Binus identity. No friend requests, no follower gates — just open the chat and start typing.',
          bullets: [
            'Real-time delivery over Socket.io',
            'Search any verified student to start a DM',
            'Block or report from inside the chat',
          ],
          demo: <MessagesDemoDMs />,
        },
        {
          id: 'group-rooms',
          title: 'Group rooms',
          description:
            'Spin up a project chat with classmates in seconds. Add members, rename the room, share a description, and keep collaborating.',
          bullets: [
            'Up to dozens of members per room',
            'Owner can rename, describe, and remove members',
            'Live typing indicators per member',
          ],
          demo: <MessagesDemoGroups />,
        },
        {
          id: 'media-files',
          title: 'Media & files',
          description:
            'Share notes, slides, screenshots, and images instantly. Files stay inside the room — and the people in it.',
          bullets: [
            'Drop multiple attachments per message',
            'Image previews render inline',
            'Files indexed for in-chat search',
          ],
          demo: <MessagesDemoMedia />,
        },
        {
          id: 'read-receipts',
          title: 'Read receipts',
          description:
            'Know when your message has been seen. Edit a typo or unsend a message entirely if you change your mind.',
          bullets: [
            'Sending · Sent · Delivered · Seen states',
            'Edit any message you have sent',
            'Unsend removes it everywhere',
          ],
          demo: <MessagesDemoReadReceipts />,
        },
      ]}
    />
  );
}
