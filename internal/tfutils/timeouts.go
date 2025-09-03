package tfutils

import (
	"context"
	"time"
)

// DefaultTimeouts defines the default timeout values for all operations for all resources.
// Individual resources can define a timeout nested attribute to override these defaults:
// See https://developer.hashicorp.com/terraform/plugin/framework/migrating/resources/timeouts
var DefaultTimeouts = struct {
	Create time.Duration
	Read   time.Duration
	Update time.Duration
	Delete time.Duration
}{
	Create: 1 * time.Minute,
	Read:   1 * time.Minute,
	Update: 1 * time.Minute,
	Delete: 1 * time.Minute,
}

func WithDefaultCreateTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, DefaultTimeouts.Create)
}

func WithDefaultReadTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, DefaultTimeouts.Read)
}

func WithDefaultUpdateTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, DefaultTimeouts.Update)
}

func WithDefaultDeleteTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, DefaultTimeouts.Delete)
}
