package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"example.com/marketplace-presence/marketplace"
)

type server struct {
	infrai *marketplace.InfraiClient
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}

	s := &server{infrai: marketplace.NewInfraiClient(key)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /workspaces/{workspace}/online", s.online)
	mux.HandleFunc("POST /workspaces/{workspace}/handoffs", s.handoff)
	log.Println("marketplace presence listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s *server) online(w http.ResponseWriter, r *http.Request) {
	presence, err := s.infrai.Presence(r.Context(), channel(r.PathValue("workspace")))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, presence)
}

func (s *server) handoff(w http.ResponseWriter, r *http.Request) {
	var h marketplace.Handoff
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&h); err != nil {
		http.Error(w, "invalid handoff request", http.StatusBadRequest)
		return
	}
	presence, err := s.infrai.Presence(r.Context(), channel(r.PathValue("workspace")))
	if err != nil {
		writeError(w, err)
		return
	}
	h.SellerOnline = presence.Has(h.SellerID)
	if !h.Ready() {
		writeJSON(w, http.StatusConflict, map[string]string{"state": "pending"})
		return
	}
	if err := s.infrai.PublishHandoff(r.Context(), channel(r.PathValue("workspace")), h); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"state": "handed_off", "order_id": h.OrderID})
}

func channel(workspace string) string {
	return "marketplace-" + strings.TrimSpace(workspace)
}

func writeError(w http.ResponseWriter, err error) {
	var apiErr *marketplace.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Code})
		return
	}
	http.Error(w, "upstream request failed", http.StatusBadGateway)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
