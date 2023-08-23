BEGIN;
--*******
-- Create tables
--*******
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE hashtags (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    slug VARCHAR,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
--*******
-- Create Notification Function
--*******
CREATE OR REPLACE FUNCTION notify_trigger() RETURNS TRIGGER AS $$
DECLARE notification_json jsonb;
BEGIN IF (TG_OP != 'DELETE') THEN IF (TG_TABLE_NAME = 'users') THEN notification_json = json_build_object(
    'user_id',
    U.id,
    'user_name',
    U.name,
    'user_created_at',
    U.created_at,
    'project_id',
    P.id,
    'project_name',
    P.name,
    'project_slug',
    P.slug,
    'project_description',
    P.description,
    'project_created_at',
    P.created_at,
    'hashtag_id',
    H.id,
    'hashtag_name',
    H.name,
    'hashtag_created_at',
    H.created_at,
    'operation',
    TG_OP,
    'table',
    TG_TABLE_NAME
) AS json_result
FROM users U
    LEFT JOIN user_projects UP ON U.id = UP.user_id
    LEFT JOIN projects P ON UP.project_id = P.id
    LEFT JOIN project_hashtags PH ON P.id = PH.project_id
    LEFT JOIN hashtags H ON PH.hashtag_id = H.id
WHERE U.id = NEW.id
ORDER BY U.id,
    P.id,
    H.id;
ELSIF (TG_TABLE_NAME = 'projects') THEN notification_json = json_build_object(
    'user_id',
    U.id,
    'user_name',
    U.name,
    'user_created_at',
    U.created_at,
    'project_id',
    P.id,
    'project_name',
    P.name,
    'project_slug',
    P.slug,
    'project_description',
    P.description,
    'project_created_at',
    P.created_at,
    'hashtag_id',
    H.id,
    'hashtag_name',
    H.name,
    'hashtag_created_at',
    H.created_at,
    'operation',
    TG_OP,
    'table',
    TG_TABLE_NAME
) AS json_result
FROM users U
    LEFT JOIN user_projects UP ON U.id = UP.user_id
    LEFT JOIN projects P ON UP.project_id = P.id
    LEFT JOIN project_hashtags PH ON P.id = PH.project_id
    LEFT JOIN hashtags H ON PH.hashtag_id = H.id
WHERE P.id = NEW.id
ORDER BY U.id,
    P.id,
    H.id;
ELSIF (TG_TABLE_NAME = 'hashtags') THEN notification_json = json_build_object(
    'user_id',
    U.id,
    'user_name',
    U.name,
    'user_created_at',
    U.created_at,
    'project_id',
    P.id,
    'project_name',
    P.name,
    'project_slug',
    P.slug,
    'project_description',
    P.description,
    'project_created_at',
    P.created_at,
    'hashtag_id',
    H.id,
    'hashtag_name',
    H.name,
    'hashtag_created_at',
    H.created_at,
    'operation',
    TG_OP,
    'table',
    TG_TABLE_NAME
) AS json_result
FROM users U
    LEFT JOIN user_projects UP ON U.id = UP.user_id
    LEFT JOIN projects P ON UP.project_id = P.id
    LEFT JOIN project_hashtags PH ON P.id = PH.project_id
    LEFT JOIN hashtags H ON PH.hashtag_id = H.id
WHERE H.id = NEW.id
ORDER BY U.id,
    P.id,
    H.id;
ELSIF (TG_TABLE_NAME = 'project_hashtags') THEN notification_json = json_build_object(
    'user_id',
    U.id,
    'user_name',
    U.name,
    'user_created_at',
    U.created_at,
    'project_id',
    P.id,
    'project_name',
    P.name,
    'project_slug',
    P.slug,
    'project_description',
    P.description,
    'project_created_at',
    P.created_at,
    'hashtag_id',
    H.id,
    'hashtag_name',
    H.name,
    'hashtag_created_at',
    H.created_at,
    'operation',
    TG_OP,
    'table',
    TG_TABLE_NAME
) AS json_result
FROM users U
    LEFT JOIN user_projects UP ON U.id = UP.user_id
    LEFT JOIN projects P ON UP.project_id = P.id
    LEFT JOIN project_hashtags PH ON P.id = PH.project_id
    LEFT JOIN hashtags H ON PH.hashtag_id = H.id
WHERE PH.hashtag_id = NEW.hashtag_id
    AND PH.project_id = NEW.project_id
ORDER BY U.id,
    P.id,
    H.id;
ELSIF (TG_TABLE_NAME = 'user_projects') THEN notification_json = json_build_object(
    'user_id',
    U.id,
    'user_name',
    U.name,
    'user_created_at',
    U.created_at,
    'project_id',
    P.id,
    'project_name',
    P.name,
    'project_slug',
    P.slug,
    'project_description',
    P.description,
    'project_created_at',
    P.created_at,
    'hashtag_id',
    H.id,
    'hashtag_name',
    H.name,
    'hashtag_created_at',
    H.created_at,
    'operation',
    TG_OP,
    'table',
    TG_TABLE_NAME
) AS json_result
FROM users U
    LEFT JOIN user_projects UP ON U.id = UP.user_id
    LEFT JOIN projects P ON UP.project_id = P.id
    LEFT JOIN project_hashtags PH ON P.id = PH.project_id
    LEFT JOIN hashtags H ON PH.hashtag_id = H.id
WHERE UP.project_id = NEW.project_id
    AND UP.user_id = NEW.user_id
ORDER BY U.id,
    P.id,
    H.id;
END IF;
PERFORM pg_notify(
    'core_db_event',
    json_build_object(
        'operation',
        TG_OP,
        'table',
        TG_TABLE_NAME,
        'payload',
        notification_json
    )::text
);
ELSIF (TG_OP = 'DELETE') THEN PERFORM pg_notify(
    'core_db_event',
    json_build_object(
        'operation',
        TG_OP,
        'table',
        TG_TABLE_NAME,
        'payload',
        row_to_json(OLD)
    )::text
);
END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;
--*******
-- Create Triggers
--*******
CREATE TRIGGER user_notify
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON users FOR EACH ROW EXECUTE PROCEDURE notify_trigger(
        'id',
        'name',
        'created_at'
    );
CREATE TRIGGER hashtag_notify
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON hashtags FOR EACH ROW EXECUTE PROCEDURE notify_trigger(
        'id',
        'name',
        'created_at'
    );
CREATE TRIGGER project_notify
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON projects FOR EACH ROW EXECUTE PROCEDURE notify_trigger(
        'id',
        'name',
        'slug',
        'description',
        'created_at'
    );
CREATE TRIGGER project_hashtags_notify
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON project_hashtags FOR EACH ROW EXECUTE PROCEDURE notify_trigger('hashtag_id', 'project_id');
CREATE TRIGGER user_projects_notify
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON user_projects FOR EACH ROW EXECUTE PROCEDURE notify_trigger('project_id', 'user_id');
--*******
-- Seed the tables with random data
--*******
INSERT INTO users (name, created_at)
SELECT 'User ' || generate_series(1, 10),
    NOW() - (random() * interval '365 days');
INSERT INTO hashtags (name, created_at)
SELECT 'Hashtag' || generate_series(1, 10),
    NOW() - (random() * interval '365 days');
INSERT INTO projects (name, slug, description, created_at)
SELECT 'Project ' || generate_series(1, 10),
    'project-' || generate_series(1, 10),
    'Description for Project ' || generate_series(1, 10),
    NOW() - (random() * interval '365 days');
INSERT INTO project_hashtags (hashtag_id, project_id)
SELECT generate_series(1, 10),
    generate_series(1, 10) ON CONFLICT DO NOTHING;
INSERT INTO user_projects (project_id, user_id)
SELECT generate_series(1, 10),
    generate_series(1, 10) ON CONFLICT DO NOTHING;
COMMIT;