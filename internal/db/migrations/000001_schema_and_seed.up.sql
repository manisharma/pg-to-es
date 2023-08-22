BEGIN;
-- Create tables
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    created_at TIMESTAMP
);
CREATE TABLE hashtags (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    created_at TIMESTAMP
);
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    slug VARCHAR,
    description TEXT,
    created_at TIMESTAMP
);
CREATE TABLE project_hashtags (
    hashtag_id INT REFERENCES hashtags(id) ON DELETE CASCADE,
    project_id INT REFERENCES projects(id) ON DELETE CASCADE,
    PRIMARY KEY (hashtag_id, project_id)
);
CREATE TABLE user_projects (
    project_id INT REFERENCES projects(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, user_id)
);
-- Seed the tables with random data
INSERT INTO users (name, created_at)
SELECT 'User ' || generate_series(1, 10),
    NOW() - (random() * interval '365 days');
INSERT INTO hashtags (name, created_at)
SELECT 'Hashtag ' || generate_series(1, 5),
    NOW() - (random() * interval '365 days');
INSERT INTO projects (name, slug, description, created_at)
SELECT 'Project ' || generate_series(1, 15),
    'project-' || generate_series(1, 15),
    'Description for Project ' || generate_series(1, 15),
    NOW() - (random() * interval '365 days');
INSERT INTO project_hashtags (hashtag_id, project_id)
SELECT (
        SELECT id
        FROM hashtags
        ORDER BY random()
        LIMIT 1
    ), (
        SELECT id
        FROM projects
        ORDER BY random()
        LIMIT 1
    )
FROM generate_series(1, 30);
INSERT INTO user_projects (project_id, user_id)
SELECT (
        SELECT id
        FROM projects
        ORDER BY random()
        LIMIT 1
    ), (
        SELECT id
        FROM users
        ORDER BY random()
        LIMIT 1
    )
FROM generate_series(1, 20);
COMMIT;