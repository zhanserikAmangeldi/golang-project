CREATE TABLE IF NOT EXISTS admin_audit_log (
   id BIGSERIAL PRIMARY KEY,
   admin_id BIGINT NOT NULL REFERENCES users(id),
   action VARCHAR(50) NOT NULL,
   target_type VARCHAR(50) NOT NULL,
   target_id BIGINT,
   details JSONB,
   ip_address INET,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);