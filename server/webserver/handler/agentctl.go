package handler

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/miky4u2/RAserver/server/common"
	"github.com/miky4u2/RAserver/server/config"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AgentCtlHandler HTTP handler function
//
func AgentCtlHandler(w http.ResponseWriter, req *http.Request) {

	// Parse templates if not done already
	parseTemplates()

	fData := struct {
		Action   string
		Agents   []string
		Feedback []string
		Type     string
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

	// Send Ctl command to agents
	if req.PostFormValue(`Action`) == `Ctl` {
		fData.Action = req.PostFormValue("Action") // This also runs req.ParseForm()
		fData.Agents = req.Form[`Agents`]
		fData.Type = req.PostFormValue("Type")

		// Create wait group
		wg := sync.WaitGroup{}

		// Create a receiving chanel to get feedback
		// and create a slice to store the feedback strings
		ch := make(chan string)
		feedback := []string{}

		// Launch go routine to capture feedback strings from chanel
		go func(ch <-chan string, feedback *[]string) {
			for fb := range ch {
				*feedback = append(*feedback, fb)
			}
		}(ch, &feedback)

		// Launch go routines to send Ctl Command to agents
		for _, agentID := range fData.Agents {
			wg.Add(1)
			go sendCtl(agentID, fData.Type, ch, &wg, config.Settings.ValidateAgentTLS)
		}

		// Wait for all go routines to finish and close the feedback chanel
		wg.Wait()
		time.Sleep(500 * time.Microsecond) // gives time to the receiving chanel to receive the last string

		close(ch)

		fData.Feedback = feedback

	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tpl.ExecuteTemplate(w, `agentctl.gohtml`, fData)
	if err != nil {
		log.Println(err)
	}
}

// Send Ctl command to agent
//
func sendCtl(agentID string, Type string, ch chan<- string, wg *sync.WaitGroup, validateAgentTLS bool) {

	defer wg.Done()

	response := struct {
		Status    string   `json:"status"`
		ErrorMsgs []string `json:"errorMsgs"`
		Output    string   `json:"output"`
	}{}

	localConfig, err := getAgentLocalConfig(agentID)
	if err != nil {
		log.Println(`Could not read agent`, agentID, `local config file -`, err)
		ch <- `Could not read agent ` + agentID + ` local config file - ` + err.Error() + "\n"
		return
	}

	agentURL := localConfig.AgentURL + `/ctl`

	req, err := http.NewRequest(`POST`, agentURL, strings.NewReader(`{"type":"`+Type+`"}`))
	req.Header.Set("Content-Type", "application/json")

	// Do we validate the agent TLS certificate ? true/false
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !validateAgentTLS},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(`Could not send request to agent`, agentID, `-`, err)
		ch <- `Could not send request to agent ` + agentID + ` - ` + err.Error() + "\n"
		return
	}
	defer resp.Body.Close()

	// Populate the response struct with received json request
	json.NewDecoder(resp.Body).Decode(&response)

	if response.Status != `done` {
		remoteErrors := strings.Join(response.ErrorMsgs[:], ",")
		err = errors.New(`Remote errors: ` + remoteErrors)
		log.Println(`Remote errors from agent`, agentID, `-`, err)
		ch <- `Remote errors from agent ` + agentID + ` - ` + err.Error() + "\n"
		return
	}

	output, err := base64.StdEncoding.DecodeString(response.Output)
	if err == nil {
		ch <- `Agent ` + agentID + ` replied: ` + string(output) + "\n\n"
	}

	return
}
