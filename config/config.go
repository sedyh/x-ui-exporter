package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
)

var CLIConfig CLI

type VersionFlag string

type CLI struct {
	Ip                 string      `name:"metrics-ip" help:"Ip to listen on" default:"0.0.0.0" env:"METRICS_IP"`
	Port               string      `name:"metrics-port" help:"Port to listen on" default:"9090" env:"METRICS_PORT"`
	ProtectedMetrics   bool        `name:"metrics-protected" help:"Whether metrics are protected by basic auth" default:"false" env:"METRICS_PROTECTED"`
	MetricsUsername    string      `name:"metrics-username" help:"Username for metrics if protected by basic auth" default:"metricsUser" env:"METRICS_USERNAME"`
	MetricsPassword    string      `name:"metrics-password" help:"Password for metrics if protected by basic auth" default:"MetricsVeryHardPassword" env:"METRICS_PASSWORD"`
	UpdateInterval     int         `name:"update-interval" help:"Interval for metrics update in seconds" default:"30" env:"UPDATE_INTERVAL"`
	TimeZone           string      `name:"timezone" help:"Timezone used in the application" default:"UTC" env:"TIMEZONE"`
	BaseURL            string      `name:"panel-base-url" help:"Panel base URL" env:"PANEL_BASE_URL"`
	ApiUsername        string      `name:"panel-username" help:"Panel username" env:"PANEL_USERNAME"`
	ApiPassword        string      `name:"panel-password" help:"Panel password" env:"PANEL_PASSWORD"`
	InsecureSkipVerify bool        `name:"insecure-skip-verify" help:"Skip SSL certificate verification (INSECURE)" default:"false" env:"INSECURE_SKIP_VERIFY"`
	ConfigFile         string      `name:"config-file" help:"Path to a YAML configuration file" env:"CONFIG_FILE"`
	Version            VersionFlag `name:"version" help:"Print version information and quit"`
}

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println("3X-UI Exporter (Fork)")
	fmt.Printf("Version:\t %s\n", vars["version"])
	fmt.Printf("Commit:\t %s\n", vars["commit"])
	fmt.Printf("Github (Marzban): https://github.com/kutovoys/marzban-exporter\n")
	fmt.Printf("GitHub (3X-UI Fork): https://github.com/hteppl/3x-ui-exporter\n")
	app.Exit(0)
	return nil
}

func Parse(version, commit string) {
	// Parse CLI flags first
	ctx := kong.Parse(&CLIConfig,
		kong.Name("x-ui-exporter"),
		kong.Description("A command-line application for exporting 3X-UI metrics."),
		kong.Vars{
			"version": version,
			"commit":  commit,
		},
	)

	// Check if a config file is provided
	if CLIConfig.ConfigFile != "" {
		// Load YAML configuration
		yamlConfig, err := LoadYAMLConfig(CLIConfig.ConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading YAML configuration file: %v\n", err)
			ctx.Exit(3)
		}

		// Use YAML config instead of CLI flags
		CLIConfig = yamlConfig.ToCLI()
	}
	// Validate the final configuration
	if err := validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ctx.Exit(2)
	}
}

func removeTrailingSlash(s string) string {
	if strings.HasSuffix(s, "/") {
		return strings.TrimSuffix(s, "/")
	}
	return s
}

func validate() error {
	if CLIConfig.BaseURL == "" {
		return errors.New("x-ui-exporter: error: --panel-base-url must be provided")
	}
	if CLIConfig.ApiUsername == "" {
		return errors.New("x-ui-exporter: error: --panel-username must be provided")
	}
	if CLIConfig.ApiPassword == "" {
		return errors.New("x-ui-exporter: error: --panel-password must be provided")
	}
	CLIConfig.BaseURL = removeTrailingSlash(CLIConfig.BaseURL)
	return nil
}
