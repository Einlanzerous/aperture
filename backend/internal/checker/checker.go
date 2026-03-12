package checker

import "context"

// Checker is implemented by each service-check strategy (HTTP, Docker, etc.).
type Checker interface {
	Check(ctx context.Context) (Status, int, int64, string)
}
