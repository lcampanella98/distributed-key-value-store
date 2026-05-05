package main

import (
	"fmt"
	"net/http"

	"github.com/lcampanella98/distributed-key-value-store/internal/api"
)

func main() {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: api.GetHandler(),
	}
	fmt.Println("serving...")

	srv.ListenAndServe()

}
