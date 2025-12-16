CREATE TABLE conversations (
                               id BIGSERIAL PRIMARY KEY,
                               is_group BOOLEAN DEFAULT FALSE,
                               created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE conversations ADD COLUMN name VARCHAR(255);