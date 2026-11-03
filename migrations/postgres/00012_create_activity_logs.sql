-- +goose Up
-- Who changed what, one row per change, published by the usecases and saved in the background.
-- Home shows the latest few; rows are never edited.
CREATE TABLE activity_logs (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_id   BIGINT REFERENCES users(id),  -- NULL when nobody was logged in
    action     TEXT NOT NULL,                -- domain.ActivityAction, e.g. slide.created
    entity_id  BIGINT,                       -- the changed row; NULL for a setting
    summary    TEXT NOT NULL,                -- Indonesian, shown after the actor's name
    created_at TIMESTAMPTZ NOT NULL          -- when it happened
);

CREATE INDEX activity_logs_created_at_idx ON activity_logs (created_at);

-- +goose Down
DROP TABLE activity_logs;
