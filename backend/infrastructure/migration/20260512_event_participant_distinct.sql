UPDATE events
SET participant_count = (
	SELECT COUNT(DISTINCT threads.user_id)::integer
	FROM thread_events
	JOIN threads ON threads.thread_id = thread_events.thread_id
	WHERE thread_events.event_id = events.id AND threads.type <> 'REPOST'
);
