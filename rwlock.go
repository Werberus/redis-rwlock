// Package rwlock is an adapter package to pkg/rwlock.
// Consider using pkg/rwlock package in new projects as this file may be eventually removed.
package rwlock

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/werberus/redis-rwlock/pkg/rwlock"
)

// Locker is an alias type to #rwlock.Locker
type Locker = rwlock.Locker

// Options is an alias type to #rwlock.Options
type Options = rwlock.Options

// New instance of RW-Locker.
// Use #rwlock.New instead.
func New(ctx context.Context, redisClient *redis.Client, keyLock, keyReadersCount, keyWriterIntent string, opts *Options) Locker {
	if opts == nil {
		opts = &Options{}
	}
	return rwlock.New(ctx, redisClient, keyLock, keyReadersCount, keyWriterIntent, *opts)
}
