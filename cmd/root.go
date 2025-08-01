/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trmm-linux-installer",
	Short: "Installs the Tactical RMM Agent on Linux",
	Long:  ``,
	//PersistentPreRun: func(cmd *cobra.Command, args []string) {
	//},

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringP("agent-download-url", "D", "https://github.com/soarinferret/rmmagent-builder/", "Manually specify the agent download URL")
}

func pExit(s string, err error) {
	if err != nil {
		pterm.Error.Println(s, err)
		os.Exit(1)
	}
}
