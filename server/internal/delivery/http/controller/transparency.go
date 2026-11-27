package controller

import "net/http"

// GetTransparency handles GET /transparency, the "server bersama" numbers. The usecase caches the
// provider's answer, and browsers and proxies may reuse it for 5 minutes.
func (ctrl *Controller) GetTransparency(w http.ResponseWriter, r *http.Request) error {
	report, err := ctrl.transparency.Get(r.Context())
	if err != nil {
		return err
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	return WriteJSON(w, http.StatusOK, dataResponse{Data: report})
}
