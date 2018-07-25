package main

import (
	"log"
	"net/http"
)

func main() {
	port := ":9595"
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("a b c d e f g -pq-pxxq-\n"))
	})
	log.Println("test server listening on", port)
	http.ListenAndServe(port, nil)
}
