DO
$$
BEGIN
   IF NOT EXISTS (
       SELECT 1 FROM pg_database WHERE datname = 'chatapp'
   ) THEN
      CREATE DATABASE chatapp;
   END IF;
END
$$;

DO
$$
BEGIN
   IF NOT EXISTS (
       SELECT 1 FROM pg_database WHERE datname = 'chatdb'
   ) THEN
      CREATE DATABASE chatdb;
   END IF;
END
$$;