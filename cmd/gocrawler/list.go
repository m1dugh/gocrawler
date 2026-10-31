package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// providedBy lists the fields of a registered transformer, built without
// options. It reports "?" for one that cannot be built that way.
func providedBy(reg *output.Registry, name string) (provides, requires []string, ok bool) {
	t, err := reg.NewTransformer(name, nil)
	if err != nil {
		return nil, nil, false
	}
	return t.Provides(), t.Requires(), true
}

func joinOrNone(list []string) string {
	if len(list) == 0 {
		return "none"
	}
	return strings.Join(list, ",")
}

func newTransformersCmd(reg *output.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "transformers",
		Short: "List the available transformers and the fields they provide",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tPROVIDES\tREQUIRES\tDEFAULT\tDESCRIPTION")
			for _, spec := range reg.Transformers() {
				provides, requires, ok := providedBy(reg, spec.Name)
				p, r := "?", "?"
				if ok {
					p, r = joinOrNone(provides), joinOrNone(requires)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", spec.Name, p, r, yesNo(spec.Default), spec.Description)
			}
			return w.Flush()
		},
	}
}

func newFieldsCmd(reg *output.Registry) *cobra.Command {
	var f flags
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "List the fields a record can hold and whether they are active",
		Long: `List every field provided by a registered transformer, and whether the
transformers selected by the config file and flags make it active.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.load(cmd, args)
			if err != nil {
				return err
			}
			steps, err := activeTransformers(cfg.Transformers, reg)
			if err != nil {
				return err
			}
			active := map[string]bool{}
			for _, t := range steps {
				active[t.Name()] = true
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "FIELD\tPROVIDED BY\tACTIVE")
			for _, spec := range reg.Transformers() {
				provides, _, _ := providedBy(reg, spec.Name)
				for _, field := range provides {
					fmt.Fprintf(w, "%s\t%s\t%s\n", field, spec.Name, yesNo(active[spec.Name]))
				}
			}
			return w.Flush()
		},
	}
	f.addConfig(cmd)
	f.addTransformers(cmd)
	return cmd
}

func newSinksCmd(reg *output.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "sinks",
		Short: "List the available sinks and their options",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
			for _, spec := range reg.Sinks() {
				fmt.Fprintf(w, "%s\t%s\n", spec.Name, spec.Description)
				for _, o := range spec.Options {
					req := ""
					if o.Required {
						req = " (required)"
					}
					fmt.Fprintf(w, "    %s\t%s%s\n", o.Name, o.Description, req)
				}
				fmt.Fprintf(w, "    format\tserialization format: %s (default: %s)\n",
					strings.Join(reg.Formats(), ", "), output.DefaultFormat)
			}
			return w.Flush()
		},
	}
}
