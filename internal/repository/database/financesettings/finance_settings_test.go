package financesettings

import (
	"context"
	"errors"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/dbtest"
)

func TestTVSetting(t *testing.T) {
	for driver, db := range dbtest.Open(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			r := New(db.Client)
			if _, err := r.GetTVSetting(ctx); !errors.Is(err, domain.ErrFinanceTVSettingNotFound) {
				t.Fatalf("before save: %v", err)
			}
			for _, p := range []domain.TVPeriod{domain.TVPeriodWeek, domain.TVPeriodToday} {
				if err := r.UpdateTVSetting(ctx, domain.FinanceTVSetting{Period: p}, db.UserID); err != nil {
					t.Fatal(err)
				}
				if got, err := r.GetTVSetting(ctx); err != nil || got.Period != p {
					t.Errorf("got %+v, %v; want %s", got, err, p)
				}
			}
		})
	}
}
