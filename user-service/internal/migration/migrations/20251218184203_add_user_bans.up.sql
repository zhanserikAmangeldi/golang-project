CREATE TABLE IF NOT EXISTS user_bans (
     id BIGSERIAL PRIMARY KEY,
     user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     banned_by BIGINT NOT NULL REFERENCES users(id),
     reason TEXT NOT NULL,
     banned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
     expires_at TIMESTAMP WITH TIME ZONE,
     is_permanent BOOLEAN DEFAULT FALSE,
     unbanned_at TIMESTAMP WITH TIME ZONE,
     unbanned_by BIGINT REFERENCES users(id),

     CONSTRAINT ban_duration_check CHECK (
         (is_permanent = TRUE AND expires_at IS NULL) OR
         (is_permanent = FALSE AND expires_at IS NOT NULL)
    )
);