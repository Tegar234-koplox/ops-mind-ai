package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"opsmind/internal/config"
	"opsmind/internal/detection"
	"opsmind/internal/domain"
	"opsmind/internal/store"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Store        *store.Store
	Config       config.Config
	Logger       *slog.Logger
	router       http.Handler
	passwordHash []byte
	limiter      *limiter
}
type bucket struct {
	Count int
	Until time.Time
}
type limiter struct {
	sync.Mutex
	Buckets map[string]bucket
}

func (l *limiter) allow(key string, max int) bool {
	l.Lock()
	defer l.Unlock()
	now := time.Now()
	if len(l.Buckets) > 10000 {
		for k, v := range l.Buckets {
			if v.Until.Before(now) {
				delete(l.Buckets, k)
			}
		}
		if len(l.Buckets) > 10000 {
			return false
		}
	}
	v := l.Buckets[key]
	if now.After(v.Until) {
		v = bucket{Until: now.Add(time.Minute)}
	}
	v.Count++
	l.Buckets[key] = v
	return v.Count <= max
}
func New(s *store.Store, c config.Config, logger *slog.Logger) (*Server, error) {
	a := &Server{Store: s, Config: c, Logger: logger, limiter: &limiter{Buckets: map[string]bucket{}}}
	if c.PasswordHash != "" {
		a.passwordHash = []byte(c.PasswordHash)
		if _, e := bcrypt.Cost(a.passwordHash); e != nil {
			return nil, errors.New("OPERATOR_PASSWORD_HASH bcrypt tidak valid")
		}
	} else {
		var e error
		a.passwordHash, e = bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
		if e != nil {
			return nil, e
		}
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, a.headers)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	r.Post("/api/v1/auth/login", a.login)
	r.Group(func(r chi.Router) {
		r.Use(a.authenticate)
		r.Get("/api/v1/auth/session", func(w http.ResponseWriter, r *http.Request) {
			write(w, 200, map[string]string{"username": c.Username, "model_mode": c.Mode})
		})
		r.Post("/api/v1/auth/logout", a.logout)
		r.Route("/api/v1", func(r chi.Router) {
			r.Get("/overview", a.overview)
			r.Get("/cases", a.cases)
			r.Get("/cases/{id}", a.caseDetail)
			r.Get("/investigations/{id}", a.runDetail)
			r.Get("/investigations/{id}/events", a.events)
			r.Get("/cases/{id}/evidence/{evidenceId}", a.evidence)
			r.Group(func(r chi.Router) {
				r.Use(a.mutationGuard, a.idempotency)
				r.Post("/detections", a.detect)
				r.Post("/cases/{id}/investigations", a.investigate)
				r.Post("/cases/{id}/follow-up-investigations", a.followup)
				r.Post("/investigations/{id}/cancel", a.cancel)
				r.Post("/recommendations/{id}/reviews", a.review)
			})
		})
	})
	r.NotFound(a.static)
	a.router = r
	return a, nil
}
func (a *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { a.router.ServeHTTP(w, r) }
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (a *Server) fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	write(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}, "request_id": middleware.GetReqID(r.Context())})
}
func (a *Server) dbError(w http.ResponseWriter, r *http.Request, e error) {
	switch {
	case errors.Is(e, store.ErrNotFound):
		a.fail(w, r, 404, "not_found", "Data tidak ditemukan.")
	case errors.Is(e, store.ErrConflict):
		a.fail(w, r, 409, "conflict", "State sudah berubah; muat ulang sebelum melanjutkan.")
	default:
		if a.Logger != nil {
			a.Logger.Error("request failed", "request_id", middleware.GetReqID(r.Context()), "path", r.URL.Path)
		}
		a.fail(w, r, 503, "dependency_unavailable", "Database atau dependensi belum siap.")
	}
}
func (a *Server) headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
func (a *Server) originOK(r *http.Request) bool { return r.Header.Get("Origin") == a.Config.Origin }

type session struct {
	User    string `json:"user"`
	Expires int64  `json:"expires"`
	Nonce   string `json:"nonce"`
}

