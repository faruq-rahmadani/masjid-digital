-- +goose Up
-- Keuangan (E7): single-entry, cash-basis bookkeeping on two axes. An account says where the money
-- is (Tunai, a bank); a pocket ("kas" in the UI) says whose money it is and what it may be spent
-- on (Kas Umum, Zakat Mal, a terikat kas). Every transaction has both. Balances are never stored:
-- the usecase sums the transactions after the last book closing. Rules per pocket kind live in
-- internal/domain, not here.

CREATE TABLE accounts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,              -- Tunai, BSI Masjid
    kind        TEXT NOT NULL,                     -- cash or bank
    number      TEXT NOT NULL DEFAULT '',          -- bank account number; '' for cash
    archived_at TIMESTAMP,                         -- hidden from the forms; kept for history
    created_by  INTEGER REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL,
    updated_at  TIMESTAMP NOT NULL
);

CREATE TABLE pockets (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    kind          TEXT NOT NULL,                   -- umum, zakat_mal, zakat_fitrah, fidyah or terikat
    name          TEXT NOT NULL UNIQUE,
    purpose       TEXT NOT NULL DEFAULT '',        -- peruntukan, shown to donors and the bendahara
    target_amount INTEGER,                         -- terikat only; NULL = no target
    surplus_plan  TEXT,                            -- terikat only: to_umum or ask_donor
    show_on_tv    BOOLEAN NOT NULL DEFAULT FALSE,  -- its summary rotates in the TV's kas header
    status        TEXT NOT NULL DEFAULT 'active',  -- active, target_reached, done or archived
    created_by    INTEGER REFERENCES users(id),
    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP NOT NULL
);

-- one row per built-in kind; only terikat pockets are created by users
CREATE UNIQUE INDEX pockets_builtin_kind_idx ON pockets (kind) WHERE kind <> 'terikat';

CREATE TABLE categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,                     -- in or out, fixed once used
    scope       TEXT NOT NULL,                     -- umum_terikat, zakat, fidyah or all
    code        TEXT UNIQUE,                       -- system categories only (asnaf, pindah, ...)
    archived_at TIMESTAMP,
    created_by  INTEGER REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL,
    updated_at  TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX categories_name_idx ON categories (scope, type, name);

CREATE TABLE transactions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    kind          TEXT NOT NULL,                   -- in or out; a pindah dana is an out + in pair
    amount        INTEGER NOT NULL CHECK (amount > 0), -- integer Rupiah
    date          TEXT NOT NULL,                   -- YYYY-MM-DD, the mosque's date
    account_id    INTEGER NOT NULL REFERENCES accounts(id),
    pocket_id     INTEGER NOT NULL REFERENCES pockets(id),
    category_id   INTEGER NOT NULL REFERENCES categories(id),
    group_id      TEXT,                            -- links the rows of one split or one pindah dana
    donor_consent BOOLEAN NOT NULL DEFAULT FALSE,  -- pindah dana out of a terikat pocket
    corrects_id   INTEGER REFERENCES transactions(id), -- a correction of an entry in a closed month
    note          TEXT NOT NULL DEFAULT '',
    donor_note    TEXT NOT NULL DEFAULT '',        -- e.g. Hamba Allah, untuk renovasi atap
    created_by    INTEGER REFERENCES users(id),
    updated_by    INTEGER REFERENCES users(id),
    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP NOT NULL,
    deleted_at    TIMESTAMP,                       -- soft delete: money is never removed
    deleted_by    INTEGER REFERENCES users(id)
);

CREATE INDEX transactions_date_idx ON transactions (date);
CREATE INDEX transactions_pocket_date_idx ON transactions (pocket_id, date);
CREATE INDEX transactions_account_date_idx ON transactions (account_id, date);
CREATE INDEX transactions_group_idx ON transactions (group_id);

-- the receipt photo, apart so listing transactions never reads the bytes
CREATE TABLE transaction_photos (
    transaction_id INTEGER PRIMARY KEY REFERENCES transactions(id),
    content_type   TEXT NOT NULL,                  -- image/jpeg, image/png or image/webp
    data           BLOB NOT NULL                    -- at most 1 MB, resized in the browser
);

CREATE TABLE transaction_revisions (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    transaction_id INTEGER NOT NULL REFERENCES transactions(id),
    action         TEXT NOT NULL,                  -- edit or delete
    previous       TEXT NOT NULL,                  -- the row before the change, as JSON
    reason         TEXT NOT NULL,
    user_id        INTEGER REFERENCES users(id),
    created_at     TIMESTAMP NOT NULL
);

