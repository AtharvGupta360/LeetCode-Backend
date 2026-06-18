-- 000001_init_schema.down.sql
-- Rollback: drop the users table

DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";
