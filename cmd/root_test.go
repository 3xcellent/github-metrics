package cmd_test

import (
	"bytes"
	"context"

	"github.com/spf13/cobra"
)

var noBoardsYAML = []byte(`---
API:
  BaseURL: https://enterprise.github.com/api/v3
  Token: token
IncludeHeaders: true
GroupName: An 3xcellent Team
Owner: 3xcellent
LoginNames:
  - 3xcellent
`)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}
func executeCommandWithContext(ctx context.Context, root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err = root.ExecuteContext(ctx)
	return buf.String(), err
}
func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	c, err = root.ExecuteC()
	return c, buf.String(), err
}
