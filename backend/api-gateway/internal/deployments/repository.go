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
	Violations  []PolicyViolation
}

// NewRepository creates a database-backed deployment repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a deployment row and returns the stored record.
func (r *Repository) Create(ctx context.Context, params CreateParams) (Deployment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Deployment{}, err
	}
	defer tx.Rollback()

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
	err = tx.QueryRowContext(
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

	if err := insertPolicyViolations(ctx, tx, deployment.ID, params.Violations); err != nil {
		return Deployment{}, err
	}
	if err := tx.Commit(); err != nil {
		return Deployment{}, err
	}

	return deployment, nil
}

// List returns all deployment records ordered by newest first.
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

// GetByID loads a single deployment by UUID.
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

// nullIfEmpty stores optional string fields as SQL NULL instead of empty text.
func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func insertPolicyViolations(ctx context.Context, tx *sql.Tx, deploymentID string, violations []PolicyViolation) error {
	if len(violations) == 0 {
		return nil
	}

	const query = `
		INSERT INTO policy_violations (
			deployment_id,
			control_no,
			message,
			severity
		)
		VALUES ($1::uuid, $2, $3, $4)
	`

	for _, violation := range violations {
		if _, err := tx.ExecContext(
			ctx,
			query,
			deploymentID,
			violation.ControlNo,
			violation.Message,
			violation.Severity,
		); err != nil {
			return err
		}
	}

	return nil
}
