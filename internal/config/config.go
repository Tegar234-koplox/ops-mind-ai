package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DSN, Port, Origin, Username, Password, PasswordHash, SessionSecret, Mode, APIBase, APIKey, Model, Phase, WebDir string
	CookieSecure, SeedIfEmpty                                                                                       bool
	CreditLimit, DevelopmentLimit, DemoLimit, RunLimit                                                              int64
	MaxRounds, MaxTools, MaxRetries                                                                                 int
	RunTimeout, ModelTimeout                                                                                        time.Duration
	OTThreshold, ProdThreshold, DowntimeThreshold, AttendanceGrace                                                  int64
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func integer(k string, d int64) (int64, error) {
	v := env(k, strconv.FormatInt(d, 10))
	n, e := strconv.ParseInt(v, 10, 64)
	if e != nil || n < 0 {
		return 0, fmt.Errorf("%s harus integer nonnegatif", k)
	}
	return n, nil
}
func money(k, d string) (int64, error) {
	s := env(k, d)
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("%s tidak valid", k)
	}
	whole, e := strconv.ParseInt(parts[0], 10, 64)
	if e != nil || whole < 0 || whole > 100000 {
		return 0, fmt.Errorf("%s tidak valid", k)
	}
	f := ""
	if len(parts) == 2 {
		f = parts[1]
	}
	if len(f) > 6 {
		return 0, fmt.Errorf("%s maksimal 6 desimal", k)
	}
	for len(f) < 6 {
		f += "0"
	}
	frac, e := strconv.ParseInt(f, 10, 64)
	if e != nil {
		return 0, e
	}
	return whole*1000000 + frac, nil
}
func Load() (Config, error) {
	c := Config{DSN: os.Getenv("DATABASE_URL"), Port: env("PORT", "8080"), Origin: env("APP_ORIGIN", "http://localhost:8080"), Username: env("OPERATOR_USERNAME", "operator"), Password: os.Getenv("OPERATOR_PASSWORD"), PasswordHash: os.Getenv("OPERATOR_PASSWORD_HASH"), SessionSecret: os.Getenv("SESSION_SECRET"), Mode: env("MODEL_MODE", "meta"), APIBase: env("MODEL_API_BASE_URL", "https://api.meta.ai/v1"), APIKey: os.Getenv("MODEL_API_KEY"), Model: env("MODEL_NAME", "muse-spark-1.3"), Phase: env("BUDGET_PHASE", "development"), WebDir: env("WEB_DIR", "web/dist"), CookieSecure: env("COOKIE_SECURE", "true") == "true", SeedIfEmpty: env("SEED_IF_EMPTY", "false") == "true"}
	if c.DSN == "" {
		return c, errors.New("DATABASE_URL wajib diisi")
	}
	if len(c.SessionSecret) < 32 {
		return c, errors.New("SESSION_SECRET minimal 32 karakter")
	}
	if c.PasswordHash == "" && len(c.Password) < 12 {
		return c, errors.New("OPERATOR_PASSWORD minimal 12 karakter atau isi OPERATOR_PASSWORD_HASH")
	}
	if c.Mode != "meta" && c.Mode != "mock" {
		return c, errors.New("MODEL_MODE harus meta atau mock")
	}
	if c.Mode == "meta" && c.APIKey == "" {
		return c, errors.New("MODEL_API_KEY belum diisi; mock hanya jika dipilih eksplisit")
	}
	u, e := url.Parse(c.APIBase)
	if e != nil || u.Host == "" || (u.Scheme != "https" && u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
		return c, errors.New("MODEL_API_BASE_URL wajib HTTPS")
	}
	o, e := url.Parse(c.Origin)
	if e != nil || o.Host == "" || o.Path != "" || (o.Scheme != "https" && o.Scheme != "http") {
		return c, errors.New("APP_ORIGIN wajib origin tanpa path")
	}
	if o.Scheme == "https" && !c.CookieSecure {
		return c, errors.New("HTTPS memerlukan COOKIE_SECURE=true")
	}
	if c.Phase != "development" && c.Phase != "demo" {
		return c, errors.New("BUDGET_PHASE tidak valid")
	}
	for k, p := range map[string]*int64{"CREDIT_LIMIT_USD": &c.CreditLimit, "DEVELOPMENT_LIMIT_USD": &c.DevelopmentLimit, "DEMO_LIMIT_USD": &c.DemoLimit, "RUN_LIMIT_USD": &c.RunLimit} {
		d := "20"
		switch k {
		case "DEVELOPMENT_LIMIT_USD":
			d = "15"
		case "DEMO_LIMIT_USD":
			d = "5"
		case "RUN_LIMIT_USD":
			d = "0.25"
		}
		v, e := money(k, d)
		if e != nil {
			return c, e
		}
		*p = v
	}
	vals := map[string]int64{"MAX_MODEL_ROUNDS": 8, "MAX_TOOL_CALLS": 16, "MAX_TRANSIENT_RETRIES": 2, "RUN_TIMEOUT_SECONDS": 180, "MODEL_TIMEOUT_SECONDS": 30, "OT_VARIANCE_BPS": 2000, "PRODUCTION_SHORTFALL_BPS": 1000, "DOWNTIME_THRESHOLD_MINUTES": 60, "ATTENDANCE_GRACE_MINUTES": 15}
	n := map[string]int64{}
	for k, d := range vals {
		v, e := integer(k, d)
		if e != nil {
			return c, e
		}
		n[k] = v
	}
	c.MaxRounds = int(n["MAX_MODEL_ROUNDS"])
	c.MaxTools = int(n["MAX_TOOL_CALLS"])
	c.MaxRetries = int(n["MAX_TRANSIENT_RETRIES"])
	c.RunTimeout = time.Duration(n["RUN_TIMEOUT_SECONDS"]) * time.Second
	c.ModelTimeout = time.Duration(n["MODEL_TIMEOUT_SECONDS"]) * time.Second
	c.OTThreshold = n["OT_VARIANCE_BPS"]
	c.ProdThreshold = n["PRODUCTION_SHORTFALL_BPS"]
	c.DowntimeThreshold = n["DOWNTIME_THRESHOLD_MINUTES"]
	c.AttendanceGrace = n["ATTENDANCE_GRACE_MINUTES"]
	if c.MaxRounds < 1 || c.MaxRounds > 8 || c.MaxTools < 1 || c.MaxTools > 16 || c.MaxRetries > 2 || c.RunTimeout <= 0 || c.RunTimeout > 180*time.Second || c.ModelTimeout <= 0 || c.ModelTimeout > 30*time.Second {
		return c, errors.New("batas agent melampaui guardrail atau nol")
	}
	if c.CreditLimit <= 0 || c.CreditLimit > 20_000_000 || c.DevelopmentLimit > 15_000_000 || c.DemoLimit > 5_000_000 || c.RunLimit <= 0 || c.RunLimit > 250000 {
		return c, errors.New("batas kredit melampaui guardrail US$20/15/5/0.25")
	}
	return c, nil
}
