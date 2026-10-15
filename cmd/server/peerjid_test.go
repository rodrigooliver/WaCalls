package main

import "testing"

func TestPeerJID(t *testing.T) {
	lid, err := peerJID("186835489816660@lid")
	if err != nil || lid.String() != "186835489816660@lid" {
		t.Fatalf("lid: %v %v", lid, err)
	}
	phone, err := peerJID("+55 92 9127-6696")
	if err != nil || phone.String() != "559291276696@s.whatsapp.net" {
		t.Fatalf("phone: %v %v", phone, err)
	}
	if _, err := peerJID("1099511693312@lid"); err == nil {
		t.Fatal("placeholder lid should be rejected")
	}
}
