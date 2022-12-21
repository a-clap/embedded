package main

import "github.com/a-clap/iot/internal/rest"

func main() {
	w, err := New()
	if err != nil {
		panic(err)
	}

	srv, err := rest.New(rest.WithWifiHandler(w))
	if err != nil {
		panic(err)
	}

	panic(srv.Run())
}
