-- Create databases
SELECT 'CREATE DATABASE chatapp'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'chatapp')\gexec

SELECT 'CREATE DATABASE chatdb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'chatdb')\gexec