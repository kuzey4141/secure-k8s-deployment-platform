package deployments

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("deployment not found")

type Repository struct {
	db *sql.DB
}

type CreateParams struct {
	AppName     string
	Image       string
	Namespace   string
	Replicas    int
	CPULimit    string
	MemoryLimit string
	Privileged  bool
	Status      string
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, params CreateParams) (Deployment, error) {
	const query = `
		INSERT INTO deployments (
			app_name,
			image,
			namespace,
			replicas,
			cpu_limit,
			memory_limit,
			privileged,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING
			id::text,
			app_name,
			image,
			namespace,
			replicas,
			COALESCE(cpu_limit, ''),
			COALESCE(memory_limit, ''),
			privileged,
			status,
			created_at,
			updated_at
	`

	var deployment Deployment
	err := r.db.QueryRowContext(
		ctx,
		query,
		params.AppName,
		params.Image,
		params.Namespace,
		params.Replicas,
		nullIfEmpty(params.CPULimit),
		nullIfEmpty(params.MemoryLimit),
		params.Privileged,
		params.Status,
	).Scan(
		&deployment.ID,
		&deployment.AppName,
		&deployment.Image,
		&deployment.Namespace,
		&deployment.Replicas,
		&deployment.CPULimit,
		&deployment.MemoryLimit,
		&deployment.Privileged,
		&deployment.Status,
		&deployment.CreatedAt,
		&deployment.UpdatedAt,
	)
	if err != nil {
		return Deployment{}, err
	}

	return deployment, nil
}

func (r *Repository) List(ctx context.Context) ([]Deployment, error) {
	const query = `
		SELECT
			id::text,
			app_name,
			image,
			namespace,
			replicas,
			COALESCE(cpu_limit, ''),
			COALESCE(memory_limit, ''),
			privileged,
			status,
			created_at,
			updated_at
		FROM deployments
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Deployment, 0)
	for rows.Next() {
		var deployment Deployment
		if err := rows.Scan(
			&deployment.ID,
			&deployment.AppName,
			&deployment.Image,
			&deployment.Namespace,
			&deployment.Replicas,
			&deployment.CPULimit,
			&deployment.MemoryLimit,
			&deployment.Privileged,
			&deployment.Status,
			&deployment.CreatedAt,
			&deployment.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, deployment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Deployment, error) {
	const query = `
		SELECT
			id::text,
			app_name,
			image,
			namespace,
			replicas,
			COALESCE(cpu_limit, ''),
			COALESCE(memory_limit, ''),
			privileged,
			status,
			created_at,
			updated_at
		FROM deployments
		WHERE id = $1::uuid
	`

	var deployment Deployment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&deployment.ID,
		&deployment.AppName,
		&deployment.Image,
		&deployment.Namespace,
		&deployment.Replicas,
		&deployment.CPULimit,
		&deployment.MemoryLimit,
		&deployment.Privileged,
		&deployment.Status,
		&deployment.CreatedAt,
		&deployment.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Deployment{}, ErrNotFound
	}
	if err != nil {
		return Deployment{}, err
	}

	return deployment, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}

	return value
}
