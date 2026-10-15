package wa

import (
	"testing"

	"go.mau.fi/whatsmeow/types"
)

func TestUsableCallLID(t *testing.T) {
	placeholder := types.NewJID("1099511693312", types.HiddenUserServer)
	real := types.NewJID("186835489816660", types.HiddenUserServer)
	phone := types.NewJID("559291276696", types.DefaultUserServer)
	if usableCallLID(placeholder) {
		t.Fatal("placeholder LID must be rejected")
	}
	if !usableCallLID(real) {
		t.Fatal("real LID must be accepted")
	}
	if usableCallLID(phone) {
		t.Fatal("phone JID is not a LID")
	}
}

func TestBrazilMobileAlternate(t *testing.T) {
	with9 := types.NewJID("5592991266696", types.DefaultUserServer)
	without := brazilMobileAlternate(with9)
	if without.User != "559291266696" {
		t.Fatalf("drop 9: %s", without.User)
	}
	back := brazilMobileAlternate(without)
	if back.User != "5592991266696" {
		t.Fatalf("insert 9: %s", back.User)
	}
	landline := types.NewJID("559232126696", types.DefaultUserServer)
	if !brazilMobileAlternate(landline).IsEmpty() {
		t.Fatal("landline should not have an alternate")
	}
}
