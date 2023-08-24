--*******
-- Remove functions, triggers & tables
--*******
BEGIN;
DROP TRIGGER IF EXISTS user_notify on users;
DROP TRIGGER IF EXISTS hashtag_notify on hashtags;
DROP TRIGGER IF EXISTS project_notify on projects;
DROP FUNCTION IF EXISTS notify_trigger;
DROP TABLE IF EXISTS project_hashtags;
DROP TABLE IF EXISTS user_projects;
DROP TABLE IF EXISTS hashtags;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
COMMIT;