package main

import (
	"encoding/json"
	"net/http"
)

func JSONdecoder(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func JSONencoder(w http.ResponseWriter, dst interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(dst)
}