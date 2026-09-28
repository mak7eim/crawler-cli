package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	URLs           []string
	MaxDepth       int
	Timeout        time.Duration
	RequestTimeout time.Duration
	Output         string
	LogFile        string
}

func Parse(args []string) (Config, error) {
	var urls string
	var depth int
	var timeout time.Duration
	var requestTimeout time.Duration
	var output string
	var logFile string

	flags := flag.NewFlagSet("crawler-cli", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&urls, "urls", "", "comma-separated start URLs")
	flags.IntVar(&depth, "depth", 1, "maximum crawl depth")
	flags.DurationVar(&timeout, "timeout", 2*time.Minute, "total crawler timeout")
	flags.DurationVar(&requestTimeout, "request-timeout", 10*time.Second, "single request timeout")
	flags.StringVar(&output, "output", "output.json", "output JSON file")
	flags.StringVar(&logFile, "log", "crawler.log", "log file")

	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}

	if strings.TrimSpace(urls) == "" {
		return Config{}, errors.New("urls cannot be empty")
	}
	if depth < 0 {
		return Config{}, errors.New("depth cannot be negative")
	}
	if timeout <= 0 {
		return Config{}, errors.New("timeout must be positive")
	}
	if requestTimeout <= 0 {
		return Config{}, errors.New("request-timeout must be positive")
	}
	if strings.TrimSpace(output) == "" {
		return Config{}, errors.New("output cannot be empty")
	}
	if strings.TrimSpace(logFile) == "" {
		return Config{}, errors.New("log cannot be empty")
	}

	urlList := strings.Split(urls, ",")

	for i := range urlList {
		urlList[i] = strings.TrimSpace(urlList[i])

		if err := ValidatorURL(urlList[i]); err != nil {
			return Config{}, fmt.Errorf("invalid URL %q: %w", urlList[i], err)
		}
	}

	config := Config{
		URLs:           urlList,
		MaxDepth:       depth,
		Timeout:        timeout,
		RequestTimeout: requestTimeout,
		Output:         output,
		LogFile:        logFile,
	}

	return config, nil
}

func ValidatorURL(rawURL string) error {
	urlParsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	if urlParsed.Scheme != "http" && urlParsed.Scheme != "https" {
		return errors.New("URL need be http or https")
	}

	if urlParsed.Host == "" {
		return errors.New("URL host is missing")
	}

	return nil
}
