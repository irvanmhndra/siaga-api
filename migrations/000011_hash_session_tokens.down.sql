-- Digests cannot be turned back into tokens: drop every session so users
-- sign in again after the rollback.
DELETE FROM user_sessions;

ALTER INDEX idx_user_sessions_refresh_token_hash RENAME TO idx_user_sessions_refresh_token;
ALTER INDEX idx_user_sessions_token_hash RENAME TO idx_user_sessions_token;
ALTER TABLE user_sessions RENAME CONSTRAINT user_sessions_refresh_token_hash_key TO user_sessions_refresh_token_key;
ALTER TABLE user_sessions RENAME CONSTRAINT user_sessions_token_hash_key TO user_sessions_token_key;

ALTER TABLE user_sessions RENAME COLUMN refresh_token_hash TO refresh_token;
ALTER TABLE user_sessions RENAME COLUMN token_hash TO token;
