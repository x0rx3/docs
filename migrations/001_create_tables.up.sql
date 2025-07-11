CREATE TABLE users (
    uuid UUID PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_login ON users(login);

CREATE TABLE sessions (
    uuid UUID PRIMARY KEY,
    user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    user_login VARCHAR(50) NOT NULL REFERENCES users(login) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_uuid ON sessions(user_uuid);
CREATE INDEX IF NOT EXISTS idx_sessions_user_login ON sessions(user_login);

CREATE TABLE documents (
    uuid      UUID PRIMARY KEY,
    name      VARCHAR(255) NOT NULL,
    mime      VARCHAR(100) NOT NULL,
    file      BOOLEAN NOT NULL DEFAULT FALSE,
    public    BOOLEAN NOT NULL DEFAULT FALSE,
    create_at TIMESTAMP NOT NULL,
    path TEXT NULL
);
CREATE TABLE document_grants (
    document_uuid UUID NOT NULL REFERENCES documents(uuid) ON DELETE CASCADE,
    user_login VARCHAR(50) NOT NULL REFERENCES users(login) ON DELETE CASCADE,
    UNIQUE (document_uuid, user_login)
);
CREATE INDEX IF NOT EXISTS idx_document_grants_document ON document_grants(document_uuid);
CREATE INDEX IF NOT EXISTS idx_document_grants_user ON document_grants(user_login);