package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/soarinferret/trmm-lam/internal/tacticalrmm"
	"github.com/spf13/cobra"
)

// https://stackoverflow.com/a/33853856/13335339
func downloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"u"},
	Short:   "Update TRMM Agent on Linux",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		agentDlUrl, _ := cmd.Flags().GetString("agent-download-url")

		rmm := tacticalrmm.New("", "", agentDlUrl)

		url, err := rmm.GetAgentDownloadUrl()
		if err != nil {
			pterm.Error.Println("Failed to retrieve agent download URL:", err)
			return
		}

		agentName := "tacticalagent"

		fname, err := exec.LookPath(agentName)
		if err == nil {
			fname, _ = filepath.Abs(fname)
		}
		if err != nil {
			pterm.Error.Println("Failed to find agent binary:", err)
			os.Exit(1)
		}

		// get the current agent version
		v := exec.Command(fname, "version")
		out, err := v.Output()
		if err != nil {
			pterm.Error.Println("Failed to get agent version:", err)
			os.Exit(1)
		}
		//out = strings.TrimSpace(out)

		pterm.Info.Println("Current agent version:", string(out))

		// Latest available version
		latest, err := rmm.GetLatestAgentVersion()
		if err != nil {
			pterm.Error.Println("Failed to get latest agent version:", err)
			os.Exit(1)
		}

		pterm.Info.Println("Latest agent version:", latest)

		if strings.Contains(latest, strings.TrimSpace(string(out))) {
			pterm.Info.Println("Agent is already up to date")
			os.Exit(0)
		}

		// move the old agent to a backup
		err = os.Rename(fname, fname+".old")

		// replace the agent
		err = downloadFile(fname, url)
		if err != nil {
			pterm.Error.Println("Failed to download agent:", err)
			os.Exit(1)
		}
		// make the agent executable
		err = os.Chmod(fname, 0755)
		if err != nil {
			pterm.Error.Println("Failed to make agent executable:", err)
			os.Exit(1)
		}

		// restart the service
		c := exec.Command("systemctl", "restart", "tacticalagent")
		err = c.Run()
		if err != nil {
			pterm.Error.Println("Failed to restart agent service:", err)
			os.Exit(1)
		}

		pterm.Info.Println("Agent updated successfully")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
