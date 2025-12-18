CREATE TABLE message_reactions (
                                   id BIGSERIAL PRIMARY KEY,
                                   message_id BIGINT REFERENCES messages(id) ON DELETE CASCADE,
                                   user_id BIGINT NOT NULL,
                                   reaction VARCHAR(10) NOT NULL,
                                   created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                   UNIQUE(message_id, user_id, reaction)
);
