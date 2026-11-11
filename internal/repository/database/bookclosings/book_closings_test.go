package bookclosings

import (
	"context"
	"errors"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/dbtest"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func TestBookClosings(t *testing.T) {
	for driver, db := range dbtest.Open(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			repo := New(db.Client)

			if _, err := repo.FindLatest(ctx); !errors.Is(err, domain.ErrBookClosingNotFound) {
				t.Fatalf("before pembukuan: err = %v, want ErrBookClosingNotFound", err)
			}

			var tunai, umum, zakat int64
			_ = db.Get(ctx, &tunai, "SELECT id FROM accounts WHERE name = 'Tunai'")
			_ = db.Get(ctx, &umum, "SELECT id FROM pockets WHERE kind = 'umum'")
			_ = db.Get(ctx, &zakat, "SELECT id FROM pockets WHERE kind = 'zakat_mal'")

			// Mulai pembukuan: Tunai holds 12.5 juta, split over Kas Umum and Zakat Mal
			opening, err := repo.Create(ctx, domain.BookClosing{
				Month: "2026-09", Kind: domain.BookClosingOpening, ClosingBalance: 12_500_000, ClosedBy: db.UserID,
				Accounts: []domain.BookClosingAccount{{AccountID: tunai, ClosingBalance: 12_500_000}},
				Pockets:  []domain.BookClosingPocket{{PocketID: umum, ClosingBalance: 10_000_000}, {PocketID: zakat, ClosingBalance: 2_500_000}},
			})
			if err != nil || opening.Kind != domain.BookClosingOpening || len(opening.Accounts) != 1 || len(opening.Pockets) != 2 || opening.Accounts[0].CountedBalance != nil {
				t.Fatalf("opening = %+v, %v", opening, err)
			}

			counted := int64(14_800_000)
			if _, err := repo.Create(ctx, domain.BookClosing{
				Month: "2026-10", Kind: domain.BookClosingClosing, TotalIn: 4_200_000, TotalOut: 1_900_000, ClosingBalance: 14_800_000,
				Accounts: []domain.BookClosingAccount{{AccountID: tunai, ClosingBalance: 14_800_000, CountedBalance: &counted}},
				Pockets:  []domain.BookClosingPocket{{PocketID: umum, ClosingBalance: 12_300_000}, {PocketID: zakat, ClosingBalance: 2_500_000}},
			}); err != nil {
				t.Fatalf("close October: %v", err)
			}

			latest, err := repo.FindLatest(ctx)
			if err != nil || latest.Month != "2026-10" || latest.TotalIn != 4_200_000 || *latest.Accounts[0].CountedBalance != counted {
				t.Errorf("latest = %+v, %v", latest, err)
			}
			all, _ := repo.FindAll(ctx)
			if len(all) != 2 || all[0].Month != "2026-09" || all[1].Month != "2026-10" {
				t.Errorf("all = %+v; want Sep then Oct", all)
			}

			var appErr *apperror.AppError
			if _, err := repo.Create(ctx, domain.BookClosing{Month: "2026-10", Kind: domain.BookClosingClosing}); !errors.As(err, &appErr) || appErr.Code != apperror.CodeAlreadyExists {
				t.Errorf("close October twice: err = %v, want ALREADY_EXISTS", err)
			}
		})
	}
}
