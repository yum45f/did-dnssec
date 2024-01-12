/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	core "github.com/yum45f/did-dnssec/pkg"
)

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: handleResolve,
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(resolveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resolveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resolveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	resolveCmd.Flags().StringP("out", "o", "", "Output json file path")
}

func handleResolve(cmd *cobra.Command, args []string) error {
	out, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}

	node, err := core.Resolve(args[0])
	if err != nil {
		return err
	}

	node.Print()

	bytes, err := node.JSON()
	if err != nil {
		return err
	}

	if out != "" {
		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err = f.Write(bytes); err != nil {
			return err
		}
	} else {
		cmd.Println(string(bytes))
	}

	return nil
}
