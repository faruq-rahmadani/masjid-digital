package transactions

import (
	"context"
	"encoding/json"
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
	transactionsTable = "transactions"
	photosTable       = "transaction_photos"
	revisionsTable    = "transaction_revisions"
)

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

// columns selects a transaction with whether it has a photo (LEFT JOIN transaction_photos p)
var columns = []string{
	"t.id", "t.kind", "t.amount", "t.date", "t.account_id", "t.pocket_id", "t.category_id", "t.group_id",
	"t.donor_consent", "t.corrects_id", "t.note", "t.donor_note", "p.transaction_id IS NOT NULL AS has_photo",
	"t.created_by", "t.updated_by", "t.created_at", "t.updated_at", "t.deleted_at", "t.deleted_by",
}

type model struct {
	ID           int64      `db:"id"`
	Kind         string     `db:"kind"`
	Amount       int64      `db:"amount"`
	Date         string     `db:"date"`
	AccountID    int64      `db:"account_id"`
	PocketID     int64      `db:"pocket_id"`
	CategoryID   int64      `db:"category_id"`
	GroupID      *string    `db:"group_id"`
	DonorConsent bool       `db:"donor_consent"`
	CorrectsID   *int64     `db:"corrects_id"`
	Note         string     `db:"note"`
	DonorNote    string     `db:"donor_note"`
	HasPhoto     bool       `db:"has_photo"`
	CreatedBy    *int64     `db:"created_by"`
	UpdatedBy    *int64     `db:"updated_by"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
	DeletedBy    *int64     `db:"deleted_by"`
}

func (m model) toDomain() domain.Transaction {
	return domain.Transaction{
		ID:           m.ID,
		Kind:         domain.CategoryType(m.Kind),
		Amount:       m.Amount,
		Date:         m.Date,
		AccountID:    m.AccountID,
		PocketID:     m.PocketID,
		CategoryID:   m.CategoryID,
		GroupID:      m.GroupID,
		DonorConsent: m.DonorConsent,
		CorrectsID:   m.CorrectsID,
		Note:         m.Note,
		DonorNote:    m.DonorNote,
		HasPhoto:     m.HasPhoto,
		CreatedBy:    database.IDOrZero(m.CreatedBy),
		UpdatedBy:    database.IDOrZero(m.UpdatedBy),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    m.DeletedAt,
		DeletedBy:    database.IDOrZero(m.DeletedBy),
	}
}

func selectQuery() sq.SelectBuilder {
	return builder.
		Select(columns...).
		From(transactionsTable + " t").
		LeftJoin(photosTable + " p ON p.transaction_id = t.id")
}

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// Create inserts a transaction and returns it. Created and updated are the same moment, so a new
// transaction is never "Diedit".
func (r *repo) Create(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	now := time.Now().UTC()
	query, args, err := builder.
		Insert(transactionsTable).
		Columns("kind", "amount", "date", "account_id", "pocket_id", "category_id", "group_id", "donor_consent",
			"corrects_id", "note", "donor_note", "created_by", "updated_by", "created_at", "updated_at").
		Values(string(tx.Kind), tx.Amount, tx.Date, tx.AccountID, tx.PocketID, tx.CategoryID, tx.GroupID, tx.DonorConsent,
					tx.CorrectsID, tx.Note, tx.DonorNote, database.NullableID(tx.CreatedBy), database.NullableID(tx.CreatedBy), now, now).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return domain.Transaction{}, errInternal(fmt.Errorf("error building create transaction query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		return domain.Transaction{}, errInternal(fmt.Errorf("error creating transaction: %w", err))
	}
	return r.FindByID(ctx, id)
}

// FindByID returns one transaction, deleted or not, or domain.ErrTransactionNotFound
func (r *repo) FindByID(ctx context.Context, id int64) (domain.Transaction, error) {
	query, args, err := selectQuery().Where(sq.Eq{"t.id": id}).ToSql()
	if err != nil {
		return domain.Transaction{}, errInternal(fmt.Errorf("error building find transaction query: %w", err))
	}

	var m model
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.Transaction{}, errNotFound()
		}
		return domain.Transaction{}, errInternal(fmt.Errorf("error finding transaction: %w", err))
	}
	return m.toDomain(), nil
}

// Find returns the transactions matching the filter, by date then ID
func (r *repo) Find(ctx context.Context, f domain.TransactionFilter) ([]domain.Transaction, error) {
	where := sq.And{}
	if f.From != "" {
		// dates are zero-padded text, so comparing them as strings orders them by date
		where = append(where, sq.GtOrEq{"t.date": f.From})
	}
	if f.To != "" {
		where = append(where, sq.LtOrEq{"t.date": f.To})
	}
	if f.PocketID != 0 {
		where = append(where, sq.Eq{"t.pocket_id": f.PocketID})
	}
	if f.AccountID != 0 {
		where = append(where, sq.Eq{"t.account_id": f.AccountID})
	}
	if f.GroupID != "" {
		where = append(where, sq.Eq{"t.group_id": f.GroupID})
	}
	if !f.IncludeDeleted {
		where = append(where, sq.Eq{"t.deleted_at": nil})
	}

	query, args, err := selectQuery().Where(where).OrderBy("t.date", "t.id").ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find transactions query: %w", err))
	}

	var models []model
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding transactions: %w", err))
	}

	txs := make([]domain.Transaction, 0, len(models))
	for _, m := range models {
		txs = append(txs, m.toDomain())
	}
	return txs, nil
}

// Update saves the editable fields and who changed it, or returns domain.ErrTransactionNotFound
func (r *repo) Update(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	query, args, err := builder.
		Update(transactionsTable).
		Set("amount", tx.Amount).
		Set("date", tx.Date).
		Set("account_id", tx.AccountID).
		Set("pocket_id", tx.PocketID).
		Set("category_id", tx.CategoryID).
		Set("donor_consent", tx.DonorConsent).
		Set("note", tx.Note).
		Set("donor_note", tx.DonorNote).
		Set("updated_by", database.NullableID(tx.UpdatedBy)).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": tx.ID}).
		ToSql()
	if err != nil {
		return domain.Transaction{}, errInternal(fmt.Errorf("error building update transaction query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return domain.Transaction{}, errInternal(fmt.Errorf("error updating transaction: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return domain.Transaction{}, errNotFound()
	}
	return r.FindByID(ctx, tx.ID)
}

// SoftDelete marks a transaction deleted; it stays for the history and is left out of balances
func (r *repo) SoftDelete(ctx context.Context, id, deletedBy int64) error {
	query, args, err := builder.
		Update(transactionsTable).
		Set("deleted_at", time.Now().UTC()).
		Set("deleted_by", database.NullableID(deletedBy)).
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building delete transaction query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return errInternal(fmt.Errorf("error deleting transaction: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return errNotFound()
	}
	return nil
}

// UsesCategory is whether any transaction, deleted ones too, is in the category
func (r *repo) UsesCategory(ctx context.Context, categoryID int64) (bool, error) {
	query, args, err := builder.Select("id").From(transactionsTable).Where(sq.Eq{"category_id": categoryID}).Limit(1).ToSql()
	if err != nil {
		return false, errInternal(fmt.Errorf("error building category use query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return false, nil
		}
		return false, errInternal(fmt.Errorf("error checking category use: %w", err))
	}
	return true, nil
}

// SavePhoto inserts or replaces a transaction's receipt photo
func (r *repo) SavePhoto(ctx context.Context, transactionID int64, photo domain.TransactionPhoto) error {
	query, args, err := builder.
		Insert(photosTable).
		Columns("transaction_id", "content_type", "data").
		Values(transactionID, photo.ContentType, photo.Data).
		// same syntax on PostgreSQL and SQLite
		Suffix("ON CONFLICT (transaction_id) DO UPDATE SET content_type = excluded.content_type, data = excluded.data").
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building save photo query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error saving photo: %w", err))
	}
	return nil
}

// GetPhoto returns a transaction's receipt photo, or domain.ErrTransactionNotFound without one
func (r *repo) GetPhoto(ctx context.Context, transactionID int64) (domain.TransactionPhoto, error) {
	query, args, err := builder.Select("content_type", "data").From(photosTable).Where(sq.Eq{"transaction_id": transactionID}).ToSql()
	if err != nil {
		return domain.TransactionPhoto{}, errInternal(fmt.Errorf("error building get photo query: %w", err))
	}

	var m struct {
		ContentType string `db:"content_type"`
		Data        []byte `db:"data"`
	}
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.TransactionPhoto{}, apperror.New(apperror.CodeNotFound, domain.ErrTransactionNotFound, "foto bukti tidak ditemukan")
		}
		return domain.TransactionPhoto{}, errInternal(fmt.Errorf("error getting photo: %w", err))
	}
	return domain.TransactionPhoto{ContentType: m.ContentType, Data: m.Data}, nil
}

// AddRevision stores the transaction as it was before an edit or a delete
func (r *repo) AddRevision(ctx context.Context, rev domain.TransactionRevision) error {
	previous, err := json.Marshal(rev.Previous)
	if err != nil {
		return errInternal(fmt.Errorf("error encoding revision: %w", err))
	}

	query, args, err := builder.
		Insert(revisionsTable).
		Columns("transaction_id", "action", "previous", "reason", "user_id", "created_at").
		Values(rev.TransactionID, string(rev.Action), string(previous), rev.Reason, database.NullableID(rev.UserID), time.Now().UTC()).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building add revision query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error adding revision: %w", err))
	}
	return nil
}

// FindRevisions returns a transaction's revisions, oldest first
func (r *repo) FindRevisions(ctx context.Context, transactionID int64) ([]domain.TransactionRevision, error) {
	query, args, err := builder.
		Select("id", "transaction_id", "action", "previous", "reason", "user_id", "created_at").
		From(revisionsTable).
		Where(sq.Eq{"transaction_id": transactionID}).
		OrderBy("id").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find revisions query: %w", err))
	}

	var models []struct {
		ID            int64     `db:"id"`
		TransactionID int64     `db:"transaction_id"`
		Action        string    `db:"action"`
		Previous      string    `db:"previous"`
		Reason        string    `db:"reason"`
		UserID        *int64    `db:"user_id"`
		CreatedAt     time.Time `db:"created_at"`
	}
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding revisions: %w", err))
	}

	revisions := make([]domain.TransactionRevision, 0, len(models))
	for _, m := range models {
		var previous domain.Transaction
		if err := json.Unmarshal([]byte(m.Previous), &previous); err != nil {
			return nil, errInternal(fmt.Errorf("error decoding revision %d: %w", m.ID, err))
		}
		revisions = append(revisions, domain.TransactionRevision{
			ID:            m.ID,
			TransactionID: m.TransactionID,
			Action:        domain.RevisionAction(m.Action),
			Previous:      previous,
			Reason:        m.Reason,
			UserID:        database.IDOrZero(m.UserID),
			CreatedAt:     m.CreatedAt,
		})
	}
	return revisions, nil
}

func errNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrTransactionNotFound, "transaksi tidak ditemukan")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data transaksi")
}
