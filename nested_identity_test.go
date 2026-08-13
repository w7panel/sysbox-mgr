package main

import (
	"os"
	"testing"
)

func TestParseMappingMode(t *testing.T) {
	if _, err := parseMappingMode("standard-subid"); err != nil {
		t.Fatal(err)
	}
	if _, err := parseMappingMode("nested-identity"); err != nil {
		t.Fatal(err)
	}
	if _, err := parseMappingMode("auto"); err != nil {
		t.Fatal(err)
	}
}

func TestNestedIdentityMapValidation(t *testing.T) {
	if !isInitialUsernsMap([]byte("0 0 4294967295\n")) {
		t.Fatal("initial user namespace map not detected")
	}
	if !mapCoversContainerRange([]byte("0 309788672 65536\n"), 65536) {
		t.Fatal("valid nested map rejected")
	}
	if mapCoversContainerRange([]byte("0 309788672 65535\n"), 65536) {
		t.Fatal("short nested map accepted")
	}
}

func TestAutoMappingModeIsExplicitOnWire(t *testing.T) {
	mode, err := parseMappingMode("auto")
	if err != nil || !mode.Valid() {
		t.Fatalf("auto mode must resolve to a valid wire mode: %v (%d)", err, mode)
	}
}

func TestEffectiveCaps(t *testing.T) {
	if !hasEffectiveCaps([]byte("CapEff:\t00000000002000c0\n"), 6, 7, 21) {
		t.Fatal("required capabilities not detected")
	}
	if hasEffectiveCaps([]byte("CapEff:\t00000000000000c0\n"), 6, 7, 21) {
		t.Fatal("missing CAP_SYS_ADMIN accepted")
	}
}

func TestValidateNestedIdentityEnvironment(t *testing.T) {
	uidMap, err := os.ReadFile("/proc/self/uid_map")
	if err != nil {
		t.Fatal(err)
	}
	err = validateNestedIdentityEnvironment()
	if isInitialUsernsMap(uidMap) {
		if err == nil {
			t.Fatal("nested identity accepted the initial user namespace")
		}
		return
	}
	if err != nil {
		t.Fatalf("valid nested environment rejected: %v", err)
	}
}
