package repo

// Permission represents repository access permission levels
type Permission int

const (
	PermNone  Permission = 0 // No access
	PermRead  Permission = 1 // Read-only access
	PermWrite Permission = 2 // Read and write access
)

// Repository represents a git repository with user permissions
type Repository struct {
	Name  string                   // Repository name
	Path  string                   // Full filesystem path to the repository
	Users map[string]Permission    // Map of username to permission level
}
