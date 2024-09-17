package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// currently unused
//const (
//	providerConfig = `
//provider "meltcloud" {
//  endpoint     = "https://app.meltcloud.io"
//  organization = "deadbeef-0000-0000-0000-000000000000"
//  api_key      = "dummy"
//}
//`
//)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"meltcloud": providerserver.NewProtocol6WithError(New("test")()),
	}
)
