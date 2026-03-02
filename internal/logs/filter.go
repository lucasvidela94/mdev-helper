// Package logs provides filtering capabilities for log entries.
package logs

import (
	"regexp"
	"strings"
)

// Filter is the interface for log entry filters.
type Filter interface {
	// Match returns true if the log entry passes the filter.
	Match(entry LogEntry) bool
}

// TextFilter filters log entries by text content.
type TextFilter struct {
	// Pattern is the text to search for (case-insensitive).
	Pattern string
	// Invert inverts the match (exclude instead of include).
	Invert bool
}

// Match returns true if the log entry matches the text filter.
func (f *TextFilter) Match(entry LogEntry) bool {
	if f.Pattern == "" {
		return true
	}

	pattern := strings.ToLower(f.Pattern)
	message := strings.ToLower(entry.Message)
	tag := strings.ToLower(entry.Tag)
	package_ := strings.ToLower(entry.Package)

	// Search in message, tag, and package
	matched := strings.Contains(message, pattern) ||
		strings.Contains(tag, pattern) ||
		strings.Contains(package_, pattern)

	if f.Invert {
		return !matched
	}
	return matched
}

// RegexFilter filters log entries using regular expressions.
type RegexFilter struct {
	// Pattern is the compiled regular expression.
	Pattern *regexp.Regexp
	// Invert inverts the match.
	Invert bool
}

// NewRegexFilter creates a new regex filter from a pattern string.
func NewRegexFilter(pattern string) (*RegexFilter, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexFilter{Pattern: re}, nil
}

// Match returns true if the log entry matches the regex filter.
func (f *RegexFilter) Match(entry LogEntry) bool {
	if f.Pattern == nil {
		return true
	}

	text := entry.Message + " " + entry.Tag + " " + entry.Package
	matched := f.Pattern.MatchString(text)

	if f.Invert {
		return !matched
	}
	return matched
}

// LevelFilter filters log entries by severity level.
type LevelFilter struct {
	// MinLevel is the minimum level to include.
	MinLevel LogLevel
}

// Match returns true if the log entry's level is >= the minimum level.
func (f *LevelFilter) Match(entry LogEntry) bool {
	return entry.Level >= f.MinLevel
}

// TagFilter filters log entries by tag.
type TagFilter struct {
	// Tag is the tag to match (case-insensitive).
	Tag string
	// Invert inverts the match.
	Invert bool
}

// Match returns true if the log entry's tag matches.
func (f *TagFilter) Match(entry LogEntry) bool {
	if f.Tag == "" {
		return true
	}

	matched := strings.EqualFold(entry.Tag, f.Tag)

	if f.Invert {
		return !matched
	}
	return matched
}

// PackageFilter filters log entries by package name.
type PackageFilter struct {
	// Package is the package to match (case-insensitive).
	Package string
	// Invert inverts the match.
	Invert bool
}

// Match returns true if the log entry's package matches.
func (f *PackageFilter) Match(entry LogEntry) bool {
	if f.Package == "" {
		return true
	}

	matched := strings.EqualFold(entry.Package, f.Package)

	if f.Invert {
		return !matched
	}
	return matched
}

// CompoundFilter combines multiple filters with AND logic.
type CompoundFilter struct {
	// Filters is the list of filters to apply.
	Filters []Filter
}

// Match returns true if all filters match.
func (f *CompoundFilter) Match(entry LogEntry) bool {
	for _, filter := range f.Filters {
		if !filter.Match(entry) {
			return false
		}
	}
	return true
}

// Add adds a filter to the compound filter.
func (f *CompoundFilter) Add(filter Filter) {
	f.Filters = append(f.Filters, filter)
}

// AnyFilter combines multiple filters with OR logic.
type AnyFilter struct {
	// Filters is the list of filters to apply.
	Filters []Filter
}

// Match returns true if any filter matches.
func (f *AnyFilter) Match(entry LogEntry) bool {
	if len(f.Filters) == 0 {
		return true
	}
	for _, filter := range f.Filters {
		if filter.Match(entry) {
			return true
		}
	}
	return false
}

// Add adds a filter to the any filter.
func (f *AnyFilter) Add(filter Filter) {
	f.Filters = append(f.Filters, filter)
}

// BuildFilter creates a compound filter from stream options.
func BuildFilter(opts StreamOptions) Filter {
	compound := &CompoundFilter{Filters: []Filter{}}

	// Add level filter
	if opts.Level > LevelVerbose {
		compound.Add(&LevelFilter{MinLevel: opts.Level})
	}

	// Add text filter
	if opts.Filter != "" {
		compound.Add(&TextFilter{Pattern: opts.Filter})
	}

	// Add tag filter
	if opts.Tag != "" {
		compound.Add(&TagFilter{Tag: opts.Tag})
	}

	// Add package filter
	if opts.Package != "" {
		compound.Add(&PackageFilter{Package: opts.Package})
	}

	return compound
}
