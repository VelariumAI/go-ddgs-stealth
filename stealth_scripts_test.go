package goddgs

import "testing"

func TestStealthScriptsLevels(t *testing.T) {
	basic := StealthScripts(StealthLevelBasic)
	strong := StealthScripts(StealthLevelStrong)
	agg := StealthScripts(StealthLevelAggressive)
	if len(basic) == 0 || len(strong) == 0 || len(agg) == 0 {
		t.Fatal("expected non-empty scripts")
	}
	if len(agg) < len(strong) || len(strong) < len(basic) {
		t.Fatalf("expected progressive script sets, got basic=%d strong=%d agg=%d", len(basic), len(strong), len(agg))
	}
}
