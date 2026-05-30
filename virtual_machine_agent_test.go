package proxmox

// Most QGA helpers are exercised in virtual_machine_test.go because they share
// the same mockClient()/mockConfig wiring and run in the same test binary.
//
// This file is the dedicated home for tests that target the
// virtual_machine_agent.go helpers specifically — added here so coverage and
// future agent-only tests have an obvious place to live.
//
// Helpers NOT covered by gock-only unit tests:
//   - TermWebSocket / VNCWebSocket: require a live websocket dialer; see the
//     file-level comment in virtual_machine.go.
//   - Any binary-streaming QGA helpers (e.g. FileWriteStream) — n/a in this
//     version of the package; if added later they should ship integration
//     coverage instead.

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

// vmAgent101 returns a *VirtualMachine wired to vmid 101 on node1 for QGA
// tests — same shape as vm101() in virtual_machine_test.go but lives next to
// the agent tests for clarity.
func vmAgent101() *VirtualMachine {
	return &VirtualMachine{client: mockClient(), Node: "node1", VMID: 101}
}

// Smoke test the AgentPing wrapper from this file's perspective so the agent
// suite has at least one direct entry point — every other QGA helper is
// already exercised in virtual_machine_test.go.
func TestAgent_PingSmoke(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	assert.Nil(t, vmAgent101().AgentPing(context.Background()))
}
