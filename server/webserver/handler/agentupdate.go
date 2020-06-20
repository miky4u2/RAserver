package handler

import (
	"archive/tar"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/miky4u2/RAserver/server/common"
	"github.com/miky4u2/RAserver/server/config"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

// AgentUpdate HTTP handler function
//
func AgentUpdate(w http.ResponseWriter, req *http.Request) {
	// Parse templates if not done already
	parseTemplates()

	fData := struct {
		Action     string
		Agents     []string
		Feedback   []string
		UpdateType string
	}{}

	// Check if IP is allowed, abort if not
	if !common.IsIPAllowed(req, config.Settings.AllowedIPs) {
		http.Error(w, http.StatusText(403), http.StatusForbidden)
		return
	}

	// No Action, load agents list in fData.Agents
	if req.PostFormValue(`Action`) == "" {
		dir := filepath.Join(config.AppBasePath, `agents`, `configs`, `local`)
		dirFiles, err := ioutil.ReadDir(dir)

		if err != nil {
			return
		}
		for _, file := range dirFiles {
			if file.Mode().IsRegular() && filepath.Ext(file.Name()) == `.json` {
				fData.Agents = append(fData.Agents, strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))
			}
		}
	}

	// Update agents
	if req.PostFormValue(`Action`) == `Update` {
		fData.Action = req.PostFormValue("Action") // This also runs req.ParseForm()
		fData.Agents = req.Form[`Agents`]
		fData.UpdateType = req.PostFormValue("UpdateType")

		// Create wait group
		wg := sync.WaitGroup{}

		// Create a receiving chanel to get feedback of the updates
		// and create a slice to store the feedback strings
		ch := make(chan string)
		feedback := []string{}

		// Launch go routine to capture feedback strings from chanel
		go func(ch <-chan string, feedback *[]string) {
			for fb := range ch {
				*feedback = append(*feedback, fb)
			}
		}(ch, &feedback)

		// Launch go routines to create tar.gz archives and notify agents to pickup the update
		for _, agentID := range fData.Agents {
			wg.Add(1)
			go updateAgent(agentID, fData.UpdateType, ch, &wg, config.Settings.ValidateAgentTLS)
		}

		// Wait for all go routines to finish and close the feedback chanel
		wg.Wait()
		time.Sleep(500 * time.Microsecond) // gives time to the receiving chanel to receive the last string

		close(ch)

		fData.Feedback = feedback
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tpl.ExecuteTemplate(w, `agentupdate.gohtml`, fData)
	if err != nil {
		log.Println(err)
	}
}

// Creates an agent update archive and sends an update request to an agent
//
func updateAgent(agentID string, UpdateType string, ch chan<- string, wg *sync.WaitGroup, validateAgentTLS bool) {

	defer wg.Done()

	err := createArchive(agentID)
	if err != nil {
		log.Println(`Could not create agent`, agentID, `archive file -`, err)
		ch <- `Could not create agent ` + agentID + ` archive file - ` + err.Error() + "\n"
		return
	}
	log.Println(`Archive for agent`, agentID, `was successfully created`)
	ch <- `Archive for agent ` + agentID + ` was successfully created` + "\n"

	err = sendUpdateRequest(agentID, UpdateType, validateAgentTLS)
	if err != nil {
		log.Println(`Error sending update request to agent`, agentID, `-`, err)
		ch <- `Error sending update request to agent ` + agentID + ` - ` + err.Error() + "\n"
		return
	}
	log.Println(`Agent`, agentID, `successfully downloaded archive and updated itself`)
	ch <- `Agent ` + agentID + ` successfully downloaded archive and updated itself` + "\n"
}

// Creates an agent update tar.gz file
//
func createArchive(agentID string) error {
	// Create the output agent tar.gz file
	file, err := os.Create(filepath.Join(config.AppBasePath, `agents`, `archives`, agentID+`.tar.gz`))
	if err != nil {
		return err
	}
	defer file.Close()

	// Create the gzip & tar writers
	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Slice to hold source and destination paths
	type path struct {
		origPath string
		destPath string
	}
	paths := []path{}

	// Get agent local config
	localConfig, err := getAgentLocalConfig(agentID)
	if err != nil {
		return err
	}

	// Work out agent binary name and path according to OS
	var agentPath string
	var agentFileName string
	if localConfig.AgentOS == `windows` {
		agentFileName = `agent.exe`
		agentPath = filepath.Join(config.AppBasePath, `agents`, `binaries`, `windows`, `agent.exe`)
	} else if localConfig.AgentOS == `linux` {
		agentFileName = `agent`
		agentPath = filepath.Join(config.AppBasePath, `agents`, `binaries`, `linux`, `agent`)
	} else if localConfig.AgentOS == `osx` {
		agentFileName = `agent`
		agentPath = filepath.Join(config.AppBasePath, `agents`, `binaries`, `osx`, `agent`)
	}
	// Add agent Binary to paths
	paths = append(paths, path{agentPath, `bin/` + agentFileName})

	// Add agent modules to paths
	for k, v := range localConfig.Modules {
		modulePath := filepath.Join(config.AppBasePath, `agents`, `modules`, v)
		paths = append(paths, path{modulePath, `modules/` + k})
	}

	// Add agent remote config to paths
	remoteConfigPath := filepath.Join(config.AppBasePath, `agents`, `configs`, `remote`, agentID+`.json`)
	paths = append(paths, path{remoteConfigPath, `conf/config.json`})

	// Add agent TLS certificate and key to paths
	tlsCertPath := filepath.Join(config.AppBasePath, `agents`, `certs`, localConfig.TLScert+`.cert`)
	tlsKeyPath := filepath.Join(config.AppBasePath, `agents`, `certs`, localConfig.TLScert+`.key`)
	paths = append(paths, path{tlsCertPath, `conf/cert.pem`}, path{tlsKeyPath, `conf/key.pem`})

	// Add files in tar.gz archive
	for _, path := range paths {
		if err := addFileToTar(tw, path.origPath, path.destPath); err != nil {
			return err
		}
	}
	return err
}

// Sends a json update request to an agent
//
func sendUpdateRequest(agentID string, UpdateType string, validateAgentTLS bool) error {
	response := struct {
		Status    string   `json:"status"`
		ErrorMsgs []string `json:"errorMsgs"`
	}{}

	localConfig, err := getAgentLocalConfig(agentID)
	if err != nil {
		return err
	}

	agentURL := localConfig.AgentURL + `/update`

	req, err := http.NewRequest(`POST`, agentURL, strings.NewReader(`{"type":"`+UpdateType+`"}`))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Do we validate the agent TLS certificate ? true/false
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !validateAgentTLS},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Populate the response struct with received json request
	json.NewDecoder(resp.Body).Decode(&response)

	if response.Status != `done` {
		remoteErrors := strings.Join(response.ErrorMsgs[:], ",")
		err = errors.New(`Remote errors: ` + remoteErrors)
		return err
	}

	return err
}

// Helper function to add files in tar file
//
func addFileToTar(tw *tar.Writer, origPath string, destPath string) error {
	file, err := os.Open(origPath)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {
		// Create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = destPath
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		header.ModTime = stat.ModTime()
		// Write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// Copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
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
