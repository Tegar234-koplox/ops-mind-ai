package config

import (
	"strings"
	"testing"
)

func TestP1ActivationIsExplicit(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/opsmind_test")
	t.Setenv("MODEL_MODE", "mock")
	t.Setenv("SESSION_SECRET", strings.Repeat("test", 10))
	t.Setenv("OPERATOR_PASSWORD", "configuration-test-password")
	t.Setenv("ENABLE_P1_RULES", "false")
	t.Setenv("SEED_P1_DEMO", "false")
	t.Setenv("COST_VARIANCE_BPS", "2000")
	cfg, err := Load()
	if err != nil || cfg.EnableP1 || cfg.SeedP1Demo || cfg.CostThreshold != 2000 {
		t.Fatal("P1 default", cfg.EnableP1, err)
	}
	t.Setenv("SEED_P1_DEMO", "true")
	if _, err = Load(); err == nil {
		t.Fatal("P1 demo seed accepted without P1 activation")
	}
	t.Setenv("ENABLE_P1_RULES", "true")
	cfg, err = Load()
	if err != nil || !cfg.EnableP1 || !cfg.SeedP1Demo {
		t.Fatal("P1 activation", err)
	}
	t.Setenv("ENABLE_P1_RULES", "invalid")
	if _, err = Load(); err == nil {
		t.Fatal("invalid P1 flag accepted")
	}
	t.Setenv("ENABLE_P1_RULES", "true")
	t.Setenv("COST_VARIANCE_BPS", "-1")
	if _, err = Load(); err == nil {
		t.Fatal("negative cost threshold accepted")
	}
}
