package store

import (
	"context"
	"encoding/json"
	"fmt"
	"opsmind/internal/domain"
)

func (s *Store) SaveHandover(ctx context.Context, report domain.Handover) error {
	hash, err := domain.HandoverHash(report.Snapshot)
	if err != nil {
		return err
	}
	if hash != report.Hash || report.ShiftID != report.Snapshot.Shift.ID {
		return fmt.Errorf("snapshot laporan tidak valid")
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO shift_handovers(id,shift_id,created_at,created_by,snapshot,snapshot_hash)
 SELECT $1,sh.id,$3,$4,$5,$6 FROM shift_instances sh JOIN production_lines l ON l.id=sh.line_id
 WHERE sh.id=$2 AND l.facility_id=$7`, report.ID, report.ShiftID, report.CreatedAt, report.CreatedBy, JSON(report.Snapshot), report.Hash, domain.FacilityID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_events(actor,action,details) VALUES($1,'handover_created',$2)`, report.CreatedBy, JSON(map[string]string{"handover_id": report.ID, "shift_id": report.ShiftID, "hash": report.Hash}))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) Handover(ctx context.Context, id string) (domain.Handover, error) {
	var report domain.Handover
	var raw []byte
	err := s.readDB().QueryRow(ctx, `SELECT h.id,h.shift_id,h.created_at,h.created_by,h.snapshot,h.snapshot_hash FROM shift_handovers h JOIN shift_instances sh ON sh.id=h.shift_id JOIN production_lines l ON l.id=sh.line_id WHERE h.id=$1 AND l.facility_id=$2`, id, domain.FacilityID).Scan(&report.ID, &report.ShiftID, &report.CreatedAt, &report.CreatedBy, &raw, &report.Hash)
	if err != nil {
		return report, normalize(err)
	}
	if err = json.Unmarshal(raw, &report.Snapshot); err != nil {
		return report, err
	}
	hash, err := domain.HandoverHash(report.Snapshot)
	if err == nil && (hash != report.Hash || report.ShiftID != report.Snapshot.Shift.ID) {
		err = fmt.Errorf("integritas snapshot laporan tidak valid")
	}
	return report, err
}

func (s *Store) Handovers(ctx context.Context, shift string, limit, offset int) ([]domain.HandoverSummary, int, error) {
	from := ` FROM shift_handovers h JOIN shift_instances sh ON sh.id=h.shift_id JOIN production_lines l ON l.id=sh.line_id WHERE l.facility_id=$1 AND ($2='' OR h.shift_id=$2)`
	var total int
	if err := s.readDB().QueryRow(ctx, "SELECT count(*)"+from, domain.FacilityID, shift).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.readDB().Query(ctx, "SELECT h.id,h.shift_id,h.created_at,h.created_by"+from+" ORDER BY h.created_at DESC,h.id DESC LIMIT $3 OFFSET $4", domain.FacilityID, shift, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.HandoverSummary{}
	for rows.Next() {
		var v domain.HandoverSummary
		if err = rows.Scan(&v.ID, &v.ShiftID, &v.CreatedAt, &v.CreatedBy); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}
