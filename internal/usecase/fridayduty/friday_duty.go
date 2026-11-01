package fridayduty

import (
	"context"
	"fmt"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// UpcomingDays is how many days of petugas Jumat the public list carries, so a TV that loses its
// connection still shows them for a month
const UpcomingDays = 30

const dateLayout = "2006-01-02"

type UseCase struct {
	duties usecase.FridayDutyRepositoryItf
	now    func() time.Time
}

func NewFridayDutyUsecase(duties usecase.FridayDutyRepositoryItf) *UseCase {
	return &UseCase{duties: duties, now: time.Now}
}

// List returns the petugas Jumat in the given range
func (u *UseCase) List(ctx context.Context, r domain.FridayDutyRange) ([]domain.FridayDuty, error) {
	if err := r.ValidateAndNormalize(); err != nil {
		return nil, err
	}
	return u.duties.FindBetween(ctx, r.From, r.To)
}

// ListUpcoming returns the petugas Jumat of the coming UpcomingDays, for the TV
func (u *UseCase) ListUpcoming(ctx context.Context) ([]domain.FridayDuty, error) {
	to := u.now().UTC().AddDate(0, 0, UpcomingDays).Format(dateLayout)
	return u.duties.FindBetween(ctx, u.earliestToday(), to)
}

// Save validates req and sets the petugas of that Friday. A Friday already past cannot be set.
func (u *UseCase) Save(ctx context.Context, req domain.SaveFridayDutyRequest) (domain.FridayDuty, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.FridayDuty{}, err
	}
	if earliest := u.earliestToday(); req.Date < earliest {
		return domain.FridayDuty{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("date %s is before %s", req.Date, earliest), "tanggal sudah lewat")
	}

	return u.duties.Save(ctx, req.ToFridayDuty())
}

// Delete removes the petugas of a Friday
func (u *UseCase) Delete(ctx context.Context, date string) error {
	return u.duties.Delete(ctx, date)
}

// earliestToday is the earliest date that can be today for the mosque. Dates are local to the
// mosque but the server runs on UTC; every Indonesian zone (UTC+7..+9) is ahead of UTC, so the
// mosque's today is the UTC today or tomorrow, never earlier.
func (u *UseCase) earliestToday() string {
	return u.now().UTC().Format(dateLayout)
}
