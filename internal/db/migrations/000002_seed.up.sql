--*******
-- Seed the tables with random data
--*******
BEGIN;
INSERT INTO users (name, created_at)
SELECT 'User ' || generate_series(1, 10),
    NOW() - (random() * interval '365 days');
INSERT INTO projects (name, slug, description, created_at)
SELECT 'Project ' || generate_series(1, 10),
    'project-' || generate_series(1, 10),
    'Description for Project ' || generate_series(1, 10),
    NOW() - (random() * interval '365 days');
INSERT INTO hashtags (name, created_at)
SELECT 'Hashtag' || generate_series(1, 10),
    NOW() - (random() * interval '365 days');
INSERT INTO project_hashtags (hashtag_id, project_id)
SELECT generate_series(1, 10),
    generate_series(1, 10) ON CONFLICT DO NOTHING;
INSERT INTO user_projects (project_id, user_id)
SELECT generate_series(1, 10),
    generate_series(1, 10) ON CONFLICT DO NOTHING;
COMMIT;