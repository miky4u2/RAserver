package api

import (
	"encoding/json"
	"github.com/miky4u2/RAserver/server/common"
	"github.com/miky4u2/RAserver/server/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// Type to hold agents configuration settings
//
type agentLocalConfig struct {
	AgentIP  []string          `jason:"agentIP"`
	AgentOS  string            `jason:"agentOS"`
	AgentURL string            `jason:"agentURL"`
	Modules  map[string]string `jason:"modules"`
	TLScert  string            `jason:"TLScert"`
}

// Download serves files to Agents (Only update archives for now)
//
func Download(w http.ResponseWriter, req *http.Request) {

	// POST method only
	if req.Method != `POST` {
		http.Error(w, http.StatusText(403), http.StatusForbidden)
		return
	}

	// Instantiate a downloadReq struct to be populated
	downloadReq := struct {
		AgentID string `json:"agentID"`
		Archive string `json:"archive"`
	}{}

	// Populate the downloadReq struct with received json request
	json.NewDecoder(req.Body).Decode(&downloadReq)

	// Validate agentID string, return forbidden if AgentID is malformated
	if !regexp.MustCompile(`^[a-zA-Z0-9]+[a-zA-Z0-9\.\-_]*$`).MatchString(downloadReq.AgentID) {
		http.Error(w, http.StatusText(403), http.StatusForbidden)
		log.Println(`Download failed, AgentID incorrect:`, downloadReq.AgentID)
		return
	}

	// Try to load agent local config
	agentLocalConfig, err := getAgentLocalConfig(downloadReq.AgentID)
	if err != nil {
		log.Println(`Cannot load agent`, downloadReq.AgentID, `local config`)
		http.Error(w, http.StatusText(403), http.StatusForbidden)
		return
	}

	// Validate agent IP
	if !common.IsIPAllowed(req, agentLocalConfig.AgentIP) {
		log.Println(`Agent`, downloadReq.AgentID, `IP not allowed`)
		http.Error(w, http.StatusText(403), http.StatusForbidden)
		return
	}

	// If an update archive is requested and the file is available, respond with file
	if archiveFile := filepath.Join(config.AppBasePath, `agents`, `archives`, downloadReq.AgentID+`.tar.gz`); downloadReq.Archive == `update` && common.FileExists(archiveFile) {
		http.ServeFile(w, req, archiveFile)
		log.Println(`Update archive tar.gz downloaded by agent:`, downloadReq.AgentID)
		return
	}

	// Respond with 403 if the request was not dealt with above
	http.Error(w, http.StatusText(403), http.StatusForbidden)
	log.Println(`Incorrect download request for agent:`, downloadReq.AgentID)
}

// Retrieves agent local json config
//
func getAgentLocalConfig(agentID string) (agentLocalConfig, error) {
	// Load agent local config
	localConfig := &agentLocalConfig{}

	configFile, err := os.Open(filepath.Join(config.AppBasePath, `agents`, `configs`, `local`, agentID+`.json`))
	defer configFile.Close()
	if err != nil {
		return *localConfig, err
	}

	jasonParser := json.NewDecoder(configFile)
	err = jasonParser.Decode(localConfig)
	if err != nil {
		return *localConfig, err
	}

	return *localConfig, err
}
