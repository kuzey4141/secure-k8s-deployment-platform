package policy

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEngineEvaluateAccepted(t *testing.T) {
	engine := mustNewEngine(t)

	decision, err := engine.Evaluate(context.Background(), Input{
		AppName:     "payment-service",
		Image:       "bugra/payment-service:v1.2.0",
		Namespace:   "production",
		Replicas:    2,
		CPULimit:    "500m",
		MemoryLimit: "512Mi",
		Privileged:  false,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("decision allowed = false, want true; violations: %v", decision.Violations)
	}
	if len(decision.Violations) != 0 {
		t.Fatalf("decision violations = %v, want none", decision.Violations)
	}
}

func TestEngineEvaluateRejected(t *testing.T) {
	engine := mustNewEngine(t)

	decision, err := engine.Evaluate(context.Background(), Input{
		AppName:    "payment-service",
		Image:      "bugra/payment-service:latest",
		Namespace:  "production",
		Replicas:   1,
		Privileged: true,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("decision allowed = true, want false")
	}
	if len(decision.Violations) != 5 {
		t.Fatalf("decision violations length = %d, want 5", len(decision.Violations))
	}
	if !hasViolation(decision.Violations, "control_4", "critical", "Privileged containers are not allowed") {
		t.Fatalf("expected privileged violation in %+v", decision.Violations)
	}
}

func hasViolation(items []Violation, controlNo, severity, message string) bool {
	for _, item := range items {
		if item.ControlNo == controlNo && item.Severity == severity && item.Message == message {
			return true
		}
	}

	return false
}

func mustNewEngine(t *testing.T) *Engine {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller returned ok=false")
	}

	policyPath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "policies")
	engine, err := NewEngine(context.Background(), policyPath)
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}

	return engine
}
