package http

import (
	"net/http"
	"strconv"
	"time"
)

func NewServer(
	h http.Handler,
	port int,
	rt time.Duration,
	rht time.Duration,
	wt time.Duration,
	it time.Duration,
) *http.Server {
	return &http.Server{
        Addr:         ":" + strconv.Itoa(port),
        Handler:      h,
        ReadTimeout:  rt,
		ReadHeaderTimeout: rht,
        WriteTimeout: wt,
		IdleTimeout: it,
    }
}
