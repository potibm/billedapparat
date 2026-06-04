package utils

import "context"

func RunWorker[T any](ctx context.Context, ch <-chan T, process func(context.Context, T)) {
	for {
		select {
		case <-ctx.Done():
			drainCtx := context.WithoutCancel(ctx)

			for {
				select {
				case req := <-ch:
					process(drainCtx, req)
				default:
					return // channel is empty
				}
			}

		case req, ok := <-ch:
			if !ok {
				return // channel closed
			}

			process(ctx, req)
		}
	}
}
