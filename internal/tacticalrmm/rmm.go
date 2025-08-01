package tacticalrmm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/soarinferret/trmm-lam/internal/meshcentral"
)

type TacticalRMM struct {
	url              string
	apiKey           string
	agentDownloadUrl string
}

func New(url string, apiKey string, agentDl string) *TacticalRMM {
	return &TacticalRMM{
		url:              url,
		apiKey:           apiKey,
		agentDownloadUrl: agentDl,
	}
}

func (rmm *TacticalRMM) ensureApiUrl() error {
	if rmm.url == "" {
		return errors.New("URL is not set")
	}
	if rmm.apiKey == "" {
		return errors.New("API Key is not set")
	}

	return nil
}

func (rmm *TacticalRMM) get(url string) (response string, err error) {
	err = rmm.ensureApiUrl()
	if err != nil {
		return response, err
	}

	// Create a new get request
	req, err := http.NewRequest("GET", rmm.url+url, nil)
	if err != nil {
		return response, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", rmm.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	return string(body), nil
}

func (rmm *TacticalRMM) GetSettings() (settings map[string]any, err error) {

	response, err := rmm.get("/core/settings/")
	if err != nil {
		return settings, err
	}

	err = json.Unmarshal([]byte(response), &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}

func (rmm *TacticalRMM) GetClients() (clients []map[string]any, err error) {

	response, err := rmm.get("/clients/")
	if err != nil {
		return clients, err
	}

	err = json.Unmarshal([]byte(response), &clients)
	if err != nil {
		return clients, err
	}

	return clients, nil
}

func (rmm *TacticalRMM) getMeshDownloadUrl() (url string, err error) {
	settings, err := rmm.GetSettings()
	if err != nil {
		return "", err
	}

	token := settings["mesh_token"].(string)
	username := settings["mesh_username"].(string)
	mesh_site := settings["mesh_site"].(string)
	device_group := settings["mesh_device_group"].(string)

	wsUrl, err := meshcentral.GetMeshWsUrl(mesh_site, username, token)
	if err != nil {
		return "", err
	}

	id, err := meshcentral.GetMeshDeviceGroupId(wsUrl, device_group)
	if err != nil {
		return "", err
	}

	meshUrl := mesh_site + "/meshagents?id=" + id + "&installflags=2&meshinstall=6"

	return meshUrl, nil
}

func parseGithubUrl(url string) (owner string, repo string, err error) {
	// should come in format of https://github.com/OWNER/REPO

	// split url by '/'
	parts := strings.Split(url, "/")

	// check if parts is less than 4, then return error
	// check if parts[2] is not equal to 'github.com', then return error
	// return parts[3] and parts[4]
	if len(parts) < 4 {
		return "", "", errors.New("Invalid URL")
	}

	if parts[2] != "github.com" {
		return "", "", errors.New("Invalid URL")
	}

	return parts[3], parts[4], nil
}

func (rmm *TacticalRMM) GetAgentToken(client int, site int) (token string, err error) {
	// generate a fake windows agent so we can get the token

	data := fmt.Sprintf(`
		{"installMethod":"manual",
		"client":%d,
		"site":%d,
		"expires":1,
		"agenttype":"server",
		"power":0,
		"rdp":0,
		"ping":0,
		"goarch":"amd64",
		"api":"%s",
		"fileName":"rmm.exe",
		"plat":"windows"}`, client, site, rmm.url)

	reader := strings.NewReader(data)

	req, err := http.NewRequest("POST", rmm.url+"/agents/installer/", reader)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", rmm.apiKey)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	// parse token from the cmd
	cmd := response["cmd"].(string)
	parts := strings.Split(cmd, "--auth ")

	return parts[1], nil
}

func (rmm *TacticalRMM) GetAgentDownloadUrl() (url string, err error) {
	type Release struct {
		TagName string `json:"tag_name"`
	}

	// check url does not contain 'github.com', then return the url
	if rmm.agentDownloadUrl != "" && !strings.Contains(rmm.agentDownloadUrl, "github.com") {
		return rmm.agentDownloadUrl, nil
	}

	// parse url to get owner and repo
	owner, repo, err := parseGithubUrl(rmm.agentDownloadUrl)

	// query api.github.com for latest release
	ghApiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(ghApiUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", err
	}

	// https://github.com/SoarinFerret/rmmagent-builder/releases/download/v2.9.0/rmmagent-linux-amd64
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/rmmagent-linux-amd64", owner, repo, release.TagName), nil
}

func (rmm *TacticalRMM) GetLatestAgentVersion() (version string, err error) {
	type Release struct {
		TagName string `json:"tag_name"`
	}

	agentUrl := rmm.agentDownloadUrl
	if rmm.agentDownloadUrl != "" && !strings.Contains(rmm.agentDownloadUrl, "github.com") {
		agentUrl = "https://github.com/amidaware/rmmagent"
	}

	owner, repo, err := parseGithubUrl(agentUrl)
	if err != nil {
		return "", err
	}

	ghApiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(ghApiUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", err
	}

	return release.TagName, nil
}

func (rmm *TacticalRMM) GenerateInstallerScript(client int, site int, agentType string) (script string, err error) {
	agentDL, err := rmm.GetAgentDownloadUrl()
	if err != nil {
		return script, err
	}

	meshDL, err := rmm.getMeshDownloadUrl()
	if err != nil {
		return script, err
	}

	token, err := rmm.GetAgentToken(client, site)
	if err != nil {
		return script, err
	}

	script = LINUX_INSTALL_SCRIPT
	script = strings.ReplaceAll(script, "agentDLChange", agentDL)
	script = strings.ReplaceAll(script, "meshDLChange", meshDL)
	script = strings.ReplaceAll(script, "apiURLChange", rmm.url)
	script = strings.ReplaceAll(script, "tokenChange", token)
	script = strings.ReplaceAll(script, "clientIDChange", fmt.Sprintf("%d", client))
	script = strings.ReplaceAll(script, "siteIDChange", fmt.Sprintf("%d", site))
	script = strings.ReplaceAll(script, "agentTypeChange", agentType)

	return script, nil
}
