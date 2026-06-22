GRANT USAGE ON SCHEMA public TO stream_user;

GRANT
    SELECT,
    INSERT,
    UPDATE,
    DELETE
ON ALL TABLES IN SCHEMA public
TO stream_user;

ALTER DEFAULT PRIVILEGES
FOR ROLE stream_migrator
IN SCHEMA public
GRANT
    SELECT,
    INSERT,
    UPDATE,
    DELETE
ON TABLES
TO stream_user;

REVOKE
    SELECT,
    INSERT,
    UPDATE,
    DELETE
ON TABLE schema_migrations
FROM stream_user;