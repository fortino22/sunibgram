'use client';

import { PostComposer } from './PostComposer';

export function CreatePostBox() {
  return (
    <PostComposer
      placeholder="What is happening?!"
      formId="create-post-box"
    />
  );
}
