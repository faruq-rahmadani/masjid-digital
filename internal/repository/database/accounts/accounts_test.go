package accounts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/dbtest"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func TestAccounts(t *testing.T) {
	for driver, db := range dbtest.Open(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			repo := New(db.Client)

			// Tunai is seeded by the migration
			all, err := repo.FindAll(ctx)
			if err != nil || len(all) != 1 || all[0].Name != "Tunai" || all[0].Kind != domain.AccountKindCash || all[0].CreatedAt.IsZero() {
				t.Fatalf("seeded = %+v, %v; want Tunai", all, err)
			}

			bank, err := repo.Create(ctx, domain.Account{Name: "BSI Masjid", Kind: domain.AccountKindBank, Number: "7123456789", CreatedBy: db.UserID})
			if err != nil || bank.ID == 0 || bank.Number != "7123456789" || bank.CreatedBy != db.UserID {
				t.Fatalf("create = %+v, %v", bank, err)
			}

			var appErr *apperror.AppError
			if _, err := repo.Create(ctx, domain.Account{Name: "BSI Masjid", Kind: domain.AccountKindBank}); !errors.As(err, &appErr) || appErr.Code != apperror.CodeAlreadyExists {
				t.Errorf("duplicate name: err = %v, want ALREADY_EXISTS", err)
			}

			archived := time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC)
			bank.Name, bank.ArchivedAt = "BSI Masjid Lama", &archived
			got, err := repo.Update(ctx, bank)
			if err != nil || got.Name != "BSI Masjid Lama" || got.ArchivedAt == nil || !got.ArchivedAt.Equal(archived) {
				t.Errorf("update = %+v, %v", got, err)
			}

			if _, err := repo.FindByID(ctx, 999); !errors.Is(err, domain.ErrAccountNotFound) {
				t.Errorf("missing: err = %v, want ErrAccountNotFound", err)
			}
		})
	}
}
