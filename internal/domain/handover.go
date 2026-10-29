package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type HandoverShift struct {
	ID              string    `json:"id"`
	LineID          string    `json:"line_id"`
	LineName        string    `json:"line_name"`
	OperationalDate string    `json:"operational_date"`
	Code            string    `json:"code"`
	Product         string    `json:"product"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	Working         bool      `json:"working"`
}

type HandoverSection struct {
	Key     string            `json:"key"`
	Title   string            `json:"title"`
	Quality string            `json:"quality"`
	Facts   map[string]string `json:"facts"`
	Notes   []string          `json:"notes"`
}

type HandoverEvidenceRef struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
}

type HandoverCase struct {
	CaseID          string                `json:"case_id"`
	RuleID          string                `json:"rule_id"`
	Severity        string                `json:"severity"`
	Status          string                `json:"status"`
	DataQuality     string                `json:"data_quality"`
	Explanation     string                `json:"explanation"`
	RunID           string                `json:"run_id,omitempty"`
	RunStatus       string                `json:"run_status,omitempty"`
	ModelMode       string                `json:"model_mode,omitempty"`
	Synthesis       string                `json:"synthesis,omitempty"`
	Recommendations []Recommendation      `json:"recommendations"`
	Evidence        []HandoverEvidenceRef `json:"evidence"`
}

type HandoverDowntime struct {
	ID          string    `json:"id"`
	MachineID   string    `json:"machine_id"`
	MachineName string    `json:"machine_name"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Reason      string    `json:"reason"`
}

type HandoverSnapshot struct {
	Version        int                `json:"version"`
	FacilityID     string             `json:"facility_id"`
	Timezone       string             `json:"timezone"`
	Synthetic      bool               `json:"synthetic"`
	DataAsOf       time.Time          `json:"data_as_of"`
	Shift          HandoverShift      `json:"shift"`
	NextShift      *HandoverShift     `json:"next_shift"`
	Sections       []HandoverSection  `json:"sections"`
	Cases          []HandoverCase     `json:"cases"`
	DowntimeEvents []HandoverDowntime `json:"downtime_events"`
	Sources        []SourceRef        `json:"sources"`
	Notes          []string           `json:"notes"`
}

type Handover struct {
	ID        string           `json:"id"`
	ShiftID   string           `json:"shift_id"`
	CreatedAt time.Time        `json:"created_at"`
	CreatedBy string           `json:"created_by"`
	Hash      string           `json:"hash"`
	Snapshot  HandoverSnapshot `json:"snapshot"`
}

type HandoverSummary struct {
	ID        string    `json:"id"`
	ShiftID   string    `json:"shift_id"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

func HandoverHash(snapshot HandoverSnapshot) (string, error) {
	b, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
