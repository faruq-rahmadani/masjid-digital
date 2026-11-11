package finance

import (
	"context"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

// ListAccounts returns every account, archived ones too, with its balance
func (u *UseCase) ListAccounts(ctx context.Context) ([]domain.AccountWithBalance, error) {
	accounts, err := u.Accounts.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	b, err := u.balances(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]domain.AccountWithBalance, 0, len(accounts))
	for _, a := range accounts {
		list = append(list, domain.AccountWithBalance{Account: a, Balance: b.accounts[a.ID]})
	}
	return list, nil
}

func (u *UseCase) CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (domain.Account, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.Account{}, err
	}
	account, err := u.Accounts.Create(ctx, domain.Account{Name: req.Name, Kind: req.Kind, Number: req.Number, CreatedBy: req.CreatedBy})
	if err != nil {
		return domain.Account{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(account.ID))
	return account, nil
}

// UpdateAccount renames an account or changes its number; the kind never changes
func (u *UseCase) UpdateAccount(ctx context.Context, req domain.UpdateAccountRequest) (domain.Account, error) {
	account, err := u.Accounts.FindByID(ctx, req.ID)
	if err != nil {
		return domain.Account{}, err
	}
	if err := req.ValidateAndNormalize(account.Kind); err != nil {
		return domain.Account{}, err
	}

	account.Name, account.Number = req.Name, req.Number
	saved, err := u.Accounts.Update(ctx, account)
	if err != nil {
		return domain.Account{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog())
	return saved, nil
}

// ArchiveAccount hides an account from the forms. Only an empty account may go, and one cash
// account always stays, since the transaction form defaults to it.
func (u *UseCase) ArchiveAccount(ctx context.Context, id int64) (domain.Account, error) {
	accounts, err := u.Accounts.FindAll(ctx)
	if err != nil {
		return domain.Account{}, err
	}
	account, activeCash := domain.Account{}, 0
	for _, a := range accounts {
		if a.ID == id {
			account = a
		}
		if a.Kind == domain.AccountKindCash && a.ArchivedAt == nil {
			activeCash++
		}
	}
	if account.ID == 0 {
		return u.Accounts.FindByID(ctx, id) // the repository's not-found error
	}
	if account.ArchivedAt != nil {
		return account, nil
	}
	if account.Kind == domain.AccountKindCash && activeCash == 1 {
		return domain.Account{}, errRule(fmt.Errorf("account %d is the last cash account", id), "akun tunai terakhir tidak bisa diarsipkan")
	}

	b, err := u.balances(ctx)
	if err != nil {
		return domain.Account{}, err
	}
	if balance := b.accounts[id]; balance != 0 {
		return domain.Account{}, errRule(fmt.Errorf("account %d holds %d", id, balance), "akun hanya bisa diarsipkan saat saldonya Rp 0")
	}

	return u.setAccountArchived(ctx, account, true)
}

// RestoreAccount brings an archived account back to the forms
func (u *UseCase) RestoreAccount(ctx context.Context, id int64) (domain.Account, error) {
	account, err := u.Accounts.FindByID(ctx, id)
	if err != nil {
		return domain.Account{}, err
	}
	if account.ArchivedAt == nil {
		return account, nil
	}
	return u.setAccountArchived(ctx, account, false)
}

func (u *UseCase) setAccountArchived(ctx context.Context, account domain.Account, archived bool) (domain.Account, error) {
	account.ArchivedAt = nil
	if archived {
		now := u.now().UTC()
		account.ArchivedAt = &now
	}
	saved, err := u.Accounts.Update(ctx, account)
	if err != nil {
		return domain.Account{}, err
	}
	u.activityLog.Publish(ctx, domain.AccountArchiveActivityLog(saved, archived))
	return saved, nil
}
