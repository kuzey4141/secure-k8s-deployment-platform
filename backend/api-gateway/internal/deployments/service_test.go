package deployments

import (
	"context"
	"errors"
	"testing"

	"github.com/kuzey/secure-deploy-platform/backend/api-gateway/internal/policy"
)

type fakeStore struct {
	created CreateParams
	result  Deployment
	err     error
}

func (s *fakeStore) Create(_ context.Context, params CreateParams) (Deployment, error) {
	s.created = params
	if s.err != nil {
		return Deployment{}, s.err
	}
	if s.result.ID == "" {
		s.result = Deployment{
			ID:         "deployment-id",
			AppName:    params.AppName,
			Image:      params.Image,
			Namespace:  params.Namespace,
			Replicas:   params.Replicas,
			Privileged: params.Privileged,
			Status:     params.Status,
		}
	}

	return s.result, nil
}

func (s *fakeStore) List(context.Context) ([]Deployment, error) {
	return nil, nil
}

func (s *fakeStore) GetByID(context.Context, string) (Deployment, error) {
	return Deployment{}, nil
}

type fakeEvaluator struct {
	decision policy.Decision
	err      error
	input    policy.Input
}

func (e *fakeEvaluator) Evaluate(_ context.Context, input policy.Input) (policy.Decision, error) {
	e.input = input
	if e.err != nil {
		return policy.Decision{}, e.err
	}

	return e.decision, nil
}

func TestServiceCreateAccepted(t *testing.T) {
	repo := &fakeStore{}
	evaluator := &fakeEvaluator{
		decision: policy.Decision{Allowed: true},
	}
	service := NewService(repo, evaluator)

	deployment, err := service.Create(context.Background(), CreateRequest{
		AppName:     " payment-service ",
		Image:       "bugra/payment-service:v1.2.0",
		Namespace:   "production",
		Replicas:    2,
		CPULimit:    "500m",
		MemoryLimit: "512Mi",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if got, want := deployment.Status, "accepted"; got != want {
		t.Fatalf("deployment status = %q, want %q", got, want)
	}
	if got, want := repo.created.Status, "accepted"; got != want {
		t.Fatalf("repo status = %q, want %q", got, want)
	}
	if len(repo.created.Violations) != 0 {
		t.Fatalf("repo violations = %v, want none", repo.created.Violations)
	}
	if got, want := evaluator.input.AppName, "payment-service"; got != want {
		t.Fatalf("policy input app_name = %q, want %q", got, want)
	}
}

func TestServiceCreateRejected(t *testing.T) {
	repo := &fakeStore{}
	evaluator := &fakeEvaluator{
		decision: policy.Decision{
			Allowed: false,
			Violations: []policy.Violation{
				{
					ControlNo: "control_4",
					Severity:  "critical",
					Message:   "Privileged containers are not allowed",
				},
			},
		},
	}
	service := NewService(repo, evaluator)

	deployment, err := service.Create(context.Background(), CreateRequest{
		AppName:    "payment-service",
		Image:      "bugra/payment-service:v1.2.0",
		Namespace:  "production",
		Replicas:   2,
		Privileged: true,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if got, want := deployment.Status, "rejected"; got != want {
		t.Fatalf("deployment status = %q, want %q", got, want)
	}
	if got, want := repo.created.Status, "rejected"; got != want {
		t.Fatalf("repo status = %q, want %q", got, want)
	}
	if len(repo.created.Violations) != 1 {
		t.Fatalf("repo violations length = %d, want 1", len(repo.created.Violations))
	}
	violation := repo.created.Violations[0]
	if violation.ControlNo != "control_4" || violation.Severity != "critical" || violation.Message != "Privileged containers are not allowed" {
		t.Fatalf("repo violation = %+v, want privileged control_4 violation", violation)
	}
}

func TestServiceCreatePolicyError(t *testing.T) {
	repo := &fakeStore{}
	evaluator := &fakeEvaluator{
		err: errors.New("policy evaluator unavailable"),
	}
	service := NewService(repo, evaluator)

	_, err := service.Create(context.Background(), CreateRequest{
		AppName:   "payment-service",
		Image:     "bugra/payment-service:v1.2.0",
		Namespace: "production",
		Replicas:  2,
	})
	if err == nil {
		t.Fatal("Create returned nil error, want non-nil")
	}
	if repo.created.Status != "" {
		t.Fatalf("repo Create should not be called when policy evaluation fails, got status %q", repo.created.Status)
	}
}
