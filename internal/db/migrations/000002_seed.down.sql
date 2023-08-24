--*******
-- Truncate tables
--*******
BEGIN;
TRUNCATE users;
TRUNCATE projects;
TRUNCATE hashtags;
TRUNCATE project_hashtags;
TRUNCATE user_projects;
COMMIT;