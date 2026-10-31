package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/m1dugh/gocrawler/internal/crawler/config"
)

// flags holds the values of the command-line flags. They are merged over
// the config file by apply: a flag that was set overrides the matching key
// of the file, except seeds and transformer selections which are added to
// it.
type flags struct {
	config       string
	seedsFile    string
	scope        string
	scopeRules   []string
	engine       string
	browserURL   string
	browserPath  string
	headless     bool
	browserFlags []string
	headers      []string
	timeout      time.Duration
	workers      int

	transformers          []string
	noTransformers        []string
	noDefaultTransformers bool

	sinks  []string
	fields []string
}

func (f *flags) addConfig(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.config, "config", "", "YAML config file; flags override its keys")
}

func (f *flags) addTransformers(cmd *cobra.Command) {
	fl := cmd.Flags()
	fl.StringArrayVar(&f.transformers, "transformer", nil, "enable a transformer: name[:key=value,...] (repeatable)")
	fl.StringArrayVar(&f.noTransformers, "no-transformer", nil, "disable a default transformer (repeatable)")
	fl.BoolVar(&f.noDefaultTransformers, "no-default-transformers", false, "start from an empty pipeline")
}

func (f *flags) addCrawl(cmd *cobra.Command) {
	fl := cmd.Flags()
	fl.StringVarP(&f.seedsFile, "seeds-file", "l", "", "file of seed urls, one per line ('-' reads stdin)")
	fl.StringVar(&f.scope, "scope", "", "scope rules file")
	fl.StringArrayVarP(&f.scopeRules, "scope-rule", "S", nil, "inline scope rule (repeatable)")
	fl.StringVar(&f.engine, "engine", "", "crawl engine: http or browser (default http)")
	fl.StringVar(&f.browserURL, "browser-url", "", "remote debugging url of an existing browser (implies --engine browser)")
	fl.StringVar(&f.browserPath, "browser-path", "", "browser binary to spawn (default: auto-detect)")
	fl.BoolVar(&f.headless, "headless", true, "run a spawned browser headless")
	fl.StringArrayVar(&f.browserFlags, "browser-flag", nil, "extra browser flag (repeatable)")
	fl.StringArrayVarP(&f.headers, "header", "H", nil, "request header \"Name: value\" (repeatable)")
	fl.DurationVar(&f.timeout, "timeout", 0, "per-fetch timeout (default 30s)")
	fl.IntVar(&f.workers, "workers", 0, "concurrent crawl workers (default 4)")
	f.addTransformers(cmd)
	fl.StringArrayVar(&f.sinks, "sink", nil, "select a sink: name[:key=value,...] (repeatable, default stdout)")
	fl.StringSliceVar(&f.fields, "fields", nil, "fields every sink writes (default: all provided)")
}

// load returns the config file's content, or an empty config when no file
// was given, with the flags applied on top.
func (f *flags) load(cmd *cobra.Command, args []string) (*config.Config, error) {
	cfg := &config.Config{}
	if f.config != "" {
		var err error
		if cfg, err = config.Load(f.config); err != nil {
			return nil, err
		}
	}
	if err := f.apply(cmd, args, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (f *flags) apply(cmd *cobra.Command, args []string, cfg *config.Config) error {
	changed := cmd.Flags().Changed

	cfg.Seeds = append(cfg.Seeds, args...)
	if changed("seeds-file") {
		seeds, err := readSeeds(f.seedsFile, cmd.InOrStdin())
		if err != nil {
			return err
		}
		cfg.Seeds = append(cfg.Seeds, seeds...)
	}

	if changed("scope") {
		cfg.Scope.File = f.scope
	}
	if changed("scope-rule") {
		cfg.Scope.Rules = f.scopeRules
	}
	if changed("engine") {
		cfg.Engine = f.engine
	}
	if changed("browser-url") {
		cfg.Browser.URL = f.browserURL
	}
	if changed("browser-path") {
		cfg.Browser.Path = f.browserPath
	}
	if changed("headless") {
		cfg.Browser.Headless = &f.headless
	}
	if changed("browser-flag") {
		cfg.Browser.Flags = f.browserFlags
	}
	if changed("header") {
		cfg.Headers = f.headers
	}
	if changed("timeout") {
		cfg.Timeout = f.timeout
	}
	if changed("workers") {
		cfg.Workers = f.workers
	}

	for _, spec := range f.transformers {
		name, opts, err := config.ParseSpec(spec)
		if err != nil {
			return usagef("--transformer: %v", err)
		}
		cfg.Transformers.Enable = append(cfg.Transformers.Enable, config.Transformer{Name: name, Options: opts})
	}
	cfg.Transformers.Disable = append(cfg.Transformers.Disable, f.noTransformers...)
	if changed("no-default-transformers") {
		defaults := !f.noDefaultTransformers
		cfg.Transformers.Defaults = &defaults
	}

	if changed("sink") {
		cfg.Output.Sinks = nil
		for _, spec := range f.sinks {
			name, opts, err := config.ParseSpec(spec)
			if err != nil {
				return usagef("--sink: %v", err)
			}
			cfg.Output.Sinks = append(cfg.Output.Sinks, config.Sink{Type: name, Options: opts})
		}
	}
	if changed("fields") {
		cfg.Output.Fields = f.fields
		for i := range cfg.Output.Sinks {
			cfg.Output.Sinks[i].Fields = nil
		}
	}
	return nil
}

// readSeeds reads one url per line from path, or from stdin for "-".
// Blank lines are skipped.
func readSeeds(path string, stdin io.Reader) ([]string, error) {
	r := stdin
	if path != "-" {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		r = file
	}

	var seeds []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			seeds = append(seeds, line)
		}
	}
	return seeds, sc.Err()
}
