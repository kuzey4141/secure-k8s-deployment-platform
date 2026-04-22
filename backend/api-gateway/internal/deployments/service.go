package deployments

import "context"

type Service struct {
	repo *Repository
}

// NewService creates the deployment service layer.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create normalizes the request and saves a new deployment with the initial pending status.
func (s *Service) Create(ctx context.Context, req CreateRequest) (Deployment, error) {
	req = req.Normalize()

	return s.repo.Create(ctx, CreateParams{
		AppName:     req.AppName,
		Image:       req.Image,
		Namespace:   req.Namespace,
		Replicas:    req.Replicas,
		CPULimit:    req.CPULimit,
		MemoryLimit: req.MemoryLimit,
		Privileged:  req.Privileged,
		Status:      "pending",
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
