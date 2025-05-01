package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
)

func main() {
    port := flag.Int("port", 8080, "server port")
    flag.Parse()

    http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Server response from port %d\n", *port)
    })

    log.Printf("Starting server on :%d", *port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}