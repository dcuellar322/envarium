package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"envarium.dev/envarium/internal/fleet/domain"
)

type fixedInventory struct{}

func (fixedInventory) Inventory(context.Context) ([]domain.Project, []domain.Environment, error) {
	return []domain.Project{{ID: "project-1"}, {ID: "project-2"}}, []domain.Environment{{ID: "env-one", ProjectID: "project-1", Availability: "available"}}, nil
}

type denyPolicy struct{}

func (denyPolicy) AuthorizeRemote(context.Context, string, string) error {
	return errors.New("disabled by policy")
}

func TestAgentResolvesCapabilityAgainstRegisteredTarget(t *testing.T) {
	controller := strings.Repeat("b", 64)
	for _, capability := range []domain.Capability{domain.CapabilityProjectOperate, domain.CapabilityEnvironmentManage} {
		operator := &operatorStub{}
		agent, err := NewAgentService(remoteIdentity(), fixedInventory{}, operator, []ControllerGrant{{Fingerprint: controller, Capabilities: []domain.Capability{capability}}})
		if err != nil {
			t.Fatal(err)
		}
		for _, request := range []domain.OperationRequest{
			{ProjectID: "project-1", EnvironmentID: "project-2"},
			{ProjectID: "project-2", EnvironmentID: "env-one"},
			{ProjectID: "env-one"},
			{ProjectID: "missing"},
		} {
			request.RequestID = "req"
			request.Action = domain.ActionStart
			request.ConfirmRisk = true
			if _, err := agent.Operate(context.Background(), controller, request); !errors.Is(err, ErrPermissionDenied) {
				t.Fatalf("accepted %#v: %v", request, err)
			}
		}
		valid := domain.OperationRequest{ProjectID: "project-1", RequestID: "valid", Action: domain.ActionStart, ConfirmRisk: true}
		if capability == domain.CapabilityEnvironmentManage {
			valid.EnvironmentID = "env-one"
		}
		if _, err := agent.Operate(context.Background(), controller, valid); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInventoryPolicyStopsAgentAndControllerReads(t *testing.T) {
	ctx := context.Background()
	controller := strings.Repeat("b", 64)
	agent, err := NewAgentService(remoteIdentity(), fixedInventory{}, &operatorStub{}, []ControllerGrant{{Fingerprint: controller, Capabilities: domain.KnownCapabilities}}, denyPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := agent.Identity(ctx, controller); !errors.Is(err, ErrPermissionDenied) {
		t.Fatal(err)
	}
	if _, err := agent.Snapshot(ctx, controller); !errors.Is(err, ErrPermissionDenied) {
		t.Fatal(err)
	}
	service, err := NewService(&memoryRepository{}, &peerStub{}, denyPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Register(ctx, registerRequest(), Actor{}); !errors.Is(err, ErrPermissionDenied) {
		t.Fatal(err)
	}
	if _, err := service.Probe(ctx, "machine", Actor{}); !errors.Is(err, ErrPermissionDenied) {
		t.Fatal(err)
	}
	if _, err := service.Snapshot(ctx, "machine"); !errors.Is(err, ErrPermissionDenied) {
		t.Fatal(err)
	}
}

type operateOnlyPolicy struct{}

func (operateOnlyPolicy) AuthorizeRemote(_ context.Context, capability, _ string) error {
	if capability == string(domain.CapabilityInventoryRead) {
		return errors.New("inventory disabled")
	}
	return nil
}

type countingPeer struct{ calls int }

func (p *countingPeer) Snapshot(context.Context, domain.Machine) (domain.Snapshot, error) {
	p.calls++
	return domain.Snapshot{}, nil
}
func (p *countingPeer) Operate(context.Context, domain.Machine, domain.OperationRequest) (domain.OperationReceipt, error) {
	return domain.OperationReceipt{}, nil
}
func TestOperationTargetReadCannotBypassInventoryPolicy(t *testing.T) {
	repository := &memoryRepository{machine: domain.Machine{ID: "machine", Enabled: true, Capabilities: domain.KnownCapabilities, GrantedCapabilities: domain.KnownCapabilities}}
	peer := &countingPeer{}
	service, _ := NewService(repository, peer, operateOnlyPolicy{})
	_, err := service.Operate(context.Background(), "machine", domain.OperationRequest{ProjectID: "project-1", Action: domain.ActionStart, ConfirmRisk: true}, Actor{})
	if !errors.Is(err, ErrPermissionDenied) || peer.calls != 0 {
		t.Fatalf("err=%v inventory calls=%d", err, peer.calls)
	}
}
