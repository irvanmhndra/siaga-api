-- Store only SHA-256 digests of session tokens, so a leaked user_sessions
-- table (backup, replica, SQL injection) cannot be replayed as live sessions.
-- Existing rows are hashed in place and the API hashes each presented token
-- the same way (hex SHA-256 of the token text), so nobody is signed out.
ALTER TABLE user_sessions RENAME COLUMN token TO token_hash;
ALTER TABLE user_sessions RENAME COLUMN refresh_token TO refresh_token_hash;

UPDATE user_sessions SET
    token_hash         = encode(sha256(convert_to(token_hash, 'UTF8')), 'hex'),
    refresh_token_hash = encode(sha256(convert_to(refresh_token_hash, 'UTF8')), 'hex');

ALTER TABLE user_sessions RENAME CONSTRAINT user_sessions_token_key TO user_sessions_token_hash_key;
ALTER TABLE user_sessions RENAME CONSTRAINT user_sessions_refresh_token_key TO user_sessions_refresh_token_hash_key;
ALTER INDEX idx_user_sessions_token RENAME TO idx_user_sessions_token_hash;
ALTER INDEX idx_user_sessions_refresh_token RENAME TO idx_user_sessions_refresh_token_hash;
