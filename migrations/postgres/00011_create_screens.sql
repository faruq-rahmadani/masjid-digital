-- +goose Up
-- TVs that opened /tv, one row per device, inserted the first time it connects. Only what an
-- admin sets lives here; last seen and the app version are kept in the cache (E9).
CREATE TABLE screens (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    device_id     TEXT NOT NULL UNIQUE,                            -- random, made by the TV and kept in its localStorage
    name          TEXT NOT NULL DEFAULT '',                        -- '' until an admin names it (shown as BARU)
    name_hint     TEXT NOT NULL DEFAULT '',                        -- from /tv?layar=…, suggested when naming
    user_agent    TEXT NOT NULL DEFAULT '',
    forced_mode   TEXT NOT NULL DEFAULT 'auto',                    -- auto, normal, jumat or dark
    forced_until  TIMESTAMPTZ,                                     -- the forced mode ends here (the mosque's next midnight)
    blocked_at    TIMESTAMPTZ,                                     -- set: hidden from the list and refused on connect
    first_seen_at TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE screens;
