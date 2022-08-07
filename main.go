package main

import (
	"flag"
	"fmt"

	"io/ioutil"
	"math/rand"

	"net/url"
	"os"
	"time"

	"mymodule/pkg/pusher"
	"mymodule/pkg/querier"

	_ "github.com/cortexproject/cortex/integration"
	"github.com/cortexproject/cortex/pkg/cortex"
	"github.com/cortexproject/cortex/pkg/util/flagext"
	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/pkg/errors"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/storage/remote"

	yaml "gopkg.in/yaml.v2"
)

const (
	configFileOption = "config.file"
	configExpandENV  = "config.expand-env"
)

func main() {
	var (
		cfg cortex.Config
	)

	configFile, _ := parseConfigFileParameter(os.Args[1:])

	flagext.RegisterFlags(&cfg)

	LoadConfig(configFile, &cfg)

	// Ignore -config.file and -config.expand-env here, since it was already parsed, but it's still present on command line.
	flagext.IgnoredFlag(flag.CommandLine, configFileOption, "Configuration file to load.")
	_ = flag.CommandLine.Bool(configExpandENV, false, "Expands ${var} or $var in config according to the values of the environment variables.")

	usage := flag.CommandLine.Usage
	flag.CommandLine.Usage = func() { /* don't do anything by default, we will print usage ourselves, but only when requested. */ }
	flag.CommandLine.Init(flag.CommandLine.Name(), flag.ContinueOnError)

	err := flag.CommandLine.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		// Print available parameters to stdout, so that users can grep/less it easily.
		flag.CommandLine.SetOutput(os.Stdout)
		usage()
		os.Exit(2)
	} else if err != nil {
		fmt.Fprintln(flag.CommandLine.Output(), "Run with -help to get list of available parameters")
		os.Exit(2)
	}

	u, err := url.Parse("http://cortex:9009/api/v1/push")

	if err != nil {
		fmt.Println("Error parsing URL for Remote Write")
		os.Exit(2)
	}

	w, err := remote.NewWriteClient("test1", &remote.ClientConfig{
		URL:     &config.URL{URL: u},
		Timeout: model.Duration(time.Minute),
		HTTPClientConfig: config.HTTPClientConfig{
			FollowRedirects: true,
			EnableHTTP2:     true,
		},
		RetryOnRateLimit: true,
	})
	if err != nil {
		fmt.Println("Error creating WriteClient")
		os.Exit(2)
	}

	u, err = url.Parse("http://cortex:9009/prometheus/api/v1/read")
	if err != nil {
		fmt.Println("Error parsing URL for Remote Read")
		os.Exit(2)
	}

	c, err := remote.NewReadClient("test2", &remote.ClientConfig{
		URL:     &config.URL{URL: u}, // requires url struct
		Timeout: model.Duration(time.Minute),
		HTTPClientConfig: config.HTTPClientConfig{
			FollowRedirects: true,
			EnableHTTP2:     true,
		},
	})
	if err != nil {
		fmt.Println("Error creating readClient")
		os.Exit(2)
	}

	cfg.ExternalQueryable = querier.NewQueryable(c)
	cfg.ExternalPusher = pusher.NewPusher(w)

	util_log.InitLogger(&cfg.Server)

	err = cfg.Validate(util_log.Logger)

	util_log.CheckFatal("error validating config", err)

	// Initialise seed for randomness usage.
	rand.Seed(time.Now().UnixNano())

	t, err := cortex.New(cfg)

	util_log.CheckFatal("running cortex", err)

	err = t.Run()

	util_log.CheckFatal("running cortex", err)
}

func LoadConfig(filename string, cfg *cortex.Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading config file")
	}

	err = yaml.UnmarshalStrict(buf, cfg)
	if err != nil {
		return errors.Wrap(err, "Error parsing config file")
	}

	return nil
}

// Parse -config.file and -config.expand-env option via separate flag set, to avoid polluting default one and calling flag.Parse on it twice.
func parseConfigFileParameter(args []string) (configFile string, expandEnv bool) {
	// ignore errors and any output here. Any flag errors will be reported by main flag.Parse() call.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)

	// usage not used in these functions.
	fs.StringVar(&configFile, configFileOption, "", "")
	fs.BoolVar(&expandEnv, configExpandENV, false, "")

	// Try to find -config.file and -config.expand-env option in the flags. As Parsing stops on the first error, eg. unknown flag, we simply
	// try remaining parameters until we find config flag, or there are no params left.
	// (ContinueOnError just means that flag.Parse doesn't call panic or os.Exit, but it returns error, which we ignore)
	for len(args) > 0 {
		_ = fs.Parse(args)
		args = args[1:]
	}

	return
}