package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"mocksmith/internal/db"
)

type ProjectsService struct {
	repo *db.Repo
}

type CreateProjectPayload struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

var (
	ErrProjectsStoreUnavailable = errors.New("projects store unavailable")
	ErrProjectNotFound          = errors.New("project not found")
)

func NewProjectsService(repo *db.Repo) *ProjectsService {
	return &ProjectsService{repo: repo}
}

func (r *ProjectsService) GetProjectById(ctx context.Context, id string) (*db.Project, error) {
	if r == nil || r.repo == nil || r.repo.Pool == nil {
		return nil, ErrProjectsStoreUnavailable
	}

	query := db.GetProjectQuery()
	var project db.Project
	err := r.repo.Pool.QueryRow(ctx, query, id).Scan(&project.ID, &project.Slug, &project.Name, &project.CreatedAt, &project.UpdatedAt)
	fmt.Println(err)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("get project by id: %w", err)
	}
	return &project, nil
}

func (r *ProjectsService) CreateProject(ctx context.Context, payload CreateProjectPayload) (*db.Project, error) {
	if r == nil || r.repo == nil || r.repo.Pool == nil {
		return nil, ErrProjectsStoreUnavailable
	}

	query := db.CreateProjectQuery()
	var project db.Project
	err := r.repo.Pool.QueryRow(ctx, query, payload.Name, payload.Slug).Scan(&project.ID, &project.Slug, &project.Name, &project.CreatedAt, &project.UpdatedAt)
	fmt.Println(err)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return &project, nil
}

func (r *ProjectsService) GetProjects(ctx context.Context) ([]*db.Project, error) {
	if r == nil || r.repo == nil || r.repo.Pool == nil {
		return nil, ErrProjectsStoreUnavailable
	}

	query := db.GetProjectsQuery()
	rows, err := r.repo.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get projects: %w", err)
	}
	defer rows.Close()

	var projects []*db.Project
	for rows.Next() {
		var project db.Project
		if err := rows.Scan(&project.ID, &project.Slug, &project.Name, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, fmt.Errorf("get projects: %w", err)
		}
		projects = append(projects, &project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get projects: %w", err)
	}
	return projects, nil
}
