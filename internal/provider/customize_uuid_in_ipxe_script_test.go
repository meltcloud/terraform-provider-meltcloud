package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestCustomizeUUIDInIPXEScriptFunction_HappyPath(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
        output "test" {
          value = provider::meltcloud::customize_uuid_in_ipxe_script("script: $${uuid} and another $${uuid}", "my-uuid")
        }
        `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "script: my-uuid and another my-uuid"),
				),
			},
		},
	})
}

func TestCustomizeUUIDInIPXEScriptFunction_Null(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
        output "test" {
          value = provider::meltcloud::customize_uuid_in_ipxe_script(null, "my-uuid")
        }
        `,
				ExpectError: regexp.MustCompile(`argument must not be null`),
			},
		},
	})
}
