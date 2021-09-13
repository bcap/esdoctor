package main

import (
	"context"
	"os"

	"esdoctor/client"
	"esdoctor/diagnosis"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	// TODO: use a context cancellable by ^C
	if err := Command().ExecuteContext(context.TODO()); err != nil {
		log.Errorf("Execution failed: %v", err)
		os.Exit(1)
	}
}

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:           "esdoctor ES_ENDPOINT",
		Short:         "runs a series of diagnostics over an Elasticsearch cluster",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
	}

	var verbosity int

	cmd.PersistentFlags().CountVarP(
		&verbosity, "verbosity", "v",
		"Controls loggging verbosity. Can be specified multiple times (eg -vv) or a count can "+
			"be passed in (--verbosity=2). Defaults to print error messages. "+
			"See also --quiet",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		setupLogging(verbosity)
		endpoint := args[0]
		return run(cmd.Context(), endpoint)
	}

	return &cmd
}

func run(ctx context.Context, endpoint string) error {
	client, err := client.New(endpoint)
	if err != nil {
		return err
	}

	diagnostics, err := diagnosis.Diagnose(ctx, client)
	if err != nil {
		return err
	}

	diagnostics.Print(os.Stdout)
	return nil
}

func setupLogging(verbosity int) {
	log.SetLevel(log.WarnLevel)
	if verbosity == 1 {
		log.SetLevel(log.InfoLevel)
	} else if verbosity == 2 {
		log.SetLevel(log.DebugLevel)
	} else if verbosity >= 3 {
		log.SetLevel(log.TraceLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		PadLevelText:  true,
		ForceColors:   true,
	})
	log.SetOutput(os.Stderr)
}
