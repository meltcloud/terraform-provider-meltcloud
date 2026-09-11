package provider

import (
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Adds the lookup pair to a data source that carries attributes of its own.
func withLookupAttributes(attributes map[string]schema.Attribute, idDescription, nameDescription string) map[string]schema.Attribute {
	maps.Copy(attributes, lookupAttributes(idDescription, nameDescription))
	return attributes
}

// A data source finds its record by id or by name, and needs exactly one of them.
func lookupAttributes(idDescription, nameDescription string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: idDescription,
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				int64validator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name")),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: nameDescription,
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("id")),
			},
		},
	}
}
