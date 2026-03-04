package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Repo) CreateProject(ctx context.Context, slug, name string) (Project, error) {
	const q = `
	INSERT INTO projects (slug,name) VALUES($1,$2) RETURNING id, created_at, updated_at, slug,name;
	`
	var p Project
	err := r.Pool.QueryRow(ctx, q, slug, name).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt, &p.Slug, &p.Name)
	return p, err
}

func (r *Repo) GetEnvironmentBySlugs(ctx context.Context, projectSlug, envSlug string) (Environment, error) {
	const q = `
		SELECT e.id, e.project_id, e.slug, e.name, e.created_at, e.updated_at
		FROM environments e
		JOIN projects p ON p.id = e.project_id
		WHERE p.slug = $1 AND e.slug = $2;
	`

	var e Environment
	err := r.Pool.QueryRow(ctx, q, projectSlug, envSlug).Scan(
		&e.ID, &e.ProjectID, &e.Slug, &e.Name, &e.CreatedAt, &e.UpdatedAt,
	)
	return e, err
}

func scanNowPtr(t time.Time) *time.Time { return &t }

var _ pgx.Tx // silence unused import if you move functions around
