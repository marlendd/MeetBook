package handler

import (
	"net/http"

	"github.com/internships-backend/test-backend-marlendd/internal/httputil"
)

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
