package netrc

import (
	"strings"
	"testing"
)

// oldFindMachine is the pre-index implementation, kept here to prove the
// indexed lookup returns identical results.
func (n *Netrc) oldFindMachine(name string) (m *Machine) {
	var def *Machine
	for _, m = range n.machines {
		if m.Name == name {
			return m
		}
		if m.IsDefault() {
			def = m
		}
	}
	if def == nil {
		return nil
	}
	return def
}

func TestIndexMatchesScan(t *testing.T) {
	inputs := []string{
		"machine a login l1 password p1\nmachine b login l2 password p2\n",
		"machine a login l1 password p1\nmachine a login DUP password DUP\n",
		"machine a login l1 password p1\ndefault login dl password dp\n",
		"default login dl password dp\n",
		"machine a login l1 password p1\nmachine b login l2 password p2\ndefault login dl password dp\n",
		"",
	}
	queries := []string{"a", "b", "absent", "", "A"}

	for i, in := range inputs {
		nrc, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("input %d: %v", i, err)
		}
		for _, q := range queries {
			got, want := nrc.FindMachine(q), nrc.oldFindMachine(q)
			if got != want {
				t.Errorf("input %d query %q: indexed=%v scan=%v", i, q, describe(got), describe(want))
			}
		}
	}
}

func describe(m *Machine) string {
	if m == nil {
		return "<nil>"
	}
	return m.Name + "/" + m.Login
}

// The index must stay correct across NewMachine and RemoveMachine.
func TestIndexAfterMutations(t *testing.T) {
	nrc, err := Parse(strings.NewReader("machine a login l1 password p1\nmachine b login l2 password p2\n"))
	if err != nil {
		t.Fatal(err)
	}

	check := func(stage string) {
		t.Helper()
		for _, q := range []string{"a", "b", "c", "absent", ""} {
			if got, want := nrc.FindMachine(q), nrc.oldFindMachine(q); got != want {
				t.Errorf("%s: query %q: indexed=%v scan=%v", stage, q, describe(got), describe(want))
			}
		}
	}

	check("initial")

	nrc.NewMachine("c", "l3", "p3", "")
	check("after NewMachine c")

	nrc.RemoveMachine("a")
	check("after RemoveMachine a")
	if nrc.FindMachine("a") != nil {
		t.Error("removed machine still resolves")
	}

	nrc.RemoveMachine("absent")
	check("after removing a nonexistent machine")
}
