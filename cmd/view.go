package cmd

import (
	"github.com/pterm/pterm"
	"github.com/soarinferret/trmm-lam/internal/tacticalrmm"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:     "view",
	Aliases: []string{"v"},
	Short:   "View client and site ids on Tactical RMM",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {

		url, _ := cmd.Flags().GetString("url")
		apikey, _ := cmd.Flags().GetString("api-key")
		agentDlUrl, _ := cmd.Flags().GetString("agent-download-url")

		rmm := tacticalrmm.New(url, apikey, agentDlUrl)

		client := -1
		site := -1

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

		pterm.Info.Println("Selected Client ID: ", client)
		pterm.Info.Println("Selected Site ID: ", site)

	},
}

func init() {
	rootCmd.AddCommand(viewCmd)

	viewCmd.Flags().StringP("api-key", "a", "", "API key for the Tactical RMM Server")
	viewCmd.Flags().StringP("url", "u", "", "URL for the Tactical RMM API Server")

	viewCmd.MarkPersistentFlagRequired("api-key")
	viewCmd.MarkPersistentFlagRequired("url")
}
