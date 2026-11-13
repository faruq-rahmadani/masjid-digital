package transactions

import (
	"context"
	"errors"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/dbtest"
)

// seeded IDs: Tunai is account 1; Kas Umum pocket 1, Zakat Mal 2; categories by code or name
type fixture struct {
	repo                             *repo
	userID, tunai, umum, zakat       int64
	bank, infaq, pindahIn, pindahOut int64
}

func setup(t *testing.T, db dbtest.DB) fixture {
	t.Helper()
	ctx := context.Background()
	f := fixture{repo: New(db.Client), userID: db.UserID}
	mustGet := func(dest *int64, query string, args ...any) {
		if err := db.Get(ctx, dest, query, args...); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
	}
	mustGet(&f.tunai, "SELECT id FROM accounts WHERE name = 'Tunai'")
	mustGet(&f.bank, "INSERT INTO accounts (name, kind, created_at, updated_at) VALUES ('BSI', 'bank', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id")
	mustGet(&f.umum, "SELECT id FROM pockets WHERE kind = 'umum'")
	mustGet(&f.zakat, "SELECT id FROM pockets WHERE kind = 'zakat_mal'")
	mustGet(&f.infaq, "SELECT id FROM categories WHERE name = 'Infaq Jumat'")
	mustGet(&f.pindahIn, "SELECT id FROM categories WHERE code = ?", domain.CategoryCodePindahIn)
	mustGet(&f.pindahOut, "SELECT id FROM categories WHERE code = ?", domain.CategoryCodePindahOut)
	return f
}

func TestTransactions(t *testing.T) {
	for driver, db := range dbtest.Open(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			f := setup(t, db)

			infaq, err := f.repo.Create(ctx, domain.Transaction{
				Kind: domain.CategoryTypeIn, Amount: 2_000_000, Date: "2026-11-06",
				AccountID: f.tunai, PocketID: f.umum, CategoryID: f.infaq, Note: "Infaq Jumat 6 Nov", CreatedBy: f.userID,
			})
			if err != nil || infaq.ID == 0 || infaq.IsEdited() || infaq.HasPhoto || infaq.UpdatedBy != f.userID {
				t.Fatalf("create = %+v, %v", infaq, err)
			}

			// setor tunai: an out + in pair sharing a group
			group := "pindah-1"
			for _, side := range []domain.Transaction{
				{Kind: domain.CategoryTypeOut, AccountID: f.tunai, CategoryID: f.pindahOut},
				{Kind: domain.CategoryTypeIn, AccountID: f.bank, CategoryID: f.pindahIn},
			} {
				side.Amount, side.Date, side.PocketID, side.GroupID, side.CreatedBy = 1_500_000, "2026-11-09", f.umum, new(group), f.userID
				if _, err := f.repo.Create(ctx, side); err != nil {
					t.Fatalf("create pindah side: %v", err)
				}
			}

			byGroup, err := f.repo.Find(ctx, domain.TransactionFilter{GroupID: group})
			if err != nil || len(byGroup) != 2 || byGroup[0].Signed()+byGroup[1].Signed() != 0 {
				t.Errorf("group = %+v, %v; want a balanced pair", byGroup, err)
			}
			nov8, _ := f.repo.Find(ctx, domain.TransactionFilter{From: "2026-11-01", To: "2026-11-08"})
			if len(nov8) != 1 || nov8[0].ID != infaq.ID {
				t.Errorf("1-8 Nov = %+v; want only the infaq", nov8)
			}
			inBank, _ := f.repo.Find(ctx, domain.TransactionFilter{AccountID: f.bank})
			if len(inBank) != 1 || inBank[0].Kind != domain.CategoryTypeIn {
				t.Errorf("bank = %+v; want the setor's in side", inBank)
			}
			if inZakat, _ := f.repo.Find(ctx, domain.TransactionFilter{PocketID: f.zakat}); len(inZakat) != 0 {
				t.Errorf("zakat = %+v; want none", inZakat)
			}

			// edit, with the old row kept as a revision
			if err := f.repo.AddRevision(ctx, domain.TransactionRevision{TransactionID: infaq.ID, Action: domain.RevisionEdit, Previous: infaq, Reason: "salah ketik", UserID: f.userID}); err != nil {
				t.Fatalf("add revision: %v", err)
			}
			edited := infaq
			edited.Amount, edited.UpdatedBy = 2_100_000, f.userID
			got, err := f.repo.Update(ctx, edited)
			if err != nil || got.Amount != 2_100_000 || !got.IsEdited() {
				t.Errorf("update = %+v, %v", got, err)
			}
			revisions, err := f.repo.FindRevisions(ctx, infaq.ID)
			if err != nil || len(revisions) != 1 || revisions[0].Previous.Amount != 2_000_000 || revisions[0].Reason != "salah ketik" {
				t.Errorf("revisions = %+v, %v", revisions, err)
			}

			// photo: saved, replaced, read back
			for _, data := range [][]byte{{1, 2, 3}, {4, 5}} {
				if err := f.repo.SavePhoto(ctx, infaq.ID, domain.TransactionPhoto{ContentType: "image/jpeg", Data: data}); err != nil {
					t.Fatalf("save photo: %v", err)
				}
			}
			photo, err := f.repo.GetPhoto(ctx, infaq.ID)
			if err != nil || len(photo.Data) != 2 {
				t.Errorf("photo = %v, %v; want the replacement", photo.Data, err)
			}
			if got, _ := f.repo.FindByID(ctx, infaq.ID); !got.HasPhoto {
				t.Error("has_photo not set")
			}
			if _, err := f.repo.GetPhoto(ctx, 999); !errors.Is(err, domain.ErrTransactionNotFound) {
				t.Errorf("no photo: err = %v", err)
			}

			// soft delete: gone from the default list, kept with IncludeDeleted
			if err := f.repo.SoftDelete(ctx, infaq.ID, f.userID); err != nil {
				t.Fatalf("soft delete: %v", err)
			}
			if err := f.repo.SoftDelete(ctx, infaq.ID, f.userID); !errors.Is(err, domain.ErrTransactionNotFound) {
				t.Errorf("delete twice: err = %v", err)
			}
			if live, _ := f.repo.Find(ctx, domain.TransactionFilter{}); len(live) != 2 {
				t.Errorf("live = %d; want the 2 pindah rows", len(live))
			}
			withDeleted, _ := f.repo.Find(ctx, domain.TransactionFilter{IncludeDeleted: true})
			if len(withDeleted) != 3 || withDeleted[0].DeletedAt == nil || withDeleted[0].DeletedBy != f.userID {
				t.Errorf("with deleted = %+v", withDeleted)
			}

			if used, err := f.repo.UsesCategory(ctx, f.infaq); err != nil || !used {
				t.Errorf("infaq used = %v, %v; want true (deleted rows count)", used, err)
			}
			var listrik int64
			_ = db.Get(ctx, &listrik, "SELECT id FROM categories WHERE name = 'Listrik'")
			if used, err := f.repo.UsesCategory(ctx, listrik); err != nil || used {
				t.Errorf("listrik used = %v, %v; want false", used, err)
			}
		})
	}
}
