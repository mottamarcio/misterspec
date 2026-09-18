package internalcmd

import (
	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/vcs"
)

// NewCommitsSinceFileCmd builds "misterspec internal commits-since-file
// --path <path>" (029-spec-wrap-up-docs, contracts/commits-since-file.md).
// It calls vcs.CommitsSinceFileAdded directly — no logic of its own
// beyond argument parsing and JSON shaping, mirroring
// fingerprint.go's own exact pattern.
func NewCommitsSinceFileCmd() *cobra.Command {
	var dir, path string

	cmd := &cobra.Command{
		Use:           "commits-since-file",
		Short:         "Return every commit since a file was first added, on the current branch",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			commits, available, err := vcs.CommitsSinceFileAdded(proj.Root, path)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			out := make([]map[string]any, 0, len(commits))
			for _, c := range commits {
				out = append(out, map[string]any{
					"hash":        c.Hash,
					"subject":     c.Subject,
					"author_date": c.AuthorDate,
				})
			}

			return WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"available": available,
				"commits":   out,
			})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&path, "path", "", "repo-relative path whose \"added\" commit anchors the range")
	cmd.MarkFlagRequired("path")
	return cmd
}
