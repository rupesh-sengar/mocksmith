package services

import (
	"context"
	"errors"
	"testing"
)

func TestProjectsService_GetProjectById_StoreUnavailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  *ProjectsService
	}{
		{
			name: "nil service",
			svc:  nil,
		},
		{
			name: "nil repo",
			svc:  NewProjectsService(nil),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.svc.GetProjectById(context.Background(), "any")
			if !errors.Is(err, ErrProjectsStoreUnavailable) {
				t.Fatalf("expected ErrProjectsStoreUnavailable, got %v", err)
			}
		})
	}
}
