package repository

import "context"

// Pinger reports whether the database connection is alive, see ADR 007.
type Pinger interface {
	Ping(ctx context.Context) error
}
