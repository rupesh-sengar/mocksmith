package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) ActivateDeployment(ctx context.Context, envID, specVersionID uuid.UUID) (Deployment, error) {
	var dep Deployment

	err := r.WithTx(ctx, func(tx pgx.Tx) error {
		const deactivate = `
			UPDATE deployments
			SET is_active = false
			WHERE environment_id = $1 AND is_active = true;
		`
		if _, err := tx.Exec(ctx, deactivate, envID); err != nil {
			return err
		}

		const insert = `
			INSERT INTO deployments (environment_id, spec_version_id, is_active, activated_at)
			VALUES ($1, $2, true, now())
			RETURNING id, environment_id, spec_version_id, is_active, created_at, activated_at;
		`

		return tx.QueryRow(ctx, insert, envID, specVersionID).Scan(
			&dep.ID, &dep.EnvironmentID, &dep.SpecVersionID, &dep.IsActive, &dep.CreatedAt, &dep.ActivatedAt,
		)
	})

	return dep, err
}

func ptrTime(t time.Time) *time.Time { return &t }
