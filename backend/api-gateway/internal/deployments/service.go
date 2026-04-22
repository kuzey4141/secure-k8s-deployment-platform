package deployments

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

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

func (s *Service) List(ctx context.Context) ([]Deployment, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (Deployment, error) {
	return s.repo.GetByID(ctx, id)
}
