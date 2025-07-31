package cmd

import (
	"os"
	"os/exec"
	"os/user"

	"github.com/pterm/pterm"
	"github.com/soarinferret/trmm-lam/internal/tacticalrmm"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "Install TRMM Agent on Linux",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {

		url, _ := cmd.Flags().GetString("url")
		apikey, _ := cmd.Flags().GetString("api-key")
		client, _ := cmd.Flags().GetInt("client")
		site, _ := cmd.Flags().GetInt("site")
		agentDlUrl, _ := cmd.Flags().GetString("agent-download-url")
		agentType, _ := cmd.Flags().GetString("type")
		force, _ := cmd.Flags().GetBool("force")

		if agentType != "server" && agentType != "workstation" {
			pterm.Error.Println("Invalid agent type. Must be server or workstation")
			return
		}

		rmm := tacticalrmm.New(url, apikey, agentDlUrl)

		if client == -1 || site == -1 {
			// interactively select client and site

			clients, err := rmm.GetClients()
			if err != nil {
				pterm.Error.Println("Failed to retrieve clients:", err)
				return
			}

			var options []string
			for _, c := range clients {
				options = append(options, c["name"].(string))
			}

			clientName, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()

			var selectedClient map[string]any
			for _, c := range clients {
				if c["name"].(string) == clientName {
					selectedClient = c
					client = int(c["id"].(float64))
					break
				}
			}

			options = []string{}
			for _, s := range selectedClient["sites"].([]interface{}) {
				options = append(options, s.(map[string]interface{})["name"].(string))
			}

			siteName, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()

			for _, s := range selectedClient["sites"].([]interface{}) {
				if s.(map[string]interface{})["name"].(string) == siteName {
					site = int(s.(map[string]interface{})["id"].(float64))
					break
				}
			}
		}

		script, err := rmm.GenerateInstallerScript(client, site, agentType)
		if err != nil {
			pterm.Error.Println("Failed to retrieve installer script:", err)
			return
		}
		//pterm.Info.Println("Script: ", script)

		f, err := os.Create("/tmp/trmm-installer.sh")
		if err != nil {
			pterm.Error.Println("Failed to create installer script on filesystem:", err)
			return
		}

		defer f.Close()

		_, err = f.WriteString(script)
		if err != nil {
			pterm.Error.Println("Failed to write installer script to filesystem:", err)
			return
		}

		pterm.Info.Println("Installer script written to /tmp/trmm-installer.sh")

		// check if running as root, if so, run the installer script
		usr, _ := user.Current()
		if usr.Uid == "0" && force {
			pterm.Info.Println("Running installer script...")
			exec.Command("bash", "/tmp/trmm-installer.sh")
			// wait for the command to finish
			err = exec.Command("bash", "/tmp/trmm-installer.sh").Run()
			if err != nil {
				pterm.Error.Println("Failed to run installer script:", err)
				return
			}
			pterm.Success.Println("Agent installed successfully!")

		} else {
			pterm.Info.Println("Run the installer script as root (or with sudo) to install the agent")
			pterm.Info.Println("Command: sudo bash /tmp/trmm-installer.sh")
		}

	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringP("api-key", "a", "", "API key for the Tactical RMM Server")
	installCmd.Flags().StringP("url", "u", "", "URL for the Tactical RMM API Server")
	installCmd.Flags().IntP("client", "c", -1, "Client ID")
	installCmd.Flags().IntP("site", "s", -1, "Site ID")
	installCmd.Flags().StringP("type", "t", "server", "Agent Type (can be server or workstation)")

	installCmd.Flags().BoolP("force", "f", false, "Don't prompt to run installer script")

	installCmd.MarkPersistentFlagRequired("api-key")
	installCmd.MarkPersistentFlagRequired("url")
}
