package main

import "strings"

type LineFilter struct {
	filterList []FilterItem
}

func (fl LineFilter) Filter(lines []LogLine) []LogLine {
	var result []LogLine
	for _, item := range lines {
		if fl.Match(item) {
			result = append(result, item)
		}
	}

	return result
}

func (fl LineFilter) Match(logLine LogLine) bool {
	for _, filter := range fl.filterList {
		if filter.Match(logLine) {
			return true
		}
	}

	return false
}

// type FilterItem func(line *LogLine) bool

type FilterItem interface {
	Match(LogLine) bool
}

type ByPrefix struct {
	Prefix string
}

func (f ByPrefix) Match(logLine LogLine) bool {
	return strings.HasPrefix(logLine.Message, f.Prefix)
}
