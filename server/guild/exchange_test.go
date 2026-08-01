package guild

import (
	"reflect"
	"testing"
)

func TestClearUsesOrderedOnePassWithoutRedistribution(t *testing.T) {
	catalog, _ := LoadCatalog([]byte(phase0Catalog))
	members := []MemberStock{
		{AccountID: "018f0000-0000-4000-8000-000000000004", Produces: "compliance", Consumes: "libraries", ReceivedUnits: 99_950},
		{AccountID: "018f0000-0000-4000-8000-000000000002", Produces: "hype", Consumes: "revenue"},
		{AccountID: "018f0000-0000-4000-8000-000000000001", Produces: "revenue", Consumes: "compliance", AvailableUnits: 401},
		{AccountID: "018f0000-0000-4000-8000-000000000003", Produces: "libraries", Consumes: "hype"},
		{AccountID: "018f0000-0000-4000-8000-000000000005", Produces: "hype", Consumes: "revenue"},
	}
	got, events, err := Clear(catalog, members, 100_000)
	if err != nil {
		t.Fatal(err)
	}
	// Offered=200, split 100/100. The first consumer can take 100; the
	// second can take 100. The cap-limited case is separately asserted below.
	if got[0].AvailableUnits != 201 || len(events) != 1 || !reflect.DeepEqual(events[0].Allocations,
		[]Allocation{{AccountID: members[1].AccountID, Units: 100}, {AccountID: members[4].AccountID, Units: 100}}) {
		t.Fatalf("got=%+v events=%+v", got, events)
	}

	limited := []MemberStock{
		{AccountID: "018f0000-0000-4000-8000-000000000001", Produces: "revenue", Consumes: "compliance", AvailableUnits: 401},
		{AccountID: "018f0000-0000-4000-8000-000000000002", Produces: "hype", Consumes: "revenue", ReceivedUnits: 99_950},
		{AccountID: "018f0000-0000-4000-8000-000000000003", Produces: "hype", Consumes: "revenue"},
	}
	got, events, err = Clear(catalog, limited, 100_000)
	if err != nil || got[0].AvailableUnits != 251 || len(events) != 1 || !reflect.DeepEqual(events[0].Allocations,
		[]Allocation{{AccountID: limited[1].AccountID, Units: 50}, {AccountID: limited[2].AccountID, Units: 100}}) {
		t.Fatalf("limited got=%+v events=%+v err=%v", got, events, err)
	}
}

func TestClearAbsentLinkAndNPCFallback(t *testing.T) {
	catalog, _ := LoadCatalog([]byte(phase0Catalog))
	member := MemberStock{AccountID: "018f0000-0000-4000-8000-000000000001", Produces: "libraries", Consumes: "hype", AvailableUnits: 1000}
	got, events, err := Clear(catalog, []MemberStock{member}, 100_000)
	if err != nil || got[0].AvailableUnits != 1000 || len(events) != 0 {
		t.Fatalf("guild got=%+v events=%v err=%v", got, events, err)
	}
	npc, event, err := ClearNPC(catalog, member, 100_000)
	if err != nil || npc.AvailableUnits != 880 || npc.ReceivedUnits != 120 || event == nil || !event.NPC {
		t.Fatalf("npc=%+v event=%+v err=%v", npc, event, err)
	}
}

func TestClearKeepsUnincorporatedMembersInert(t *testing.T) {
	catalog, _ := LoadCatalog([]byte(phase0Catalog))
	members := []MemberStock{
		{AccountID: "018f0000-0000-4000-8000-000000000001", Produces: "libraries", Consumes: "hype", AvailableUnits: 100},
		{AccountID: "018f0000-0000-4000-8000-000000000002"},
	}
	got, clearings, err := Clear(catalog, members, 100_000)
	if err != nil || len(got) != 2 || len(clearings) != 0 || got[1] != members[1] {
		t.Fatalf("states=%+v clearings=%+v err=%v", got, clearings, err)
	}
	invalid := append([]MemberStock(nil), members...)
	invalid[1].AvailableUnits = 1
	if _, _, err := Clear(catalog, invalid, 100_000); err == nil {
		t.Fatal("partially initialized faction stock was accepted")
	}
}
