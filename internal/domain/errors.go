package domain

import "errors"

var (
    ErrNoAvailableServers = errors.New("no available servers")
    ErrInvalidServerList  = errors.New("invalid server list")
)