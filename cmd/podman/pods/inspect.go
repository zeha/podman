package pods

import (
	"context"
	"os"

	"github.com/containers/common/pkg/report"
	"github.com/containers/podman/v3/cmd/podman/common"
	"github.com/containers/podman/v3/cmd/podman/registry"
	"github.com/containers/podman/v3/cmd/podman/validate"
	"github.com/containers/podman/v3/libpod/define"
	"github.com/containers/podman/v3/pkg/domain/entities"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	inspectOptions = entities.PodInspectOptions{}
)

var (
	inspectDescription = `Display the configuration for a pod by name or id

	By default, this will render all results in a JSON array.`

	inspectCmd = &cobra.Command{
		Use:               "inspect [options] POD [POD...]",
		Short:             "Displays a pod configuration",
		Long:              inspectDescription,
		RunE:              inspect,
		ValidArgsFunction: common.AutocompletePods,
		Example:           `podman pod inspect podID`,
	}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: inspectCmd,
		Parent:  podCmd,
	})
	flags := inspectCmd.Flags()

	formatFlagName := "format"
	flags.StringVarP(&inspectOptions.Format, formatFlagName, "f", "json", "Format the output to a Go template or json")
	_ = inspectCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(define.InspectPodData{}))

	validate.AddLatestFlag(inspectCmd, &inspectOptions.Latest)
}

func inspect(cmd *cobra.Command, args []string) error {
	if len(args) < 1 && !inspectOptions.Latest {
		return errors.Errorf("you must provide the name or id of a running pod")
	}
	if len(args) > 0 && inspectOptions.Latest {
		return errors.Errorf("--latest and containers cannot be used together")
	}

	if !inspectOptions.Latest {
		inspectOptions.NameOrID = args[0]
	}
	responses, err := registry.ContainerEngine().PodInspect(context.Background(), inspectOptions)
	if err != nil {
		return err
	}

	if report.IsJSON(inspectOptions.Format) {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "     ")
		return enc.Encode(responses)
	}

	row := report.NormalizeFormat(inspectOptions.Format)

	t, err := report.NewTemplate("inspect").Parse(row)
	if err != nil {
		return err
	}

	w, err := report.NewWriterDefault(os.Stdout)
	if err != nil {
		return err
	}
	err = t.Execute(w, *responses)
	w.Flush()
	return err
}
