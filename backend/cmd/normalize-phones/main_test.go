package main

import "testing"

func TestBuildPlan(t *testing.T) {
	p := buildPlan([]row{
		{1, "3834123456"},     // changes
		{2, "+5493834123457"}, // already canonical
		{3, "abc"},            // invalid
		{4, "-"},              // maintenance placeholder, ignored
	})

	if len(p.changes) != 1 || p.changes[0] != (change{ID: 1, Old: "3834123456", New: "+5493834123456"}) {
		t.Errorf("unexpected changes: %+v", p.changes)
	}
	if len(p.invalid) != 1 || p.invalid[0].ID != 3 {
		t.Errorf("unexpected invalid rows: %+v", p.invalid)
	}
}

func TestSplitCollisions(t *testing.T) {
	users := []row{
		{1, "3834123456"},     // -> +5493834123456, collides with user 2
		{2, "+5493834123456"}, // already canonical
		{3, "0383 15-412-9999"},
		{4, "383 412 9999"}, // both 3 and 4 -> +5493834129999
		{5, "3834111111"},   // safe
	}
	changes := buildPlan(users).changes

	safe, collisions := splitCollisions(users, changes)

	if len(safe) != 1 || safe[0].ID != 5 {
		t.Errorf("unexpected safe changes: %+v", safe)
	}
	if len(collisions) != 3 {
		t.Errorf("got %d collisions, want 3 (users 1, 3, 4): %+v", len(collisions), collisions)
	}
}
