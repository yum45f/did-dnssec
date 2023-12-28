/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	core "github.com/yum45f/did-dnssec/pkg"
	"golang.org/x/net/idna"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new zone file from a DID document",
	Long:  `Create a new zone file from a DID document.`,
	RunE:  handleCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("basefqdn", "b", "", "Base FQDN (e.g. example.com.)")
	createCmd.Flags().StringP("didjson", "d", "", "DID document file path")
	createCmd.Flags().StringP("out", "o", "", "Output file path")

	createCmd.MarkFlagRequired("basefqdn")
	createCmd.MarkFlagRequired("didjson")
	createCmd.MarkFlagRequired("out")
}

func handleCreate(cmd *cobra.Command, args []string) error {
	if err := cmd.ParseFlags(args); err != nil {
		return err
	}

	json, err := cmd.Flags().GetString("didjson")
	if err != nil {
		return err
	}
	if json == "" {
		return fmt.Errorf("didjson is required")
	}
	if fi, err := os.Stat(json); err != nil {
		return err
	} else if fi.IsDir() {
		return fmt.Errorf("didjson must be a file")
	}

	out, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}
	if out == "" {
		return fmt.Errorf("out is required")
	}

	base, err := cmd.Flags().GetString("basefqdn")
	if err != nil {
		return err
	}
	// check if base is a valid FQDN
	if _, err := idna.ToASCII(base); err != nil {
		return fmt.Errorf("basefqdn is not a valid FQDN")
	}

	f, err := os.Open(json)
	if err != nil {
		return err
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	doc, err := core.NewDocumentTreeFromJSON(bytes)
	if err != nil {
		return err
	}

	f, err = os.Create(out)
	if err != nil {
		return err
	}

	if err := doc.DumpRRs(f, base); err != nil {
		return err
	} else {
		fmt.Printf("Dumped to %s\n", out)
	}

	return nil
}
