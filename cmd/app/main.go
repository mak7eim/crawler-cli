package main

import (
	"context"
	"crawler-cli/internal/config"
	"crawler-cli/internal/crawler"
	"crawler-cli/internal/fetcher"
	"crawler-cli/internal/logger"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		return
	}

	log, err := logger.New(cfg.LogFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logger error:", err)
		return
	}
	defer log.Close()

	log.Info("crawler start urls=%v depth=%d timeout=%s request_timeout=%s",
		cfg.URLs, cfg.MaxDepth, cfg.Timeout, cfg.RequestTimeout)

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ctx, cancel := context.WithTimeout(signalCtx, cfg.Timeout)
	defer cancel()

	httpFetcher := fetcher.New(cfg.RequestTimeout)
	craw := crawler.New(cfg.MaxDepth, httpFetcher, log)
	tree := craw.Run(ctx, cfg.URLs)

	if err := ctx.Err(); err != nil {
		log.Info("crawler stopped: %v", err)
	}

	if err := crawler.WriteJSON(cfg.Output, tree); err != nil {
		log.Error("write output: %v", err)
		fmt.Fprintln(os.Stderr, "output error:", err)
		return
	}

	log.Info("crawler finished top_level_pages=%d output=%s", len(tree), cfg.Output)
	fmt.Printf("done: %d top-level pages %s\n", len(tree), cfg.Output)
}
