ALTER TABLE wakers RENAME TO agents;

ALTER TABLE tasks RENAME COLUMN waker_id TO agent_id;
ALTER TABLE schedules RENAME COLUMN waker_id TO agent_id;
ALTER TABLE webhooks RENAME COLUMN waker_id TO agent_id;
ALTER TABLE memory_entries RENAME COLUMN waker_id TO agent_id;
