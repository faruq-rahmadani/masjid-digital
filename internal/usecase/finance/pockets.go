package finance

import (
	"context"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

// ListPockets returns every kas, archived ones too, with its balance
func (u *UseCase) ListPockets(ctx context.Context) ([]domain.PocketWithBalance, error) {
	pockets, err := u.Pockets.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	b, err := u.balances(ctx)
	if err != nil {
		return nil, err
	}

	progress, err := u.progress(ctx, pockets)
	if err != nil {
		return nil, err
	}

	list := make([]domain.PocketWithBalance, 0, len(pockets))
	for _, p := range pockets {
		item := domain.PocketWithBalance{Pocket: p, Balance: b.pockets[p.ID]}
		if pr, ok := progress[p.ID]; ok {
			item.Progress = &pr
		}
		list = append(list, item)
	}
	return list, nil
}

// CreatePocket adds a terikat kas; users never add built-in ones
func (u *UseCase) CreatePocket(ctx context.Context, req domain.CreatePocketRequest) (domain.Pocket, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.Pocket{}, err
	}
	pocket, err := u.Pockets.Create(ctx, req.ToPocket())
	if err != nil {
		return domain.Pocket{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(pocket.ID))
	return pocket, nil
}

// UpdatePocket edits a terikat kas, or only "tampil di TV" of a built-in one. The kind never
// changes.
func (u *UseCase) UpdatePocket(ctx context.Context, req domain.UpdatePocketRequest) (domain.Pocket, error) {
	pocket, err := u.Pockets.FindByID(ctx, req.ID)
	if err != nil {
		return domain.Pocket{}, err
	}
	if pocket.Status == domain.PocketStatusArchived {
		return domain.Pocket{}, errRule(fmt.Errorf("pocket %d is archived", pocket.ID), "kas yang diarsipkan tidak bisa diubah, pulihkan dulu")
	}
	changed, err := req.Apply(pocket)
	if err != nil {
		return domain.Pocket{}, err
	}

	saved, err := u.Pockets.Update(ctx, changed)
	if err != nil {
		return domain.Pocket{}, err
	}
	u.activityLog.Publish(ctx, domain.PocketUpdatedActivityLog(saved))
	return saved, nil
}

// ArchivePocket closes a terikat kas at Rp 0; the built-in kas always stay
func (u *UseCase) ArchivePocket(ctx context.Context, id int64) (domain.Pocket, error) {
	pocket, err := u.Pockets.FindByID(ctx, id)
	if err != nil {
		return domain.Pocket{}, err
	}
	if pocket.Kind.IsBuiltIn() {
		return domain.Pocket{}, errRule(fmt.Errorf("pocket %d is built in", id), "kas bawaan tidak bisa diarsipkan")
	}
	if pocket.Status == domain.PocketStatusArchived {
		return pocket, nil
	}

	b, err := u.balances(ctx)
	if err != nil {
		return domain.Pocket{}, err
	}
	if balance := b.pockets[id]; balance != 0 {
		return domain.Pocket{}, errRule(fmt.Errorf("pocket %d holds %d", id, balance), "kas hanya bisa diarsipkan saat saldonya Rp 0")
	}

	pocket.Status = domain.PocketStatusArchived
	return u.savePocketArchive(ctx, pocket, true)
}

// RestorePocket reopens an archived terikat kas
func (u *UseCase) RestorePocket(ctx context.Context, id int64) (domain.Pocket, error) {
	pocket, err := u.Pockets.FindByID(ctx, id)
	if err != nil {
		return domain.Pocket{}, err
	}
	if pocket.Status != domain.PocketStatusArchived {
		return pocket, nil
	}

	pocket.Status = domain.PocketStatusActive
	return u.savePocketArchive(ctx, pocket, false)
}

func (u *UseCase) savePocketArchive(ctx context.Context, pocket domain.Pocket, archived bool) (domain.Pocket, error) {
	saved, err := u.Pockets.Update(ctx, pocket)
	if err != nil {
		return domain.Pocket{}, err
	}
	u.activityLog.Publish(ctx, domain.PocketArchiveActivityLog(saved, archived))
	return saved, nil
}
