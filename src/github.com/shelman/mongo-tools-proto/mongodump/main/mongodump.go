package main

import (
	"fmt"
	"github.com/shelman/mongo-tools-proto/common/db"
	"github.com/shelman/mongo-tools-proto/common/log"
	commonopts "github.com/shelman/mongo-tools-proto/common/options"
	"github.com/shelman/mongo-tools-proto/common/util"
	"github.com/shelman/mongo-tools-proto/mongodump"
	"github.com/shelman/mongo-tools-proto/mongodump/options"
	"os"
)

func main() {
	// initialize command-line opts
	opts := commonopts.New("mongodump", "0.0.1", "<options>")

	inputOpts := &options.InputOptions{}
	opts.AddOptions(inputOpts)
	outputOpts := &options.OutputOptions{}
	opts.AddOptions(outputOpts)

	_, err := opts.Parse()
	if err != nil {
		fmt.Printf("error parsing command line options: %v\n\n", err)
		opts.Help = true //TODO HACK HACK HACK, change options package at some point
		opts.PrintHelp()
		os.Exit(2)
	}

	// print help, if specified
	if opts.PrintHelp() {
		return
	}

	// print version, if specified
	if opts.PrintVersion() {
		return
	}

	// init logger
	log.InitToolLogger(opts.Verbosity)

	// create a session provider to connect to the db
	sessionProvider, err := db.InitSessionProvider(opts)
	if err != nil {
		util.Panicf("error initializing database session: %v", err)
	}

	dump := mongodump.MongoDump{
		ToolOptions:     opts,
		OutputOptions:   outputOpts,
		InputOptions:    inputOpts,
		SessionProvider: sessionProvider,
	}

	_ = dump.Dump()

}
