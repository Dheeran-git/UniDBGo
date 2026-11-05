package federation

import (
	"hash/fnv"
	"testing"
)

func TestFNV64a(t *testing.T) {
	for _, s := range []string{"", "a", "n:42", "s:hello world"} {
		h := fnv.New64a()
		h.Write([]byte(s))
		if got, want := fnv64a(s), h.Sum64(); got != want {
			t.Errorf("fnv64a(%q) = %x, want %x", s, got, want)
		}
	}
}

func TestNormalizeKeyAcrossDrivers(t *testing.T) {
	same := [][]any{
		{int64(42), []byte("42"), "42", float64(42), int32(42), uint8(42)},
		{"alice", []byte("alice")},
		{3.5, "3.5", []byte("3.50")},
	}
	for _, group := range same {
		want, _ := NormalizeKey(group[0])
		for _, v := range group[1:] {
			if got, _ := NormalizeKey(v); got != want {
				t.Errorf("NormalizeKey(%#v) = %q, want %q", v, got, want)
			}
		}
	}
	if _, ok := NormalizeKey(nil); ok {
		t.Error("NULL must not produce a join key")
	}
}

func TestHashJoinMatchesNestedLoop(t *testing.T) {
	users := []Row{
		{"u.id": int64(1), "u.name": "a"},
		{"u.id": int64(2), "u.name": "b"},
		{"u.id": nil, "u.name": "null"},
		{"u.id": int64(3), "u.name": "c"},
	}
	orders := []Row{
		{"o.id": []byte("10"), "o.user_id": []byte("1")},
		{"o.id": []byte("11"), "o.user_id": []byte("1")},
		{"o.id": []byte("12"), "o.user_id": []byte("3")},
		{"o.id": []byte("13"), "o.user_id": nil},
		{"o.id": []byte("14"), "o.user_id": []byte("9")},
	}
	for _, swap := range []bool{false, true} {
		l, r, lk, rk := users, orders, "u.id", "o.user_id"
		if swap {
			l, r, lk, rk = orders, users, "o.user_id", "u.id"
		}
		h := HashJoin(l, r, lk, rk)
		n := NestedLoopJoin(l, r, lk, rk)
		if len(h) != 3 || len(n) != 3 {
			t.Fatalf("swap=%v: hash join %d rows, nested loop %d rows, want 3", swap, len(h), len(n))
		}
		for _, row := range h {
			if _, ok := row["u.name"]; !ok {
				t.Errorf("merged row lacks left columns: %v", row)
			}
		}
	}
}

func TestNormalizeKeyOtherTypes(t *testing.T) {
	type uuid [4]byte
	a, okA := NormalizeKey(uuid{1, 2, 3, 4})
	b, okB := NormalizeKey(uuid{1, 2, 3, 4})
	c, _ := NormalizeKey(uuid{9, 9, 9, 9})
	if !okA || !okB || a != b || a == c {
		t.Fatalf("same-typed keys: %q %q %q", a, b, c)
	}
	if k, _ := NormalizeKey([4]int{1, 2, 3, 4}); k == a {
		t.Fatal("different Go types must not join")
	}
	if _, ok := NormalizeKey(nil); ok {
		t.Fatal("NULL must not join")
	}
}
