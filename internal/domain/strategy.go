package domain

import "context"

type Strategy interface {
	NextServer() (*Server, error)
	UpdateServers([]*Server)
	Shutdown(ctx context.Context) error
}
