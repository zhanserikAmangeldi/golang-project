CREATE TABLE message_reads (
                               message_id BIGINT REFERENCES messages(id) ON DELETE CASCADE,
                               user_id BIGINT NOT NULL,
                               read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                               PRIMARY KEY (message_id, user_id)
);