func (a *Server) token() string {
	b := base64.RawURLEncoding.EncodeToString(store.JSON(session{a.Config.Username, time.Now().Add(8 * time.Hour).Unix(), domain.ID()}))
	mac := hmac.New(sha256.New, []byte(a.Config.SessionSecret))
	mac.Write([]byte(b))
	return b + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func (a *Server) validToken(v string) bool {
	p := strings.Split(v, ".")
	if len(p) != 2 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(a.Config.SessionSecret))
	mac.Write([]byte(p[0]))
	sig, e := base64.RawURLEncoding.DecodeString(p[1])
	if e != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return false
	}
	raw, e := base64.RawURLEncoding.DecodeString(p[0])
	if e != nil {
		return false
	}
	var s session
	if json.Unmarshal(raw, &s) != nil {
		return false
	}
	return s.User == a.Config.Username && s.Expires > time.Now().Unix() && s.Expires <= time.Now().Add(8*time.Hour+time.Minute).Unix()
}
func (a *Server) cookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: "opsmind_session", Value: value, Path: "/", HttpOnly: true, Secure: a.Config.CookieSecure, SameSite: http.SameSiteStrictMode, MaxAge: maxAge})
}
func (a *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, e := r.Cookie("opsmind_session")
		if e != nil || !a.validToken(cookie.Value) {
			a.fail(w, r, 401, "unauthenticated", "Silakan masuk sebagai operator.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16384)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errors.New("body tidak valid")
	}
	return nil
}
func (a *Server) login(w http.ResponseWriter, r *http.Request) {
	if !a.originOK(r) {
		a.fail(w, r, 403, "origin_denied", "Origin tidak diizinkan.")
		return
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if !a.limiter.allow("login:"+ip, 5) {
		a.fail(w, r, 429, "rate_limited", "Terlalu banyak percobaan; tunggu satu menit.")
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if decode(w, r, &body) != nil {
		a.fail(w, r, 400, "validation", "Isi username dan password.")
		return
	}
	pwErr := bcrypt.CompareHashAndPassword(a.passwordHash, []byte(body.Password))
	if subtle.ConstantTimeCompare([]byte(body.Username), []byte(a.Config.Username)) != 1 || pwErr != nil {
		a.fail(w, r, 401, "invalid_credentials", "Kredensial tidak cocok.")
		return
	}
	a.cookie(w, a.token(), 8*3600)
	write(w, 200, map[string]string{"username": a.Config.Username})
}
func (a *Server) logout(w http.ResponseWriter, r *http.Request) {
	if !a.originOK(r) {
		a.fail(w, r, 403, "origin_denied", "Origin tidak diizinkan.")
		return
	}
	a.cookie(w, "", -1)
	write(w, 200, map[string]bool{"ok": true})
}
func (a *Server) mutationGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.originOK(r) {
			a.fail(w, r, 403, "origin_denied", "Origin tidak diizinkan.")
			return
		}
		if !a.limiter.allow("mutation:"+a.Config.Username, 30) {
			a.fail(w, r, 429, "rate_limited", "Batas operasi tercapai; tunggu satu menit.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type recorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *recorder) Header() http.Header { return r.header }
func (r *recorder) WriteHeader(n int)   { r.status = n }
func (r *recorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = 200
	}
	return r.body.Write(b)
}
func (a *Server) idempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if len(key) < 8 || len(key) > 128 {
			a.fail(w, r, 400, "idempotency_required", "Idempotency-Key harus 8â€“128 karakter.")
			return
		}
		raw, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 16384))
		if e != nil {
			a.fail(w, r, 400, "validation", "Payload terlalu besar.")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(raw))
		sum := sha256.Sum256(raw)
		hash := hex.EncodeToString(sum[:])
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		tag, e := a.Store.Pool.Exec(ctx, "INSERT INTO api_idempotency(actor,key,path,body_hash) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING", a.Config.Username, key, r.URL.Path, hash)
		if e != nil {
			a.dbError(w, r, e)
			return
		}
		if tag.RowsAffected() == 0 {
			var p, h string
			var status int
			var response []byte
			if e = a.Store.Pool.QueryRow(ctx, "SELECT path,body_hash,status,response FROM api_idempotency WHERE actor=$1 AND key=$2", a.Config.Username, key).Scan(&p, &h, &status, &response); e != nil {
				a.dbError(w, r, e)
				return
			}
			if p != r.URL.Path || h != hash {
				a.fail(w, r, 409, "idempotency_conflict", "Kunci pernah dipakai untuk operasi berbeda.")
				return
			}
			if status == 0 {
				w.Header().Set("Retry-After", "2")
				a.fail(w, r, 409, "idempotency_pending", "Hasil operasi belum terekam; periksa status sebelum mengirim operasi baru.")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Idempotency-Replayed", "true")
			w.WriteHeader(status)
			_, _ = w.Write(response)
			return
		}
		rec := &recorder{header: make(http.Header)}
		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = 200
		}
		saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer saveCancel()
		if _, e = a.Store.Pool.Exec(saveCtx, "UPDATE api_idempotency SET status=$3,response=$4 WHERE actor=$1 AND key=$2", a.Config.Username, key, rec.status, rec.body.Bytes()); e != nil {
			a.dbError(w, r, e)
			return
		}
		for k, v := range rec.header {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.status)
		_, _ = w.Write(rec.body.Bytes())
	})
}
func (a *Server) overview(w http.ResponseWriter, r *http.Request) {
	budget, e := a.Store.Budget(r.Context())
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	var cases, active, review, employees int
	if e = a.Store.Pool.QueryRow(r.Context(), "SELECT (SELECT count(*) FROM anomaly_cases),(SELECT count(*) FROM investigation_runs WHERE status IN('queued','running')),(SELECT count(*) FROM recommendations x LEFT JOIN review_actions v ON v.recommendation_id=x.id WHERE v.id IS NULL AND x.run_id=(SELECT id FROM investigation_runs WHERE case_id=x.case_id ORDER BY created_at DESC,id DESC LIMIT 1)),(SELECT count(*) FROM employees)").Scan(&cases, &active, &review, &employees); e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 200, map[string]any{"cases": cases, "active_runs": active, "pending_reviews": review, "employees": employees, "facility": "FACTORY-01", "model_mode": a.Config.Mode, "model_name": a.Config.Model, "synthetic": true, "seed": 20261027, "as_of": domain.AsOf, "budget_micros": budget, "credit_limit_micros": a.Config.CreditLimit, "phase": a.Config.Phase})
}
func (a *Server) cases(w http.ResponseWriter, r *http.Request) {
	limit, offset := 25, 0
	var e error
	if q := r.URL.Query().Get("limit"); q != "" {
		limit, e = strconv.Atoi(q)
		if e != nil || limit < 1 || limit > 100 {
			a.fail(w, r, 400, "validation", "limit harus 1â€“100.")
			return
		}
	}
	if q := r.URL.Query().Get("offset"); q != "" {
		offset, e = strconv.Atoi(q)
		if e != nil || offset < 0 {
			a.fail(w, r, 400, "validation", "offset tidak valid.")
			return
		}
	}
	v, total, e := a.Store.Cases(r.Context(), limit, offset, r.URL.Query().Get("status"), r.URL.Query().Get("severity"))
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 200, map[string]any{"items": v, "total": total})
}
func (a *Server) caseDetail(w http.ResponseWriter, r *http.Request) {
	c, e := a.Store.Case(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	var run *domain.Run
	latest, e := a.Store.LatestRun(r.Context(), c.ID)
	if e == nil {
		run = &latest
	} else if !errors.Is(e, store.ErrNotFound) {
		a.dbError(w, r, e)
		return
	}
	recs, e := a.Store.Recommendations(r.Context(), c.ID)
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	audit, e := a.Store.Audit(r.Context(), c.ID)
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 200, map[string]any{"case": c, "latest_run": run, "recommendations": recs, "audit": audit})
}
func (a *Server) runDetail(w http.ResponseWriter, r *http.Request) {
	run, e := a.Store.Run(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	ev, e := a.Store.EvidenceList(r.Context(), run.ID)
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	events, e := a.Store.Events(r.Context(), run.ID, 0)
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 200, map[string]any{"run": run, "evidence": ev, "events": events})
}
func (a *Server) evidence(w http.ResponseWriter, r *http.Request) {
	v, e := a.Store.Evidence(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "evidenceId"))
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 200, v)
}
func (a *Server) detect(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ShiftID string `json:"shift_id"`
	}
	if decode(w, r, &body) != nil {
		a.fail(w, r, 400, "validation", "Body JSON tidak valid.")
		return
	}
	engine := detection.Engine{Store: a.Store, Config: a.Config}
	v, e := engine.Execute(r.Context(), body.ShiftID, domain.AsOf)
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 200, map[string]any{"items": v, "as_of": domain.AsOf})
}
func (a *Server) investigate(w http.ResponseWriter, r *http.Request) { a.start(w, r, false) }
func (a *Server) followup(w http.ResponseWriter, r *http.Request)    { a.start(w, r, true) }
func (a *Server) start(w http.ResponseWriter, r *http.Request, followup bool) {
	var body struct {
		Instruction string `json:"instruction"`
	}
	if decode(w, r, &body) != nil || len(body.Instruction) > 2000 || (followup && strings.TrimSpace(body.Instruction) == "") {
		a.fail(w, r, 400, "validation", "Instruksi follow-up wajib diisi, maksimal 2.000 karakter.")
		return
	}
	run, e := a.Store.Enqueue(r.Context(), chi.URLParam(r, "id"), a.Config.Mode, body.Instruction, a.Config.Username)
	if e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 202, run)
}
func (a *Server) cancel(w http.ResponseWriter, r *http.Request) {
	var body struct{}
	if decode(w, r, &body) != nil {
		a.fail(w, r, 400, "validation", "Gunakan body {}.")
		return
	}
	if e := a.Store.Cancel(r.Context(), chi.URLParam(r, "id"), a.Config.Username); e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 200, map[string]bool{"ok": true})
}
func (a *Server) review(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if decode(w, r, &body) != nil || (body.Decision != "approved" && body.Decision != "rejected") || len(body.Reason) > 2000 || (body.Decision == "rejected" && strings.TrimSpace(body.Reason) == "") {
		a.fail(w, r, 400, "validation", "Keputusan harus approved/rejected; penolakan wajib disertai alasan.")
		return
	}
	if e := a.Store.Review(r.Context(), chi.URLParam(r, "id"), body.Decision, body.Reason, a.Config.Username); e != nil {
		a.dbError(w, r, e)
		return
	}
	write(w, 200, map[string]bool{"ok": true})
}
func (a *Server) events(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, e := a.Store.Run(r.Context(), id); e != nil {
		a.dbError(w, r, e)
		return
	}
	after := int64(0)
	cursor := r.Header.Get("Last-Event-ID")
	if cursor == "" {
		cursor = r.URL.Query().Get("after")
	}
	if cursor != "" {
		n, e := strconv.ParseInt(cursor, 10, 64)
		if e != nil || n < 0 {
			a.fail(w, r, 400, "validation", "Cursor tidak valid.")
			return
		}
		after = n
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		a.fail(w, r, 503, "stream_unavailable", "Gunakan polling.")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(200)
	flusher.Flush()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()
	for {
		batch, e := a.Store.Events(r.Context(), id, after)
		if e != nil {
			return
		}
		for _, event := range batch {
			if _, e = fmt.Fprintf(w, "id: %d\nevent: progress\ndata: %s\n\n", event.ID, store.JSON(event)); e != nil {
				return
			}
			after = event.ID
		}
		flusher.Flush()
		run, e := a.Store.Run(r.Context(), id)
		if e != nil {
			return
		}
		if domain.Terminal(run.Status) && len(batch) < 500 {
			return
		}
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		case <-heartbeat.C:
			if _, e = fmt.Fprint(w, ": heartbeat\n\n"); e != nil {
				return
			}
			flusher.Flush()
		}
	}
}
func (a *Server) static(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		a.fail(w, r, 404, "not_found", "Endpoint tidak ditemukan.")
		return
	}
	if r.Method != "GET" && r.Method != "HEAD" {
		w.WriteHeader(405)
		return
	}
	root := a.Config.WebDir
	clean := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	name := filepath.Join(root, filepath.FromSlash(clean))
	info, e := os.Stat(name)
	if e != nil || info.IsDir() {
		if path.Ext(clean) != "" {
			http.NotFound(w, r)
			return
		}
		name = filepath.Join(root, "index.html")
	}
	if _, e = os.Stat(name); e != nil {
		http.Error(w, "Frontend belum dibangun. Jalankan npm run build di web.", 503)
		return
	}
	http.ServeFile(w, r, name)
}
