package main

import (
	"github.com/spf13/cobra"

	"github.com/m1dugh/gocrawler/internal/crawler/output/builtin"
)

func newRootCmd() *cobra.Command {
	var f flags
	reg := builtin.NewRegistry()

	root := &cobra.Command{
		Use:   "gocrawler [flags] <seed-url>...",
		Short: "A web crawler with basic and browser-based engines",
		Long: `gocrawler crawls from the seed urls, following the links that are in scope,
and writes one record per crawled page to the selected sinks.

A scope is required: give rules with --scope and/or --scope-rule. Rules are
regexes matched as a prefix of the url; a leading '!' excludes.`,
		Example: `  gocrawler --scope scope.txt https://example.com/
  gocrawler -S '^https://example\.com/' --sink file:path=crawl.jsonl https://example.com/
  gocrawler --config gocrawler.yaml`,
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.load(cmd, args)
			if err != nil {
				return err
			}

			runCfg, closeAll, err := build(cfg, reg)
			if err != nil {
				return err
			}
			defer closeAll()

			Run(runCfg)
			return nil
		},
	}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError{err} })
	root.SetVersionTemplate("gocrawler {{.Version}}\n")

	f.addConfig(root)
	f.addCrawl(root)

	root.AddCommand(newFieldsCmd(reg), newTransformersCmd(reg), newSinksCmd(reg))
	return root
}
