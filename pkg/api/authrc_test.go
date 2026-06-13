package api

import (
	"reflect"
	"testing"
)

func TestRecordPageVisit(t *testing.T) {
	t.Run("appends in visit order", func(t *testing.T) {
		rc := &AuthRC{}
		rc.RecordPageVisit("a")
		rc.RecordPageVisit("b")
		rc.RecordPageVisit("c")
		if got, want := rc.RecentPages, []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("RecentPages = %v, want %v", got, want)
		}
	})

	t.Run("revisiting moves slug to end without duplicating", func(t *testing.T) {
		rc := &AuthRC{RecentPages: []string{"a", "b", "c"}}
		rc.RecordPageVisit("a")
		if got, want := rc.RecentPages, []string{"b", "c", "a"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("RecentPages = %v, want %v", got, want)
		}
	})

	t.Run("caps at maxRecentPages dropping the oldest", func(t *testing.T) {
		rc := &AuthRC{}
		for _, s := range []string{"a", "b", "c", "d", "e", "f"} {
			rc.RecordPageVisit(s)
		}
		if got, want := rc.RecentPages, []string{"b", "c", "d", "e", "f"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("RecentPages = %v, want %v", got, want)
		}
		if len(rc.RecentPages) != maxRecentPages {
			t.Fatalf("len(RecentPages) = %d, want %d", len(rc.RecentPages), maxRecentPages)
		}
	})

	t.Run("empty slug is a no-op", func(t *testing.T) {
		rc := &AuthRC{RecentPages: []string{"a"}}
		rc.RecordPageVisit("")
		if got, want := rc.RecentPages, []string{"a"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("RecentPages = %v, want %v", got, want)
		}
	})
}
