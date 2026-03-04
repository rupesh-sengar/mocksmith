package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func shaHex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func (r *Repo) CreateSpecVersion(
	ctx context.Context,
	envID uuid.UUID,
	rawSpec string,
	compilerVersion string,
	compileStatus string, // "compiled" | "failed"
	compileErrors json.RawMessage, // jsonb
	compiledSnapshot json.RawMessage, // jsonb
) (SpecVersion, error) {
	var out SpecVersion

	err := r.WithTx(ctx, func(tx pgx.Tx) error {
		// Lock “version counter” via environments row (simple + effective)
		// Alternative: a separate counter table.
		const lockEnv = `SELECT 1 FROM environments WHERE id=$1 FOR UPDATE;`
		if _, err := tx.Exec(ctx, lockEnv, envID); err != nil {
			return err
		}

		const maxV = `SELECT COALESCE(MAX(version), 0) FROM spec_versions WHERE environment_id=$1;`
		var curr int32
		if err := tx.QueryRow(ctx, maxV, envID).Scan(&curr); err != nil {
			return err
		}
		next := curr + 1

		const ins = `
			INSERT INTO spec_versions (
				environment_id, version, raw_spec_text, raw_spec_sha256,
				compiler_version, compile_status, compile_errors, compiled_snapshot
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id, environment_id, version, raw_spec_text, raw_spec_sha256,
			          compiler_version, compile_status, compile_errors, compiled_snapshot, created_at;
		`

		return tx.QueryRow(ctx, ins,
			envID, next, rawSpec, shaHex(rawSpec),
			compilerVersion, compileStatus, compileErrors, compiledSnapshot,
		).Scan(
			&out.ID, &out.EnvironmentID, &out.Version, &out.RawSpecText, &out.RawSpecSHA256,
			&out.CompilerVersion, &out.CompileStatus, &out.CompileErrors, &out.CompiledSnap, &out.CreatedAt,
		)
	})

	return out, err
}
