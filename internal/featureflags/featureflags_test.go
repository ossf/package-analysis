package featureflags

import (
	"reflect"
	"testing"
)

func resetRegistry() {
	flagRegistry = make(map[string]*FeatureFlag)
}

func TestFlagDefault_True(t *testing.T) {
	resetRegistry()
	ff := new("TestFlag", true)
	if !ff.Enabled() {
		t.Error("Enabled() = false; want true")
	}
}

func TestFlagDefault_False(t *testing.T) {
	resetRegistry()
	ff := new("TestFlag", false)
	if ff.Enabled() {
		t.Error("Enabled() = true; want false")
	}
}

func TestFlagUpdate_SingleFlag(t *testing.T) {
	resetRegistry()
	ff := new("TestFlag", false)
	Update("TestFlag")

	if !ff.Enabled() {
		t.Error("Enabled() = false; want true")
	}
}

func TestFlagUpdate_SingleFlagOff(t *testing.T) {
	resetRegistry()
	ff := new("TestFlag", true)
	Update("-TestFlag")

	if ff.Enabled() {
		t.Error("Enabled() = true; want false")
	}
}

func TestFlagUpdate_MultiFlags(t *testing.T) {
	resetRegistry()
	new("TestFlag1", false)
	new("TestFlag2", true)
	new("TestFlag3", false)
	Update("TestFlag1,-TestFlag2,TestFlag3")
	want := map[string]bool{
		"TestFlag1": true,
		"TestFlag2": false,
		"TestFlag3": true,
	}
	if got := State(); !reflect.DeepEqual(want, got) {
		t.Errorf("State() = %v; want %v", got, want)
	}
}

func TestFlagUpdate_MultiFlags_EmptyString(t *testing.T) {
	resetRegistry()
	new("TestFlag1", false)
	new("TestFlag2", true)
	new("TestFlag3", false)
	Update("")
	want := map[string]bool{
		"TestFlag1": false,
		"TestFlag2": true,
		"TestFlag3": false,
	}
	if got := State(); !reflect.DeepEqual(want, got) {
		t.Errorf("State() = %v; want %v", got, want)
	}
}

func TestFlagUpdate_Error(t *testing.T) {
	resetRegistry()
	err := Update("TestFlag")
	if err == nil {
		t.Errorf("Update() = nil; want an error")
	}
}
