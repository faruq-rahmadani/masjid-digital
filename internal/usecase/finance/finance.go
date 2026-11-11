// Package finance is the Keuangan usecase (E7): kas, akun, kategori, pembukuan and, in later
// phases, transactions and tutup buku. Balances are computed here from the last closing plus the
// transactions after it; the repositories never sum.
package finance

import (
	"context"
	"errors"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// Repositories are the stores the finance usecase reads and writes
type Repositories struct {
	Transactor     usecase.TransactorItf
	Accounts       usecase.AccountRepositoryItf
	Pockets        usecase.PocketRepositoryItf
	Categories     usecase.CategoryRepositoryItf
	Transactions   usecase.TransactionRepositoryItf
	BookClosings   usecase.BookClosingRepositoryItf
	MosqueProfiles usecase.MosqueProfileRepositoryItf // the mosque's time zone decides "today"
}

type UseCase struct {
	Repositories
	activityLog usecase.ActivityLogPublisherItf
	now         func() time.Time
}

func New(repos Repositories, activityLog usecase.ActivityLogPublisherItf) *UseCase {
	return &UseCase{Repositories: repos, activityLog: activityLog, now: time.Now}
}

// balances is what every account and every kas holds now
type balances struct {
	accounts map[int64]int64
	pockets  map[int64]int64
}

// balances starts from the last closing (the opening before any month is closed) and adds the
// transactions dated after its month. Before pembukuan starts everything is Rp 0.
func (u *UseCase) balances(ctx context.Context) (balances, error) {
	b := balances{accounts: map[int64]int64{}, pockets: map[int64]int64{}}

	closing, err := u.BookClosings.FindLatest(ctx)
	if errors.Is(err, domain.ErrBookClosingNotFound) {
		return b, nil
	}
	if err != nil {
		return balances{}, err
	}
	for _, a := range closing.Accounts {
		b.accounts[a.AccountID] = a.ClosingBalance
	}
	for _, p := range closing.Pockets {
		b.pockets[p.PocketID] = p.ClosingBalance
	}

	txs, err := u.Transactions.Find(ctx, domain.TransactionFilter{From: domain.FirstDayAfter(closing.Month)})
	if err != nil {
		return balances{}, err
	}
	for _, tx := range txs {
		b.accounts[tx.AccountID] += tx.Signed()
		b.pockets[tx.PocketID] += tx.Signed()
	}
	return b, nil
}

// today is the current time on the mosque's clock; WIB until the profile is saved
func (u *UseCase) today(ctx context.Context) (time.Time, error) {
	profile, err := u.MosqueProfiles.Get(ctx)
	if err != nil && !errors.Is(err, domain.ErrMosqueProfileNotFound) {
		return time.Time{}, err
	}
	return u.now().In(profile.Timezone.Location()), nil
}

func errRule(err error, message string) *apperror.AppError {
	return apperror.New(apperror.CodeBusinessRule, err, message)
}
