package provider

import (
	"strings"
	"terraform-provider-meltcloud/internal/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBGPValuesKeepKnownPasswordVersions(t *testing.T) {
	asn, multiplier := int64(65012), int64(5)
	instance := &client.NestInstance{
		BGPLocalASN: &asn,
		BGPPeers: []client.BGPPeer{
			{IP: "10.0.10.2", ASN: 65000, BFD: true, BFDDetectMultiplier: &multiplier},
			{IP: "10.0.10.1", ASN: 65000},
		},
	}
	known := &NestInstanceBGPModel{Peers: []NestInstanceBGPPeerModel{
		{IP: types.StringValue("10.0.10.1"), PasswordWOVersion: types.Int64Value(2)},
	}}

	bgp := bgpValues(instance, known)

	if bgp.LocalASN.ValueInt64() != asn || len(bgp.Peers) != 2 {
		t.Fatalf("got %+v; want local ASN %d and two peers", bgp, asn)
	}
	first, second := bgp.Peers[0], bgp.Peers[1]
	if first.IP.ValueString() != "10.0.10.1" || first.PasswordWOVersion.ValueInt64() != 2 || !second.PasswordWOVersion.IsNull() {
		t.Errorf("got %v with %v, then %v; want the known peer first with its password version", first.IP, first.PasswordWOVersion, second.PasswordWOVersion)
	}
	if !first.PasswordWO.IsNull() {
		t.Errorf("got password %v; want none, as it is write-only", first.PasswordWO)
	}
	if second.BFD == nil || second.BFD.DetectMultiplier.ValueInt64() != multiplier || !second.BFD.ReceiveInterval.IsNull() {
		t.Errorf("bfd: got %+v; want detect multiplier %d only", second.BFD, multiplier)
	}
	if first.BFD != nil {
		t.Errorf("bfd: got %+v; want none", first.BFD)
	}
}

func TestBGPValuesKeepKnownOrder(t *testing.T) {
	instance := &client.NestInstance{BGPPeers: []client.BGPPeer{{IP: "10.0.10.1"}, {IP: "10.0.10.2"}, {IP: "10.0.10.3"}}}
	known := &NestInstanceBGPModel{Peers: []NestInstanceBGPPeerModel{{IP: types.StringValue("10.0.10.3")}, {IP: types.StringValue("10.0.10.1")}}}

	var got []string
	for _, peer := range bgpValues(instance, known).Peers {
		got = append(got, peer.IP.ValueString())
	}
	if strings.Join(got, ",") != "10.0.10.3,10.0.10.1,10.0.10.2" {
		t.Errorf("got %v; want the known peers in their order, then the others", got)
	}
}

func TestBGPValuesWithoutBGP(t *testing.T) {
	if bgp := bgpValues(&client.NestInstance{BGPPeers: []client.BGPPeer{}}, nil); bgp != nil {
		t.Errorf("got %+v; want no bgp block", bgp)
	}
}

func TestBGPPeersInput(t *testing.T) {
	planned := NestInstanceBGPPeerModel{
		IP: types.StringValue("10.0.10.1"), ASN: types.Int64Value(65000), PasswordWO: types.StringNull(),
		Multihop: types.BoolValue(true), BFD: &NestInstanceBFDModel{ReceiveInterval: types.Int64Value(300)},
	}
	configured := planned
	configured.PasswordWO = types.StringValue("secret")

	peer := bgpPeersInput([]NestInstanceBGPPeerModel{planned}, []NestInstanceBGPPeerModel{configured})[0]
	if peer.Password == nil || *peer.Password != "secret" || !peer.Multihop || !peer.BFD || *peer.BFDReceiveInterval != 300 || peer.BFDTransmitInterval != nil {
		t.Errorf("got %+v; want the configured password, multihop, BFD with only the receive interval", peer)
	}
}

func TestBGPChanged(t *testing.T) {
	peer := func(ip string, passwordVersion int64) NestInstanceBGPPeerModel {
		p := NestInstanceBGPPeerModel{IP: types.StringValue(ip), ASN: types.Int64Value(65000), Multihop: types.BoolValue(false), PasswordWOVersion: types.Int64Null()}
		if passwordVersion != 0 {
			p.PasswordWOVersion = types.Int64Value(passwordVersion)
		}
		return p
	}
	bgp := func(peers ...NestInstanceBGPPeerModel) NestInstanceBGPModel {
		return NestInstanceBGPModel{LocalASN: types.Int64Value(65010), Peers: peers}
	}
	withBFD := peer("10.0.10.2", 0)
	withBFD.BFD = &NestInstanceBFDModel{ReceiveInterval: types.Int64Value(300)}

	tests := []struct {
		name         string
		state, plan  NestInstanceBGPModel
		wantReplaced bool
	}{
		{"same peers in another order", bgp(peer("10.0.10.1", 1), peer("10.0.10.2", 0)), bgp(peer("10.0.10.2", 0), peer("10.0.10.1", 1)), false},
		{"password version the state does not know", bgp(peer("10.0.10.1", 0)), bgp(peer("10.0.10.1", 1)), false},
		{"changed password version", bgp(peer("10.0.10.1", 1)), bgp(peer("10.0.10.1", 2)), true},
		{"password removed", bgp(peer("10.0.10.1", 1)), bgp(peer("10.0.10.1", 0)), true},
		{"another peer", bgp(peer("10.0.10.1", 0)), bgp(peer("10.0.10.3", 0)), true},
		{"added peer", bgp(peer("10.0.10.1", 0)), bgp(peer("10.0.10.1", 0), peer("10.0.10.2", 0)), true},
		{"BFD turned on", bgp(peer("10.0.10.2", 0)), bgp(withBFD), true},
		{"changed local ASN", bgp(), NestInstanceBGPModel{LocalASN: types.Int64Value(65011)}, true},
	}
	for _, tt := range tests {
		if got := bgpChanged(tt.state, tt.plan); got != tt.wantReplaced {
			t.Errorf("%s: got %v; want %v", tt.name, got, tt.wantReplaced)
		}
	}
}
