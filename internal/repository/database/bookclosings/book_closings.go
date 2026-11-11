package bookclosings

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

const (
	closingsTable = "book_closings"
	accountsTable = "book_closing_accounts"
	pocketsTable  = "book_closing_pockets"
)

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

var columns = []string{
	"id", "month", "kind", "total_in", "total_out", "closing_balance", "closed_by", "closed_at", "reopened_at", "reopen_reason",
}

type model struct {
	ID             int64      `db:"id"`
	Month          string     `db:"month"`
	Kind           string     `db:"kind"`
	TotalIn        int64      `db:"total_in"`
	TotalOut       int64      `db:"total_out"`
	ClosingBalance int64      `db:"closing_balance"`
	ClosedBy       *int64     `db:"closed_by"`
	ClosedAt       time.Time  `db:"closed_at"`
	ReopenedAt     *time.Time `db:"reopened_at"`
	ReopenReason   string     `db:"reopen_reason"`
}

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// FindAll returns the closings not reopened, oldest month first, with their balances
func (r *repo) FindAll(ctx context.Context) ([]domain.BookClosing, error) {
	return r.find(ctx, 0)
}

// FindLatest returns the newest closing not reopened, or domain.ErrBookClosingNotFound before
// pembukuan starts
func (r *repo) FindLatest(ctx context.Context) (domain.BookClosing, error) {
	closings, err := r.find(ctx, 1)
	if err != nil {
		return domain.BookClosing{}, err
	}
	if len(closings) == 0 {
		return domain.BookClosing{}, apperror.New(apperror.CodeNotFound, domain.ErrBookClosingNotFound, "pembukuan belum dimulai")
	}
	return closings[0], nil
}

// find returns the active closings with their balances: the newest `limit` (newest first) when
// limit > 0, else all (oldest first)
func (r *repo) find(ctx context.Context, limit uint64) ([]domain.BookClosing, error) {
	q := builder.Select(columns...).From(closingsTable).Where(sq.Eq{"reopened_at": nil})
	if limit > 0 {
		q = q.OrderBy("month DESC").Limit(limit)
	} else {
		q = q.OrderBy("month")
	}
	query, args, err := q.ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find closings query: %w", err))
	}

	var models []model
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding closings: %w", err))
	}

	closings := make([]domain.BookClosing, 0, len(models))
	for _, m := range models {
		closing := domain.BookClosing{
			ID:             m.ID,
			Month:          m.Month,
			Kind:           domain.BookClosingKind(m.Kind),
			TotalIn:        m.TotalIn,
			TotalOut:       m.TotalOut,
			ClosingBalance: m.ClosingBalance,
			ClosedBy:       database.IDOrZero(m.ClosedBy),
			ClosedAt:       m.ClosedAt,
			ReopenedAt:     m.ReopenedAt,
			ReopenReason:   m.ReopenReason,
		}
		if closing.Accounts, err = r.accounts(ctx, m.ID); err != nil {
			return nil, err
		}
		if closing.Pockets, err = r.pockets(ctx, m.ID); err != nil {
			return nil, err
		}
		closings = append(closings, closing)
	}
	return closings, nil
}

func (r *repo) accounts(ctx context.Context, closingID int64) ([]domain.BookClosingAccount, error) {
	query, args, err := builder.
		Select("account_id", "closing_balance", "counted_balance").
		From(accountsTable).
		Where(sq.Eq{"book_closing_id": closingID}).
		OrderBy("account_id").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building closing accounts query: %w", err))
	}
	var models []struct {
		AccountID      int64  `db:"account_id"`
		ClosingBalance int64  `db:"closing_balance"`
		CountedBalance *int64 `db:"counted_balance"`
	}
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding closing accounts: %w", err))
	}
	balances := make([]domain.BookClosingAccount, 0, len(models))
	for _, m := range models {
		balances = append(balances, domain.BookClosingAccount{AccountID: m.AccountID, ClosingBalance: m.ClosingBalance, CountedBalance: m.CountedBalance})
	}
	return balances, nil
}

func (r *repo) pockets(ctx context.Context, closingID int64) ([]domain.BookClosingPocket, error) {
	query, args, err := builder.
		Select("pocket_id", "closing_balance").
		From(pocketsTable).
		Where(sq.Eq{"book_closing_id": closingID}).
		OrderBy("pocket_id").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building closing pockets query: %w", err))
	}
	var models []struct {
		PocketID       int64 `db:"pocket_id"`
		ClosingBalance int64 `db:"closing_balance"`
	}
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding closing pockets: %w", err))
	}
	balances := make([]domain.BookClosingPocket, 0, len(models))
	for _, m := range models {
		balances = append(balances, domain.BookClosingPocket{PocketID: m.PocketID, ClosingBalance: m.ClosingBalance})
	}
	return balances, nil
}

// Create inserts a closing with its balances. The usecase runs it in a transaction, so a closing
// never exists without its balances. A month already closed is CodeAlreadyExists.
func (r *repo) Create(ctx context.Context, c domain.BookClosing) (domain.BookClosing, error) {
	query, args, err := builder.
		Insert(closingsTable).
		Columns("month", "kind", "total_in", "total_out", "closing_balance", "closed_by", "closed_at").
		Values(c.Month, string(c.Kind), c.TotalIn, c.TotalOut, c.ClosingBalance, database.NullableID(c.ClosedBy), time.Now().UTC()).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return domain.BookClosing{}, errInternal(fmt.Errorf("error building create closing query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return domain.BookClosing{}, apperror.New(apperror.CodeAlreadyExists, err, "bulan ini sudah tutup buku")
		}
		return domain.BookClosing{}, errInternal(fmt.Errorf("error creating closing: %w", err))
	}

	for _, a := range c.Accounts {
		q, args, err := builder.Insert(accountsTable).
			Columns("book_closing_id", "account_id", "closing_balance", "counted_balance").
			Values(id, a.AccountID, a.ClosingBalance, a.CountedBalance).ToSql()
		if err != nil {
			return domain.BookClosing{}, errInternal(fmt.Errorf("error building closing account query: %w", err))
		}
		if _, err := r.client.Exec(ctx, q, args...); err != nil {
			return domain.BookClosing{}, errInternal(fmt.Errorf("error saving closing account: %w", err))
		}
	}
	for _, p := range c.Pockets {
		q, args, err := builder.Insert(pocketsTable).
			Columns("book_closing_id", "pocket_id", "closing_balance").
			Values(id, p.PocketID, p.ClosingBalance).ToSql()
		if err != nil {
			return domain.BookClosing{}, errInternal(fmt.Errorf("error building closing pocket query: %w", err))
		}
		if _, err := r.client.Exec(ctx, q, args...); err != nil {
			return domain.BookClosing{}, errInternal(fmt.Errorf("error saving closing pocket: %w", err))
		}
	}
	return r.FindLatest(ctx)
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data tutup buku")
}
