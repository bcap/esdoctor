package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"esdoctor/client"
	"esdoctor/diagnosis"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	// TODO: use a context cancellable by ^C
	cmd := Command()
	if err := cmd.ExecuteContext(context.TODO()); err != nil {
		if cmd.SilenceUsage {
			log.Errorf("Execution failed: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "Usage error: %v\n", err)
		}
		os.Exit(1)
	}
}

func Command() *cobra.Command {

	cmd := cobra.Command{
		Use:           "esdoctor <ELASTICSEARCH_HTTP_ENDPOINT>",
		Short:         "runs a series of diagnostics over an Elasticsearch cluster",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
	}

	cmd.Long = "" +
		"Runs a series of diagnostics over an Elasticsearch cluster, while printing the results " +
		"to stdout.\n\n" +
		"Printing format can be controlled with the -f|--format flag:\n" +
		"- text:      prints comments, one each line. This is the default format and intended for humans. " +
		"Levels can be controlled with the -A, -i, -s, -s and -w flags\n" +
		"- json:      prints all comments in json format. Can be used for machine consumption\n" +
		"- json-dump: dumps the whole diagnostics state as json, including supporting data, processed " +
		"data and comments. Useful for getting a detailed view of the cluster state and metadata. " +
		"ATTENTION: This dump will be quite extensive due to the sheer amount of data and the fact that " +
		"multiple paths to the same data will be generated"

	cmd.Example = strings.Join([]string{
		"1. Runs diagnostics, printing only warning comments",
		"  esdoctor https://some.address:9200 -w",
		"2. Runs diagnostics, printing summary, advice and warning comments",
		"  esdoctor https://some.address:9200 -saw",
		"3. Runs diagnostics, printing all comments while logging more verbosily",
		"  esdoctor https://some.address:9200 -A -vv",
		"4. Runs diagnostics, printing comments in json format",
		"  esdoctor https://some.address:9200 -f json",
		"5. Runs diagnostics, dumping the whole diagnosis state as json",
		"  esdoctor https://some.address:9200 -f json-dump",
	}, "\n")

	var verbosity int
	cmd.PersistentFlags().CountVarP(
		&verbosity, "verbosity", "v",
		"Controls loggging verbosity. Can be specified multiple times (eg -vv) or a count "+
			"can be passed in (--verbosity=2). Defaults to print error and warning messages",
	)

	var format string
	cmd.PersistentFlags().StringVarP(
		&format, "format", "f", "text",
		"Format in which to print results. Can be: text, json or json-dump",
	)

	var jsonFormat bool
	cmd.PersistentFlags().BoolVarP(
		&jsonFormat, "json", "j", false,
		"Same as --format=json",
	)

	var jsonDumpFormat bool
	cmd.PersistentFlags().BoolVarP(
		&jsonDumpFormat, "json-dump", "J", false,
		"Same as --format=json-dump",
	)

	var infoLevel bool
	cmd.PersistentFlags().BoolVarP(
		&infoLevel, "info", "i", false,
		"Also prints informational comments. Only relevant for the text format",
	)

	var summaryLevel bool
	cmd.PersistentFlags().BoolVarP(
		&summaryLevel, "summary", "s", false,
		"Also prints summary comments. Only relevant for the text format",
	)

	var adviceLevel bool
	cmd.PersistentFlags().BoolVarP(
		&adviceLevel, "advice", "a", false,
		"Also prints advice comments. Only relevant for the text format",
	)

	var warningLevel bool
	cmd.PersistentFlags().BoolVarP(
		&warningLevel, "warning", "w", false,
		"Also prints warning comments. Only relevant for the text format",
	)

	var allTypes bool
	cmd.PersistentFlags().BoolVarP(
		&allTypes, "all", "A", false,
		"Print all comments regardless of type. Only relevant for the text format. "+
			"Also check the --info, --summary, --advice and --warning flags",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var writer diagnosis.CommentWriter
		if format == "json" || jsonFormat {
			writer = diagnosis.NewJSONCommentWriter(os.Stdout, false)
		} else if format == "json-dump" || jsonDumpFormat {
			writer = diagnosis.NewJSONCommentWriter(os.Stdout, true)
		} else if format == "text" {
			var types []diagnosis.CommentType
			if !allTypes {
				types = []diagnosis.CommentType{}
				if infoLevel {
					types = append(types, diagnosis.Info)
				}
				if summaryLevel {
					types = append(types, diagnosis.Summary)
				}
				if adviceLevel {
					types = append(types, diagnosis.Advice)
				}
				if warningLevel {
					types = append(types, diagnosis.Warning)
				}
			}

			if types != nil && len(types) == 0 {
				return errors.New(
					"need to specify at least one level of comments to be printed when running with text format. " +
						"Use -A for all comments or a combination of the -i, -s, -a and -w flags",
				)
			}
			writer = diagnosis.NewTextCommentWriter(os.Stdout, types, true)
		} else {
			return fmt.Errorf("unrecognized format %q", format)
		}

		// From now forward any failures are execution failures and not usage errors. Setting this
		// will suprress printing the error as an usage error
		cmd.SilenceUsage = true

		setupLogging(verbosity)
		endpoint := args[0]
		clientOpts := []client.Option{}
		if log.IsLevelEnabled(log.TraceLevel) {
			clientOpts = append(clientOpts, client.WithBodyLogging())
		}
		client, err := client.New(endpoint, clientOpts...)
		if err != nil {
			return err
		}

		diagnosis, err := diagnosis.Diagnose(cmd.Context(), client, diagnosis.WithOutput(writer))

		if diagnosis != nil {
			diagnosis.Comments()
		}

		return err
	}

	return &cmd
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
