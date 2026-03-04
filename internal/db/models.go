package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Environment struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"projectId"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SpecVersion struct {
	ID              uuid.UUID       `json:"id"`
	EnvironmentID   uuid.UUID       `json:"environmentId"`
	Version         int32           `json:"version"`
	RawSpecText     string          `json:"rawSpecText"`
	RawSpecSHA256   string          `json:"rawSpecSha256"`
	CompilerVersion string          `json:"compilerVersion"`
	CompileStatus   string          `json:"compileStatus"` // compiled|failed
	CompileErrors   json.RawMessage `json:"compileErrors"` // jsonb
	CompiledSnap    json.RawMessage `json:"compiledSnapshot"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type Deployment struct {
	ID            uuid.UUID  `json:"id"`
	EnvironmentID uuid.UUID  `json:"environmentId"`
	SpecVersionID uuid.UUID  `json:"specVersionId"`
	IsActive      bool       `json:"isActive"`
	CreatedAt     time.Time  `json:"createdAt"`
	ActivatedAt   *time.Time `json:"activatedAt"`
}
