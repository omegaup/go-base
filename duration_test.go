package base

import (
	"testing"
	"time"
)

func TestDurationMarshalRoundtrip(t *testing.T) {
	d1 := Duration(time.Duration(30) * time.Second)
	serialized, err := d1.MarshalJSON()
	if err != nil {
		t.Fatalf(err.Error())
	}
	var d2 Duration
	if err = d2.UnmarshalJSON(serialized); err != nil {
		t.Fatalf(err.Error())
	}
	if d1 != d2 {
		t.Errorf("expected %v got %v", d1.String(), d2.String())
	}
}

func TestDurationNil(t *testing.T) {
	var d, zeroValue Duration
	if err := d.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf(err.Error())
	}
	if zeroValue != d {
		t.Errorf("expected %v got %v", zeroValue.String(), d.String())
	}
}

func TestDurationNumber(t *testing.T) {
	var d Duration
	if err := d.UnmarshalJSON([]byte("1.5")); err != nil {
		t.Fatalf(err.Error())
	}
	if 1.5 != d.Seconds() {
		t.Errorf("expected 1.5s got %v", d.String())
	}
	if 1500 != d.Milliseconds() {
		t.Errorf("expected 1500ms got %v", d.String())
	}
}

func TestDurationInvalidNumber(t *testing.T) {
	var d, zeroValue Duration
	if err := d.UnmarshalJSON([]byte("1.x")); err != nil {
		t.Fatalf(err.Error())
	}
	if zeroValue != d {
		t.Errorf("expected %v got %v", zeroValue.String(), d.String())
	}
}

func TestDurationInvalidString(t *testing.T) {
	var d Duration
	if err := d.UnmarshalJSON([]byte("\"1")); err == nil {
		t.Errorf("Expected an error, but got nil")
	}
	if err := d.UnmarshalJSON([]byte("\"1.x\"")); err == nil {
		t.Errorf("Expected an error, but got nil")
	}
}

func TestDurationComparison(t *testing.T) {
	d1 := Duration(time.Duration(1) * time.Second)
	d2 := Duration(time.Duration(2) * time.Second)
	if d1 != MinDuration(d1, d2) {
		t.Errorf("expected %v got %v", d1, MinDuration(d1, d2))
	}
	if d1 != MinDuration(d2, d1) {
		t.Errorf("expected %v got %v", d1, MinDuration(d2, d1))
	}
	if d2 != MaxDuration(d1, d2) {
		t.Errorf("expected %v got %v", d2, MaxDuration(d1, d2))
	}
	if d2 != MaxDuration(d2, d1) {
		t.Errorf("expected %v got %v", d2, MaxDuration(d2, d1))
	}
}
