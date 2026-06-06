package api

import (
	"net/http"

	"github.com/bulaya-ute/appctl/internal/db"
)

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := db.GetConfig(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	var patch map[string]string
	if err := decodeJSON(r, &patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	for k, v := range patch {
		if err := db.SetConfigValue(s.db, k, v); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	cfg, err := db.GetConfig(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}
