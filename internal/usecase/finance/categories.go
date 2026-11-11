package finance

import (
	"context"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

// ListCategories returns every category, archived ones too, with how many transactions use it
func (u *UseCase) ListCategories(ctx context.Context) ([]domain.CategoryWithCount, error) {
	categories, err := u.Categories.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	txs, err := u.Transactions.Find(ctx, domain.TransactionFilter{})
	if err != nil {
		return nil, err
	}

	counts := make(map[int64]int, len(categories))
	for _, tx := range txs {
		counts[tx.CategoryID]++
	}
	list := make([]domain.CategoryWithCount, 0, len(categories))
	for _, c := range categories {
		list = append(list, domain.CategoryWithCount{Category: c, TransactionCount: counts[c.ID]})
	}
	return list, nil
}

// CreateCategory adds an own category for Kas Umum and terikat kas; zakat and fidyah only use
// the system ones
func (u *UseCase) CreateCategory(ctx context.Context, req domain.CreateCategoryRequest) (domain.Category, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.Category{}, err
	}
	category, err := u.Categories.Create(ctx, req.ToCategory())
	if err != nil {
		return domain.Category{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(category.ID))
	return category, nil
}

// UpdateCategory renames an own category
func (u *UseCase) UpdateCategory(ctx context.Context, req domain.UpdateCategoryRequest) (domain.Category, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.Category{}, err
	}
	category, err := u.ownCategory(ctx, req.ID, "diubah namanya")
	if err != nil {
		return domain.Category{}, err
	}

	category.Name = req.Name
	saved, err := u.Categories.Update(ctx, category)
	if err != nil {
		return domain.Category{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog())
	return saved, nil
}

// DeleteCategory removes an own category no transaction ever used; a used one is archived instead
func (u *UseCase) DeleteCategory(ctx context.Context, id int64) error {
	category, err := u.ownCategory(ctx, id, "dihapus")
	if err != nil {
		return err
	}
	used, err := u.Transactions.UsesCategory(ctx, id)
	if err != nil {
		return err
	}
	if used {
		return errRule(fmt.Errorf("category %d has transactions", id), "kategori sudah dipakai transaksi, arsipkan saja")
	}

	if err := u.Categories.Delete(ctx, id); err != nil {
		return err
	}
	u.activityLog.Publish(ctx, domain.CategoryDeletedActivityLog(category))
	return nil
}

// ArchiveCategory hides a category from the form; asnaf and fidyah recipients may be hidden too
func (u *UseCase) ArchiveCategory(ctx context.Context, id int64) (domain.Category, error) {
	category, err := u.Categories.FindByID(ctx, id)
	if err != nil {
		return domain.Category{}, err
	}
	if !category.CanArchive() {
		return domain.Category{}, errRule(fmt.Errorf("system category %d cannot be archived", id), "kategori sistem ini tidak bisa diarsipkan")
	}
	if category.ArchivedAt != nil {
		return category, nil
	}

	now := u.now().UTC()
	category.ArchivedAt = &now
	return u.saveCategoryArchive(ctx, category, true)
}

// RestoreCategory shows an archived category in the form again
func (u *UseCase) RestoreCategory(ctx context.Context, id int64) (domain.Category, error) {
	category, err := u.Categories.FindByID(ctx, id)
	if err != nil {
		return domain.Category{}, err
	}
	if category.ArchivedAt == nil {
		return category, nil
	}

	category.ArchivedAt = nil
	return u.saveCategoryArchive(ctx, category, false)
}

func (u *UseCase) saveCategoryArchive(ctx context.Context, category domain.Category, archived bool) (domain.Category, error) {
	saved, err := u.Categories.Update(ctx, category)
	if err != nil {
		return domain.Category{}, err
	}
	u.activityLog.Publish(ctx, domain.CategoryArchiveActivityLog(saved, archived))
	return saved, nil
}

// ownCategory returns a category pengurus made; a system one carries rules and keeps its name.
// done is what was refused, for the message ("kategori sistem tidak bisa <done>").
func (u *UseCase) ownCategory(ctx context.Context, id int64, done string) (domain.Category, error) {
	category, err := u.Categories.FindByID(ctx, id)
	if err != nil {
		return domain.Category{}, err
	}
	if category.IsSystem() {
		return domain.Category{}, errRule(fmt.Errorf("category %d is a system category", id), "kategori sistem tidak bisa "+done)
	}
	return category, nil
}
