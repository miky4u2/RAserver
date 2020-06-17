package handler

import (
	"github.com/miky4u2/RAserver/server/common"
	"github.com/miky4u2/RAserver/server/config"
	"log"
	"net/http"
	"os"
	"time"
)

// ServerCtlHandler HTTP handler function
//
func ServerCtlHandler(w http.ResponseWriter, req *http.Request) {

	// Parse templates if not done already
	parseTemplates()

	fData := struct {
		Action   string
		Feedback []string
		Type     string
	}{}

	// Check if IP is allowed, abort if not
	if !common.IsIPAllowed(req, config.Settings.AllowedIPs) {
		http.Error(w, http.StatusText(403), http.StatusForbidden)
		return
	}

	// Process control command
	if req.PostFormValue(`Action`) == `Ctl` {
		fData.Action = req.PostFormValue("Action") // This also runs req.ParseForm()
		fData.Type = req.PostFormValue("Type")

		// If the control command type is 'stop'
		if fData.Type == `stop` {
			fData.Feedback = []string{"Shutting down RAserver now, bye bye...\n"}
			log.Println(`Ctl Stop received. Shutting down RAserver now`)

			go func() { time.Sleep(2 * time.Second); os.Exit(0) }()
		}

		// If the control command type is 'status'
		if fData.Type == `status` {
			fData.Feedback = []string{`RAserver version ` + config.Version + " running normally\n"}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tpl.ExecuteTemplate(w, `serverctl.gohtml`, fData)
	if err != nil {
		log.Println(err)
	}
}
