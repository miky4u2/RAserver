package handler

import (
	"log"
	"net/http"

	"github.com/miky4u2/RAserver/server/common"
	"github.com/miky4u2/RAserver/server/config"
)

// Index HTTP handler function
//
func Index(w http.ResponseWriter, req *http.Request) {

	// Parse templates if not done already
	parseTemplates()

	// Check if IP is allowed, abort if not
	if !common.IsIPAllowed(req, config.Settings.AllowedIPs) {
		http.Error(w, http.StatusText(403), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tpl.ExecuteTemplate(w, `index.gohtml`, nil)
	if err != nil {
		log.Println(err)
	}
}
