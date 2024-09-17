package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"strings"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &CustomizeUUIDInIPXEScriptFunction{}

type CustomizeUUIDInIPXEScriptFunction struct{}

func NewCustomizeUUIDInIPXEScriptFunction() function.Function {
	return &CustomizeUUIDInIPXEScriptFunction{}
}

func (f *CustomizeUUIDInIPXEScriptFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "customize_uuid_in_ipxe_script"
}

func (f *CustomizeUUIDInIPXEScriptFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Customizes the UUID of a Machine in an iPXE script",
		Description: "Some providers allow booting from a custom iPXE Script (i.e. Equinix Metal). However, most of these providers don't allow to know the UUID of the server beforehand, which would be necessary " +
			"to create and manage the Machine in Terraform. This function allows overriding the UUID that's used to identify the Machine on meltcloud so that it can be pre-registered and configured in Terraform.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "ipxe_script",
				Description: "The iPXE script to boot the Machine",
			},
			function.StringParameter{
				Name:        "uuid",
				Description: "Desired UUID of the Machine to be used on meltcloud",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *CustomizeUUIDInIPXEScriptFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var iPXEScript string
	var uuid string
	var customizedScript string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &iPXEScript, &uuid))

	customizedScript = strings.ReplaceAll(iPXEScript, "${uuid}", uuid)

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, customizedScript))
}
