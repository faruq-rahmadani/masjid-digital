package pockets

import (
	"context"
	"errors"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/dbtest"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func TestPockets(t *testing.T) {
	for driver, db := range dbtest.Open(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			repo := New(db.Client)

			// the four built-in pockets are seeded; only Kas Umum shows on the TV
			all, err := repo.FindAll(ctx)
			if err != nil || len(all) != 4 {
				t.Fatalf("seeded = %+v, %v", all, err)
			}
			kinds := []domain.PocketKind{domain.PocketKindUmum, domain.PocketKindZakatMal, domain.PocketKindZakatFitrah, domain.PocketKindFidyah}
			for i, p := range all {
				if p.Kind != kinds[i] || p.ShowOnTV != (p.Kind == domain.PocketKindUmum) || p.Status != domain.PocketStatusActive || p.Purpose == "" {
					t.Errorf("seeded[%d] = %+v", i, p)
				}
			}

			target, plan := int64(150_000_000), domain.SurplusPlanToUmum
			renovasi, err := repo.Create(ctx, domain.Pocket{
				Kind: domain.PocketKindTerikat, Name: "Renovasi Atap", Purpose: "Penggantian atap dan talang masjid.",
				TargetAmount: &target, SurplusPlan: &plan, ShowOnTV: true, Status: domain.PocketStatusActive, CreatedBy: db.UserID,
			})
			if err != nil || renovasi.TargetAmount == nil || *renovasi.TargetAmount != target || renovasi.SurplusPlan == nil || *renovasi.SurplusPlan != plan {
				t.Fatalf("create = %+v, %v", renovasi, err)
			}

			var appErr *apperror.AppError
			if _, err := repo.Create(ctx, domain.Pocket{Kind: domain.PocketKindTerikat, Name: "Kas Umum", Status: domain.PocketStatusActive}); !errors.As(err, &appErr) || appErr.Code != apperror.CodeAlreadyExists {
				t.Errorf("duplicate name: err = %v, want ALREADY_EXISTS", err)
			}

			renovasi.Status, renovasi.ShowOnTV, renovasi.TargetAmount = domain.PocketStatusDone, false, nil
			got, err := repo.Update(ctx, renovasi)
			if err != nil || got.Status != domain.PocketStatusDone || got.ShowOnTV || got.TargetAmount != nil || got.Kind != domain.PocketKindTerikat {
				t.Errorf("update = %+v, %v", got, err)
			}

			if _, err := repo.FindByID(ctx, 999); !errors.Is(err, domain.ErrPocketNotFound) {
				t.Errorf("missing: err = %v, want ErrPocketNotFound", err)
			}
		})
	}
}
