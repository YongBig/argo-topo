/**
 * @Author: dy
 * @Description:
 * @File: config
 * @Date: 2022/5/7 16:31
 */
package controllers

import "time"

const (
	DefaultMaxConcurrentReconciles = 1
	DefaultCacheSyncTimeout        = time.Hour
)

type Config struct {
	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles which can be run.
	// Defaults to 1.
	MaxConcurrentReconciles int
	// CacheSyncTimeout refers to the time limit set to wait for syncing caches.
	// Defaults to 2 minutes if not set.
	CacheSyncTimeout time.Duration
}

func CreateConfig() Config {
	return Config{
		MaxConcurrentReconciles: DefaultMaxConcurrentReconciles,
		CacheSyncTimeout:        DefaultCacheSyncTimeout,
	}
}
