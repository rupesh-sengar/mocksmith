package db

func GetProjectQuery() string {
	return `
	SELECT id, slug, name, created_at, updated_at
	FROM projects
	WHERE id = $1;
	`
}

func CreateProjectQuery() string {
	return `
	INSERT INTO projects (name, slug)
	VALUES ($1, $2)
	RETURNING id, slug, name, created_at, updated_at;
	`
}

func GetProjectsQuery() string {
	return `
	SELECT id, slug, name, created_at, updated_at
	FROM projects;
	`
}
