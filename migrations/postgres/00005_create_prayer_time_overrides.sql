-- +goose Up
-- One prayer time replaced on one date, e.g. Jumat Dzuhur at 11:58. Settings that apply every
-- day (corrections, Imsak, weekly rules, Hijri offset) live in configurations under prayer_settings.
-- date and time are local to the mosque with no zone, so they are kept as text (sortable, same on
-- both engines).
CREATE TABLE prayer_time_overrides (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    date       TEXT NOT NULL,                -- local date of the mosque, YYYY-MM-DD
    prayer     TEXT NOT NULL,                -- subuh, dzuhur, ashar, maghrib, isya (checked in domain.Prayer)
    time       TEXT NOT NULL,                -- local time, HH:MM
    remarks    TEXT NOT NULL DEFAULT '',
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (date, prayer)                    -- one override per prayer per date
);

-- +goose Down
DROP TABLE prayer_time_overrides;
