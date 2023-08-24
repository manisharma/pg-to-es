--*******
-- Truncate tables
--*******
BEGIN;
TRUNCATE user_projects;
TRUNCATE project_hashtags;
TRUNCATE hashtags;
TRUNCATE projects;
TRUNCATE users;
COMMIT;