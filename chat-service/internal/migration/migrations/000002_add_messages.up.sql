CREATE TABLE messages (
                          id BIGSERIAL PRIMARY KEY,
                          conversation_id BIGINT REFERENCES conversations(id) ON DELETE CASCADE,
                          sender_id BIGINT NOT NULL,
                          content TEXT NOT NULL,
                          file_url VARCHAR(500),
                          file_name VARCHAR(255),
                          file_size BIGINT,
                          mime_type VARCHAR(100),
                          message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file', 'audio', 'video')),
                          deleted_at TIMESTAMP WITH TIME ZONE,
                          edited_at TIMESTAMP WITH TIME ZONE,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_messages_conv_id ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
