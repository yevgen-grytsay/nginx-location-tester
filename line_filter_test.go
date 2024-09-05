package main

import "testing"

func TestEmptyFilter(t *testing.T) {
	filter := LineFilter{}

	lines := []LogLine{
		{Message: "message #1"},
		{Message: "message #2"},
	}

	result := filter.Filter(lines)

	if len(result) > 0 {
		t.Fatalf(`len(result) = %d, want 0, error`, len(result))
	}
}

func TestMatch(t *testing.T) {
	filter := LineFilter{filterList: []FilterItem{
		ByPrefix{Prefix: "prefix: "},
	}}

	lines := []LogLine{
		{Message: "prefix: #1"},
		{Message: "message #2"},
	}

	result := filter.Filter(lines)

	if len(result) != 1 {
		t.Fatalf(`len(result) = %d, want 1, error`, len(result))
	}

	expected := LogLine{Message: "prefix: #1"}
	if result[0] != expected {
		t.Fatalf(`expected %v, got %v, error`, expected, result[0])
	}
}
