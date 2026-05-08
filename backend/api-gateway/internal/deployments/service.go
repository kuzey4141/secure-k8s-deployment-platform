package deployments

import (
	"context"
	"fmt"

	"github.com/kuzey/secure-deploy-platform/backend/api-gateway/internal/policy"
)

type Store interface {
	Create(ctx context.Context, params CreateParams) (Deployment, error)
	List(ctx context.Context) ([]Deployment, error)
	GetByID(ctx context.Context, id string) (Deployment, error)
}

type Service struct {
	repo      Store
	evaluator policy.Evaluator
}

// NewService creates the deployment service layer.
func NewService(repo Store, evaluator policy.Evaluator) *Service {
	return &Service{
		repo:      repo,
		evaluator: evaluator,
	}
}

// Create normalizes the request, evaluates policy rules, and stores the request with its resulting status.
func (s *Service) Create(ctx context.Context, req CreateRequest) (Deployment, error) {
	req = req.Normalize()

	status := "accepted"
	violations := make([]PolicyViolation, 0)
	if s.evaluator != nil {
		decision, err := s.evaluator.Evaluate(ctx, policy.Input{
			AppName:     req.AppName,
			Image:       req.Image,
			Namespace:   req.Namespace,
			Replicas:    req.Replicas,
			CPULimit:    req.CPULimit,
			MemoryLimit: req.MemoryLimit,
			Privileged:  req.Privileged,
		})
		if err != nil {
			return Deployment{}, fmt.Errorf("evaluate deployment policies: %w", err)
		}
		violations = toPolicyViolations(decision.Violations)
		if !decision.Allowed {
			status = "rejected"
		}
	}

	return s.repo.Create(ctx, CreateParams{
		AppName:     req.AppName,
		Image:       req.Image,
		Namespace:   req.Namespace,
		Replicas:    req.Replicas,
		CPULimit:    req.CPULimit,
		MemoryLimit: req.MemoryLimit,
		Privileged:  req.Privileged,
		Status:      status,
		Violations:  violations,
	})
}

// List returns deployment history from the repository layer.
func (s *Service) List(ctx context.Context) ([]Deployment, error) {
	return s.repo.List(ctx)
}

// GetByID returns one deployment record by identifier.
func (s *Service) GetByID(ctx context.Context, id string) (Deployment, error) {
	return s.repo.GetByID(ctx, id)
}

func toPolicyViolations(items []policy.Violation) []PolicyViolation {
	if len(items) == 0 {
		return nil
	}

	violations := make([]PolicyViolation, 0, len(items))
	for _, item := range items {
		violations = append(violations, PolicyViolation{
			ControlNo: item.ControlNo,
			Severity:  item.Severity,
			Message:   item.Message,
		})
	}

	return violations
}
