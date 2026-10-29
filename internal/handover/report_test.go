package handover

import (
	"encoding/json"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"strings"
	"testing"
)

func TestIncompleteHandoverKeepsUncertaintyAndExactMoney(t *testing.T) {
	data := store.HandoverData{
		AsOf:      domain.AsOf,
		Shift:     domain.HandoverShift{ID: "shift", Working: true},
		Operation: store.Shift{Required: 38, Planned: 0, Actual: 0, Complete: false},
		Staff:     store.Staffing{Scheduled: 38, Present: 31, Leave: 3, Unverified: 4},
		Overtime:  store.Overtime{Observed: 9007199254740993, Complete: false},
		Expenses:  store.ExpenseVariance{Observed: 9007199254740993, PostingClosed: false},
		Cases:     []domain.HandoverCase{}, Sources: []domain.SourceRef{}, DowntimeEvents: []domain.HandoverDowntime{},
	}
	snapshot := Build(data)
	if snapshot.Sections[0].Facts["unverified"] != "4" || snapshot.Sections[0].Quality != "needs_verification" {
		t.Fatal(snapshot.Sections[0])
	}
	if _, ok := snapshot.Sections[1].Facts["shortfall_bps"]; ok {
		t.Fatal("zero target produced a percentage")
	}
	finance := snapshot.Sections[2]
	if finance.Quality != "needs_verification" || finance.Facts["expense_actual_rupiah"] != "9007199254740993" {
		t.Fatal(finance)
	}
	if _, ok := finance.Facts["variance_bps"]; ok {
		t.Fatal("incomplete baseline produced variance")
	}
	if _, ok := finance.Facts["expense_variance_bps"]; ok {
		t.Fatal("unclosed expenses produced variance")
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"expense_actual_rupiah":"9007199254740993"`) {
		t.Fatal("money lost integer precision")
	}
	if !strings.Contains(strings.Join(snapshot.Notes, " "), "bukan bukti") {
		t.Fatal("missing case coverage caveat")
	}
	before, err := domain.HandoverHash(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var roundtrip domain.HandoverSnapshot
	if err = json.Unmarshal(raw, &roundtrip); err != nil {
		t.Fatal(err)
	}
	after, _ := domain.HandoverHash(roundtrip)
	if before != after {
		t.Fatal("snapshot hash changed after JSON round trip")
	}
	roundtrip.Sections[0].Facts["unverified"] = "0"
	tampered, _ := domain.HandoverHash(roundtrip)
	if tampered == before {
		t.Fatal("snapshot hash ignored changed facts")
	}
}
