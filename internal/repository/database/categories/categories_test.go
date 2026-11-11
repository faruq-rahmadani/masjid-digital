package categories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/dbtest"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func TestCategories(t *testing.T) {
	for driver, db := range dbtest.Open(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			repo := New(db.Client)

			all, err := repo.FindAll(ctx)
			if err != nil {
				t.Fatalf("find all: %v", err)
			}
			system, user := 0, 0
			for _, c := range all {
				if c.IsSystem() {
					system++
				} else {
					user++
				}
			}
			// pindah in/out, 2 zakat in, 8 asnaf, fidyah in + 2 out; 9 editable defaults for umum
			if system != 15 || user != 9 {
				t.Errorf("seeded system = %d, user = %d; want 15, 9", system, user)
			}

			amil, err := repo.FindByCode(ctx, domain.CategoryCodeAmil)
			if err != nil || amil.Name != "Amil" || amil.Type != domain.CategoryTypeOut || amil.Scope != domain.CategoryScopeZakat {
				t.Errorf("amil = %+v, %v", amil, err)
			}

			sound, err := repo.Create(ctx, domain.Category{Name: "Sound system", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeUmumTerikat, CreatedBy: db.UserID})
			if err != nil || sound.IsSystem() || sound.CreatedBy != db.UserID {
				t.Fatalf("create = %+v, %v", sound, err)
			}

			var appErr *apperror.AppError
			if _, err := repo.Create(ctx, domain.Category{Name: "Listrik", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeUmumTerikat}); !errors.As(err, &appErr) || appErr.Code != apperror.CodeAlreadyExists {
				t.Errorf("duplicate: err = %v, want ALREADY_EXISTS", err)
			}
			// the same name is fine as an income category
			if _, err := repo.Create(ctx, domain.Category{Name: "Listrik", Type: domain.CategoryTypeIn, Scope: domain.CategoryScopeUmumTerikat}); err != nil {
				t.Errorf("same name, other type: %v", err)
			}

			archived := time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC)
			sound.Name, sound.ArchivedAt = "Sound & mic", &archived
			if got, err := repo.Update(ctx, sound); err != nil || got.Name != "Sound & mic" || got.ArchivedAt == nil {
				t.Errorf("update = %+v, %v", got, err)
			}

			if err := repo.Delete(ctx, sound.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
			if _, err := repo.FindByID(ctx, sound.ID); !errors.Is(err, domain.ErrCategoryNotFound) {
				t.Errorf("deleted: err = %v, want ErrCategoryNotFound", err)
			}
			if err := repo.Delete(ctx, sound.ID); !errors.Is(err, domain.ErrCategoryNotFound) {
				t.Errorf("delete again: err = %v, want ErrCategoryNotFound", err)
			}
		})
	}
}
