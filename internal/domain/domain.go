package domain

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"time"
)

const FacilityID = "FACTORY-01"

var AsOf = time.Date(2026, 10, 27, 6, 15, 0, 0, time.FixedZone("Asia/Bangkok", 7*3600))

func ID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

type SourceRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
type Case struct {
	ID          string            `json:"id"`
	RuleID      string            `json:"rule_id"`
	EntityID    string            `json:"entity_id"`
	ShiftID     string            `json:"shift_id"`
	Department  string            `json:"department"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	DataQuality string            `json:"data_quality"`
	Expected    map[string]string `json:"expected"`
	Observed    map[string]string `json:"observed"`
	Explanation string            `json:"explanation"`
	Sources     []SourceRef       `json:"sources"`
	DetectedAt  time.Time         `json:"detected_at"`
}
type Evidence struct {
	ID           string            `json:"id"`
	CaseID       string            `json:"case_id"`
	RunID        string            `json:"run_id"`
	ShiftID      string            `json:"shift_id"`
	Tool         string            `json:"tool"`
	Domain       string            `json:"domain"`
	Facts        map[string]string `json:"facts"`
	Sources      []SourceRef       `json:"sources"`
	Completeness string            `json:"completeness"`
	Notes        []string          `json:"notes"`
	Hash         string            `json:"hash"`
	RetrievedAt  time.Time         `json:"retrieved_at"`
}
type Hypothesis struct {
	Statement       string   `json:"statement"`
	EvidenceIDs     []string `json:"evidence_ids"`
	Confidence      string   `json:"confidence"`
	CounterEvidence []string `json:"counter_evidence"`
}
type ProposedAction struct {
	Title       string   `json:"title"`
	Rationale   string   `json:"rationale"`
	EvidenceIDs []string `json:"evidence_ids"`
	Department  string   `json:"department"`
}
type Report struct {
	Summary                string           `json:"summary"`
	ObservationEvidenceIDs []string         `json:"observation_evidence_ids"`
	Hypotheses             []Hypothesis     `json:"hypotheses"`
	InformationGaps        []string         `json:"information_gaps"`
	Recommendations        []ProposedAction `json:"recommendations"`
}
type Run struct {
	ID           string          `json:"id"`
	CaseID       string          `json:"case_id"`
	Status       string          `json:"status"`
	Mode         string          `json:"mode"`
	Instruction  string          `json:"instruction"`
	Error        string          `json:"error,omitempty"`
	Report       *Report         `json:"report,omitempty"`
	Checkpoint   json.RawMessage `json:"-"`
	CreatedAt    time.Time       `json:"created_at"`
	StartedAt    *time.Time      `json:"started_at"`
	FinishedAt   *time.Time      `json:"finished_at"`
	ModelRounds  int             `json:"model_rounds"`
	ToolCalls    int             `json:"tool_calls"`
	InputTokens  int64           `json:"input_tokens"`
	OutputTokens int64           `json:"output_tokens"`
	CostMicros   int64           `json:"cost_micros"`
}
type Recommendation struct {
	ID     string `json:"id"`
	RunID  string `json:"run_id"`
	CaseID string `json:"case_id"`
	ProposedAction
	Decision string `json:"decision,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Reviewer string `json:"reviewer,omitempty"`
}
type Event struct {
	ID        int64     `json:"id"`
	RunID     string    `json:"run_id"`
	Kind      string    `json:"kind"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func Terminal(s string) bool { return s != "queued" && s != "running" }
func BPS(observed, expected int64) (int64, error) {
	if expected <= 0 {
		return 0, errors.New("baseline harus positif")
	}
	n := new(big.Int).Sub(big.NewInt(observed), big.NewInt(expected))
	n.Mul(n, big.NewInt(10000))
	n.Quo(n, big.NewInt(expected))
	if !n.IsInt64() {
		return 0, errors.New("variance overflow")
	}
	return n.Int64(), nil
}
func AdjustedMedian(costs, targets []int64, target int64) (int64, error) {
	if len(costs) < 4 || len(costs) != len(targets) || target <= 0 {
		return 0, errors.New("minimal empat pembanding lengkap")
	}
	r := make([]*big.Rat, len(costs))
	for i, c := range costs {
		if targets[i] <= 0 || c < 0 {
			return 0, errors.New("pembanding tidak valid")
		}
		r[i] = new(big.Rat).SetFrac(new(big.Int).Mul(big.NewInt(c), big.NewInt(target)), big.NewInt(targets[i]))
	}
	sort.Slice(r, func(i, j int) bool { return r[i].Cmp(r[j]) < 0 })
	m := r[len(r)/2]
	if len(r)%2 == 0 {
		m = new(big.Rat).Quo(new(big.Rat).Add(r[len(r)/2-1], m), big.NewRat(2, 1))
	}
	q, rem := new(big.Int), new(big.Int)
	q.QuoRem(m.Num(), m.Denom(), rem)
	if new(big.Int).Mul(rem, big.NewInt(2)).Cmp(m.Denom()) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	if !q.IsInt64() {
		return 0, errors.New("money overflow")
	}
	return q.Int64(), nil
}

type Interval struct{ Start, End time.Time }

func UnionMinutes(intervals []Interval, start, end time.Time) int64 {
	valid := []Interval{}
	for _, v := range intervals {
		if v.Start.Before(start) {
			v.Start = start
		}
		if v.End.After(end) {
			v.End = end
		}
		if v.End.After(v.Start) {
			valid = append(valid, v)
		}
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].Start.Before(valid[j].Start) })
	if len(valid) == 0 {
		return 0
	}
	a, b := valid[0].Start, valid[0].End
	var total time.Duration
	for _, v := range valid[1:] {
		if !v.Start.After(b) {
			if v.End.After(b) {
				b = v.End
			}
		} else {
			total += b.Sub(a)
			a, b = v.Start, v.End
		}
	}
	total += b.Sub(a)
	return int64(total / time.Minute)
}
func LeaveCovers(leaveStart, leaveEnd, shiftStart, shiftEnd time.Time, approved bool) bool {
	return approved && !leaveStart.After(shiftStart) && !leaveEnd.Before(shiftEnd)
}
