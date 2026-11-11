package finance

import (
	"context"
	"errors"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

// GetBookkeeping returns whether pembukuan has started and, once it has, its opening balances
func (u *UseCase) GetBookkeeping(ctx context.Context) (domain.Bookkeeping, error) {
	closings, err := u.BookClosings.FindAll(ctx)
	if err != nil {
		return domain.Bookkeeping{}, err
	}
	for _, c := range closings {
		if c.Kind == domain.BookClosingOpening {
			return domain.NewBookkeeping(c), nil
		}
	}
	return domain.Bookkeeping{}, nil
}

// StartBookkeeping is "Mulai pembukuan", once: the opening balance of every account and kas,
// saved as the closing of the month before StartMonth. It is never pemasukan.
func (u *UseCase) StartBookkeeping(ctx context.Context, req domain.StartBookkeepingRequest) (domain.Bookkeeping, error) {
	today, err := u.today(ctx)
	if err != nil {
		return domain.Bookkeeping{}, err
	}
	if err := req.ValidateAndNormalize(today.Format("2006-01")); err != nil {
		return domain.Bookkeeping{}, err
	}

	started, err := u.GetBookkeeping(ctx)
	if err != nil {
		return domain.Bookkeeping{}, err
	}
	if started.Started {
		return domain.Bookkeeping{}, errRule(errors.New("bookkeeping already started"), "pembukuan sudah dimulai")
	}

	accountIDs, err := u.openAccountIDs(ctx, req.Accounts)
	if err != nil {
		return domain.Bookkeeping{}, err
	}
	pocketIDs, err := u.openPocketIDs(ctx, req.Pockets)
	if err != nil {
		return domain.Bookkeeping{}, err
	}

	var opening domain.BookClosing
	err = u.Transactor.WithTx(ctx, func(ctx context.Context) error {
		var err error
		opening, err = u.BookClosings.Create(ctx, req.OpeningClosing(accountIDs, pocketIDs))
		return err
	})
	if err != nil {
		return domain.Bookkeeping{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(opening.ID))
	return domain.NewBookkeeping(opening), nil
}

// openAccountIDs returns every account not archived, after checking each one given is among them
func (u *UseCase) openAccountIDs(ctx context.Context, given []domain.OpeningBalance) ([]int64, error) {
	accounts, err := u.Accounts.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	open := make(map[int64]bool, len(accounts))
	ids := make([]int64, 0, len(accounts))
	for _, a := range accounts {
		if a.ArchivedAt == nil {
			open[a.ID] = true
			ids = append(ids, a.ID)
		}
	}
	return ids, checkGiven(given, open, "akun")
}

// openPocketIDs returns every kas still taking transactions, after checking each one given is
// among them
func (u *UseCase) openPocketIDs(ctx context.Context, given []domain.OpeningBalance) ([]int64, error) {
	pockets, err := u.Pockets.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	open := make(map[int64]bool, len(pockets))
	ids := make([]int64, 0, len(pockets))
	for _, p := range pockets {
		if p.IsOpen() {
			open[p.ID] = true
			ids = append(ids, p.ID)
		}
	}
	return ids, checkGiven(given, open, "kas")
}

func checkGiven(given []domain.OpeningBalance, open map[int64]bool, what string) error {
	for _, b := range given {
		if !open[b.ID] {
			return errRule(errors.New(what+" not found or archived"), what+" tidak ditemukan atau sudah diarsipkan")
		}
	}
	return nil
}
