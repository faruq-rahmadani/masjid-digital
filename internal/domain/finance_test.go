package domain

import (
	"testing"
	"time"
)

func TestPocketKindRules(t *testing.T) {
	tests := []struct {
		kind                           PocketKind
		builtIn, zakat, moves, consent bool
		scope                          CategoryScope
	}{
		{PocketKindUmum, true, false, true, false, CategoryScopeUmumTerikat},
		{PocketKindTerikat, false, false, true, true, CategoryScopeUmumTerikat},
		{PocketKindZakatMal, true, true, false, false, CategoryScopeZakat},
		{PocketKindZakatFitrah, true, true, false, false, CategoryScopeZakat},
		{PocketKindFidyah, true, true, false, false, CategoryScopeFidyah},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			k := tt.kind
			if k.IsBuiltIn() != tt.builtIn || k.IsZakat() != tt.zakat || k.CanMoveFunds() != tt.moves ||
				k.NeedsDonorConsent() != tt.consent || k.CategoryScope() != tt.scope {
				t.Errorf("rules of %s = builtIn %v zakat %v moves %v consent %v scope %s", k,
					k.IsBuiltIn(), k.IsZakat(), k.CanMoveFunds(), k.NeedsDonorConsent(), k.CategoryScope())
			}
		})
	}
	if PocketKind("wakaf").IsValid() || PocketKind("wakaf").IsBuiltIn() {
		t.Error("an unknown kind must not be valid")
	}
}

func TestCategoryFitsPocket(t *testing.T) {
	infaq := Category{Scope: CategoryScopeUmumTerikat}
	zakatMal := Category{Scope: CategoryScopeZakat, Code: new(CategoryCodeZakatMal)}
	fakir := Category{Scope: CategoryScopeZakat, Code: new("asnaf_fakir")}
	fidyahFakir := Category{Scope: CategoryScopeFidyah, Code: new("fidyah_fakir")}
	pindah := Category{Scope: CategoryScopeAll, Code: new(CategoryCodePindahOut)}

	tests := []struct {
		name string
		c    Category
		kind PocketKind
		want bool
	}{
		{"infaq into Kas Umum", infaq, PocketKindUmum, true},
		{"infaq into a terikat kas", infaq, PocketKindTerikat, true},
		{"infaq into zakat", infaq, PocketKindZakatMal, false},
		{"zakat mal into Zakat Mal", zakatMal, PocketKindZakatMal, true},
		{"zakat mal into Zakat Fitrah", zakatMal, PocketKindZakatFitrah, false},
		{"zakat mal into Kas Umum", zakatMal, PocketKindUmum, false},
		{"asnaf fakir from Zakat Fitrah", fakir, PocketKindZakatFitrah, true},
		{"zakat asnaf from Fidyah", fakir, PocketKindFidyah, false},
		{"fidyah fakir from Fidyah", fidyahFakir, PocketKindFidyah, true},
		{"pindah dana anywhere", pindah, PocketKindZakatMal, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.FitsPocket(tt.kind); got != tt.want {
				t.Errorf("FitsPocket(%s) = %v, want %v", tt.kind, got, tt.want)
			}
		})
	}
	if !pindah.IsPindah() || !pindah.IsSystem() || infaq.IsPindah() || infaq.IsSystem() {
		t.Error("system and pindah flags wrong")
	}
}

func TestTransactionHelpers(t *testing.T) {
	created := time.Date(2026, 11, 6, 3, 0, 0, 0, time.UTC)
	tx := Transaction{Kind: CategoryTypeOut, Amount: 300_000, CreatedAt: created, UpdatedAt: created}
	if tx.Signed() != -300_000 || tx.IsEdited() {
		t.Errorf("out: signed %d, edited %v", tx.Signed(), tx.IsEdited())
	}
	tx.Kind, tx.UpdatedAt = CategoryTypeIn, created.Add(time.Hour)
	if tx.Signed() != 300_000 || !tx.IsEdited() {
		t.Errorf("in, edited later: signed %d, edited %v", tx.Signed(), tx.IsEdited())
	}
}
