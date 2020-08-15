package util

import "context"

// GetContextWithCancel --
func GetContextWithCancel(parent context.Context) (context.Context, func()) {
	return context.WithCancel(parent)
}
