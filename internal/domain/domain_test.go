package domain

import (
	"math"
	"testing"
	"time"
)

func TestMoneyAndBaseline(t *testing.T) {
	for _, tc := range []struct{ observed, expected, want int64 }{{1380000, 1000000, 3800}, {830, 1000, -1700}, {1200000, 1000000, 2000}, {9007199254740993, 9007199254740993, 0}} {
		got, e := BPS(tc.observed, tc.expected)
		if e != nil || got != tc.want {
			t.Fatalf("%+v got %d %v", tc, got, e)
		}
	}
	if _, e := BPS(1, 0); e == nil {
		t.Fatal("zero baseline accepted")
	}
	if _, e := BPS(math.MaxInt64, 1); e == nil {
		t.Fatal("overflow accepted")
	}
	got, e := AdjustedMedian([]int64{100, 200, 300, 400}, []int64{100, 200, 300, 400}, 1000)
	if e != nil || got != 1000 {
		t.Fatalf("activity adjusted median %d %v", got, e)
	}
	got, e = AdjustedMedian([]int64{1000000, 1000000, 1000000, 1000000}, []int64{1000, 1000, 1000, 1000}, 1000)
	if e != nil || got != 1000000 {
		t.Fatal(got, e)
	}
	if _, e = AdjustedMedian([]int64{1, 2, 3}, []int64{1, 1, 1}, 1); e == nil {
		t.Fatal("insufficient comparators accepted")
	}
}
func TestOvernightUnionAndLeave(t *testing.T) {
	start := time.Date(2026, 10, 26, 22, 0, 0, 0, time.UTC)
	end := start.Add(8 * time.Hour)
	intervals := []Interval{{start.Add(-time.Hour), start.Add(70 * time.Minute)}, {start.Add(time.Hour), start.Add(110 * time.Minute)}, {end, end.Add(time.Hour)}}
	if got := UnionMinutes(intervals, start, end); got != 110 {
		t.Fatalf("union %d", got)
	}
	if !LeaveCovers(start, end, start, end, true) || LeaveCovers(start, end.Add(-time.Second), start, end, true) || LeaveCovers(start, end, start, end, false) {
		t.Fatal("leave reconciliation")
	}
}
