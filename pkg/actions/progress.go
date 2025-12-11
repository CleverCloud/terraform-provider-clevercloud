package actions

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

func Progress(res *action.InvokeResponse, message string, args ...any) {
	res.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf(message, args...),
	})
}

func ProgressWrapper(res *action.InvokeResponse) func(message string, args ...any) {
	return func(message string, args ...any) {
		Progress(res, message, args...)
	}
}
