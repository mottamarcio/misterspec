package internalcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
)

// entityTypeNames maps a --type flag value to ids.EntityType. Package-
// local rather than added to internal/ids, since no prior feature
// needed a name-to-type parser — only the inverse, EntityType.String()
// (research.md).
var entityTypeNames = map[string]ids.EntityType{
	"program":   ids.Program,
	"feature":   ids.Feature,
	"spec":      ids.Spec,
	"task":      ids.Task,
	"knowledge": ids.Knowledge,
	"learning":  ids.Learning,
}

// NewChildrenCmd builds "misterspec internal children <id> [--type T]"
// (docs/architecture-specification.md §9.5). It calls
// operations.Children directly — no logic of its own beyond argument/
// flag parsing and JSON shaping (spec.md FR-009).
func NewChildrenCmd() *cobra.Command {
	var dir, typeFilter string

	cmd := &cobra.Command{
		Use:           "children <id>",
		Short:         "Return structural children",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var filterType *ids.EntityType
			if typeFilter != "" {
				t, ok := entityTypeNames[typeFilter]
				if !ok {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: unrecognized --type %q", ErrInvalidArgument, typeFilter))
				}
				filterType = &t
			}

			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			children, err := operations.Children(proj.Root, proj.Config, args[0], filterType)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			result := make([]map[string]any, 0, len(children))
			for _, c := range children {
				result = append(result, map[string]any{
					"id":   c.ID.String(),
					"type": c.Type.String(),
				})
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{"children": result})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&typeFilter, "type", "", "filter children to this entity type")
	return cmd
}