CREATE INDEX transaction_revisions_transaction_idx ON transaction_revisions (transaction_id);

-- A month's closing: its totals and each account's and pocket's balance at the end, which the
-- next balances start from. The opening (Mulai pembukuan) is a closing of the month before the
-- start, with the opening balances.
CREATE TABLE book_closings (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    month           TEXT NOT NULL,                 -- YYYY-MM
    kind            TEXT NOT NULL,                 -- opening or closing
    total_in        INTEGER NOT NULL DEFAULT 0,
    total_out       INTEGER NOT NULL DEFAULT 0,
    closing_balance INTEGER NOT NULL,
    closed_by       INTEGER REFERENCES users(id),
    closed_at       TIMESTAMP NOT NULL,
    reopened_at     TIMESTAMP,                     -- P1: a reopened closing stays for history
    reopen_reason   TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX book_closings_month_idx ON book_closings (month) WHERE reopened_at IS NULL;

CREATE TABLE book_closing_accounts (
    book_closing_id INTEGER NOT NULL REFERENCES book_closings(id),
    account_id      INTEGER NOT NULL REFERENCES accounts(id),
    closing_balance INTEGER NOT NULL,                  -- computed
    counted_balance INTEGER,                           -- entered by the bendahara; NULL for the opening
    PRIMARY KEY (book_closing_id, account_id)
);

CREATE TABLE book_closing_pockets (
    book_closing_id INTEGER NOT NULL REFERENCES book_closings(id),
    pocket_id       INTEGER NOT NULL REFERENCES pockets(id),
    closing_balance INTEGER NOT NULL,                  -- computed, carried into the next month
    PRIMARY KEY (book_closing_id, pocket_id)
);

-- Seeds. The built-in pockets and the system categories carry rules in internal/domain and are
-- never renamed or deleted. Tunai always exists; banks are added by the bendahara.
INSERT INTO accounts (name, kind, created_at, updated_at) VALUES ('Tunai', 'cash', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

INSERT INTO pockets (kind, name, purpose, show_on_tv, created_at, updated_at) VALUES
    ('umum', 'Kas Umum', 'Infaq tanpa tujuan khusus: kotak amal, infaq Jumat, transfer umum.', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('zakat_mal', 'Zakat Mal', 'Zakat mal dari muzakki, disalurkan ke 8 asnaf.', FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('zakat_fitrah', 'Zakat Fitrah', 'Zakat fitrah, disalurkan sebelum sholat Idul Fitri.', FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('fidyah', 'Fidyah', 'Fidyah puasa, disalurkan ke fakir dan miskin.', FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

INSERT INTO categories (name, type, scope, code, created_at, updated_at) VALUES
    ('Pindah dana', 'in', 'all', 'pindah_in', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Pindah dana', 'out', 'all', 'pindah_out', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Zakat mal', 'in', 'zakat', 'zakat_mal', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Zakat fitrah', 'in', 'zakat', 'zakat_fitrah', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Fakir', 'out', 'zakat', 'asnaf_fakir', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Miskin', 'out', 'zakat', 'asnaf_miskin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Amil', 'out', 'zakat', 'asnaf_amil', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Mualaf', 'out', 'zakat', 'asnaf_mualaf', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Riqab', 'out', 'zakat', 'asnaf_riqab', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Gharimin', 'out', 'zakat', 'asnaf_gharimin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Fisabilillah', 'out', 'zakat', 'asnaf_fisabilillah', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Ibnu sabil', 'out', 'zakat', 'asnaf_ibnu_sabil', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Fidyah', 'in', 'fidyah', 'fidyah', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Fakir', 'out', 'fidyah', 'fidyah_fakir', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Miskin', 'out', 'fidyah', 'fidyah_miskin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Kotak amal harian', 'in', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Infaq Jumat', 'in', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Infaq transfer / QRIS', 'in', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Donasi program', 'in', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Listrik', 'out', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Air', 'out', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Kebersihan', 'out', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Kegiatan', 'out', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Honor', 'out', 'umum_terikat', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- +goose Down
DROP TABLE book_closing_pockets;
DROP TABLE book_closing_accounts;
DROP TABLE book_closings;
DROP TABLE transaction_revisions;
DROP TABLE transaction_photos;
DROP TABLE transactions;
DROP TABLE categories;
DROP TABLE pockets;
DROP TABLE accounts;
