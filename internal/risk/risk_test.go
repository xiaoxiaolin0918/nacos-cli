package risk

import "testing"

func TestHighRiskGate(t *testing.T) {
	p := HighRisk()
	err := p.Check(false, "config.publish")
	if err == nil || !IsConfirmation(err) {
		t.Fatalf("%v", err)
	}
	g := err.(GateResult)
	if g.Action != "config.publish" || g.Hint == "" {
		t.Fatalf("%#v", g)
	}
	if err := p.Check(true, "config.publish"); err != nil {
		t.Fatal(err)
	}
	if ExitConfirmationRequired != 10 {
		t.Fatalf("exit=%d", ExitConfirmationRequired)
	}
	if err := ReadOnly().Check(false, "doctor"); err != nil {
		t.Fatal(err)
	}
	if err := Mutation().Check(false, "auth.login"); err != nil {
		t.Fatal(err)
	}
}
