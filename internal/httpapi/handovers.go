package httpapi

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"opsmind/internal/domain"
	"opsmind/internal/handover"
	"opsmind/internal/store"
	"strconv"
	"strings"
	"time"
)

func handoverPagination(r *http.Request) (int, int, error) {
	limit, offset := 30, 0
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 100 {
			return 0, 0, errors.New("limit harus 1 sampai 100")
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 {
			return 0, 0, errors.New("offset harus nol atau positif")
		}
	}
	return limit, offset, nil
}

func (a *Server) handoverShifts(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := handoverPagination(r)
	if err != nil {
		a.fail(w, r, 400, "validation", err.Error())
		return
	}
	line, date := r.URL.Query().Get("line_id"), r.URL.Query().Get("date")
	if len(line) > 80 {
		a.fail(w, r, 400, "validation", "line_id tidak valid.")
		return
	}
	if date != "" {
		if _, err = time.Parse("2006-01-02", date); err != nil {
			a.fail(w, r, 400, "validation", "Tanggal operasional harus YYYY-MM-DD.")
			return
		}
	}
	items, total, err := a.Store.HandoverShifts(r.Context(), line, date, limit, offset, domain.AsOf)
	if err != nil {
		a.dbError(w, r, err)
		return
	}
	lines, err := a.Store.HandoverLines(r.Context())
	if err != nil {
		a.dbError(w, r, err)
		return
	}
	write(w, 200, map[string]any{"items": items, "total": total, "lines": lines, "as_of": domain.AsOf})
}
func (a *Server) handovers(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := handoverPagination(r)
	if err != nil {
		a.fail(w, r, 400, "validation", err.Error())
		return
	}
	shift := r.URL.Query().Get("shift_id")
	if len(shift) > 80 {
		a.fail(w, r, 400, "validation", "shift_id tidak valid.")
		return
	}
	items, total, err := a.Store.Handovers(r.Context(), shift, limit, offset)
	if err != nil {
		a.dbError(w, r, err)
		return
	}
	write(w, 200, map[string]any{"items": items, "total": total})
}
func (a *Server) createHandover(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ShiftID string `json:"shift_id"`
	}
	if decode(w, r, &body) != nil || strings.TrimSpace(body.ShiftID) == "" || len(body.ShiftID) > 80 {
		a.fail(w, r, 400, "validation", "Pilih shift yang sudah selesai.")
		return
	}
	report, err := (handover.Service{Store: a.Store}).Create(r.Context(), body.ShiftID, a.Config.Username)
	if errors.Is(err, store.ErrShiftNotFinished) {
		a.fail(w, r, 409, "shift_not_finished", "Shift belum selesai pada waktu evaluasi data.")
		return
	}
	if err != nil {
		a.dbError(w, r, err)
		return
	}
	write(w, 201, report)
}
func (a *Server) handoverDetail(w http.ResponseWriter, r *http.Request) {
	report, err := a.Store.Handover(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.dbError(w, r, err)
		return
	}
	write(w, 200, report)
}
