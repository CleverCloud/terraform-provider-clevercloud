package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// SyncTCPRedirection reconciles the application TCP redirection with the plan.
// state is the previously managed redirection (nil on Create).
// It returns the redirection to persist, with its allocated port, or nil when
// no redirection is planned.
func SyncTCPRedirection(ctx context.Context, cc *client.Client, organisation, applicationID string, plan, state *TCPRedirection, diags *diag.Diagnostics) *TCPRedirection {
	// nothing planned and nothing managed: leave remote redirections untouched
	if plan == nil && state == nil {
		return nil
	}

	redirsRes := tmp.GetTCPRedirections(ctx, cc, organisation, applicationID)
	if redirsRes.HasError() {
		diags.AddError("failed to get application TCP redirections", redirsRes.Error().Error())
		return state
	}

	var result *TCPRedirection
	for _, redir := range *redirsRes.Payload() {
		if plan != nil && redir.Namespace == plan.Namespace.ValueString() {
			// redirection already exists, adopt its allocated port
			result = &TCPRedirection{Namespace: pkg.FromStr(redir.Namespace), Port: pkg.FromI(redir.Port)}
			continue
		}

		// only remove redirections previously managed by this resource
		if state != nil && redir.Namespace == state.Namespace.ValueString() {
			deleteRes := tmp.DeleteTCPRedirection(ctx, cc, organisation, applicationID, redir.Port, redir.Namespace)
			if deleteRes.HasError() {
				diags.AddError(fmt.Sprintf("failed to delete TCP redirection on namespace %q", redir.Namespace), deleteRes.Error().Error())
			}
		}
	}

	if plan != nil && result == nil {
		createRes := tmp.CreateTCPRedirection(ctx, cc, organisation, applicationID, tmp.CreateTCPRedirectionRequest{
			Namespace: plan.Namespace.ValueString(),
		})
		if createRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to create TCP redirection on namespace %q", plan.Namespace.ValueString()), createRes.Error().Error())
			return nil
		}
		result = &TCPRedirection{Namespace: plan.Namespace, Port: pkg.FromI(createRes.Payload().Port)}
	}

	return result
}

// ReadTCPRedirection refreshes the managed TCP redirection from the API to
// detect drift. Redirections created outside Terraform are not adopted.
func ReadTCPRedirection(ctx context.Context, cc *client.Client, organisation, applicationID string, state *TCPRedirection, diags *diag.Diagnostics) *TCPRedirection {
	if state == nil {
		return nil
	}

	redirsRes := tmp.GetTCPRedirections(ctx, cc, organisation, applicationID)
	if redirsRes.HasError() {
		diags.AddError("failed to get application TCP redirections", redirsRes.Error().Error())
		return state
	}

	for _, redir := range *redirsRes.Payload() {
		if redir.Namespace == state.Namespace.ValueString() {
			return &TCPRedirection{Namespace: pkg.FromStr(redir.Namespace), Port: pkg.FromI(redir.Port)}
		}
	}

	// redirection was removed outside Terraform
	return nil
}

// UseStatePortWhenNamespaceUnchanged keeps the port value from state as long
// as the redirection namespace does not change: changing the namespace
// recreates the redirection, so a new port will be allocated.
func UseStatePortWhenNamespaceUnchanged() planmodifier.Int64 {
	return tcpRedirectionPortModifier{}
}

type tcpRedirectionPortModifier struct{}

func (tcpRedirectionPortModifier) Description(context.Context) string {
	return "Keep the allocated port from state while the namespace is unchanged"
}

func (m tcpRedirectionPortModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (tcpRedirectionPortModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, res *planmodifier.Int64Response) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	namespacePath := req.Path.ParentPath().AtName("namespace")

	var planNamespace, stateNamespace types.String
	res.Diagnostics.Append(req.Plan.GetAttribute(ctx, namespacePath, &planNamespace)...)
	res.Diagnostics.Append(req.State.GetAttribute(ctx, namespacePath, &stateNamespace)...)
	if res.Diagnostics.HasError() {
		return
	}

	if planNamespace.Equal(stateNamespace) {
		res.PlanValue = req.StateValue
	}
}
