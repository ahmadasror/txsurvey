package tests

import (
	"context"
	"testing"

	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/repository"
)

// TestUserCap_BlocksNewSignupsButLetsExistingIn verifies the creator-count cap:
// new accounts are refused once the table is full, while returning users (same
// google_sub) always pass.
func TestUserCap_BlocksNewSignupsButLetsExistingIn(t *testing.T) {
	h := newHarness(t) // seeds 1 user + applies the *_test DB safety guard
	ctx := context.Background()
	if _, err := h.pool.Exec(ctx, `TRUNCATE users CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	repo := repository.NewUserRepo(h.pool)
	prof := func(n string) model.GoogleProfile {
		return model.GoogleProfile{Sub: "sub-" + n, Email: n + "@cap.test", Name: n}
	}

	const cap = 2
	if u, capped, err := repo.UpsertByGoogleSubCapped(ctx, prof("a"), cap); err != nil || capped || u == nil {
		t.Fatalf("user a (1/2) should be created: capped=%v err=%v", capped, err)
	}
	if _, capped, err := repo.UpsertByGoogleSubCapped(ctx, prof("b"), cap); err != nil || capped {
		t.Fatalf("user b (2/2) should be created: capped=%v err=%v", capped, err)
	}
	// Third NEW account is over the cap → refused.
	if u, capped, err := repo.UpsertByGoogleSubCapped(ctx, prof("c"), cap); err != nil || !capped || u != nil {
		t.Fatalf("user c (3rd) must be capped: u=%v capped=%v err=%v", u, capped, err)
	}
	// A returning user (existing google_sub) still signs in at/over the cap.
	if _, capped, err := repo.UpsertByGoogleSubCapped(ctx, prof("a"), cap); err != nil || capped {
		t.Fatalf("existing user a must pass at cap: capped=%v err=%v", capped, err)
	}
}
