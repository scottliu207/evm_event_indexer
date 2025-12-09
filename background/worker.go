package background

import "context"

type Worker interface {
	Run(ctx context.Context) error
}
