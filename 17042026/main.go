
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "UVI Go Backend Running"})
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
