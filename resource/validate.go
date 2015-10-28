package resource

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	Value = iota
	Values
	Contains
)

type TestResult struct {
	Successful   bool
	Title        string
	ResourceType string
	TestType     int
	Property     string
	Err          error
	Expected     []string
	Found        []string
	Duration     time.Duration
}

func ValidateValues(res IDer, property string, expectedValues []string, method func() ([]string, error)) TestResult {
	title := res.ID()
	typ := reflect.TypeOf(res)
	typs := strings.Split(typ.String(), ".")[1]
	startTime := time.Now()
	foundValues, err := method()
	if err != nil {
		return TestResult{
			Successful:   false,
			ResourceType: typs,
			TestType:     Values,
			Title:        title,
			Property:     property,
			Err:          err,
			Duration:     time.Now().Sub(startTime),
		}
	}
	set := make(map[string]bool)
	for _, v := range foundValues {
		set[v] = true
	}

	var bad []string
	for _, v := range expectedValues {
		if _, found := set[v]; !found {
			bad = append(bad, v)
		}
	}

	if len(bad) > 0 {
		return TestResult{
			Successful:   false,
			ResourceType: typs,
			TestType:     Values,
			Title:        title,
			Property:     property,
			Expected:     expectedValues,
			Found:        foundValues,
			Duration:     time.Now().Sub(startTime),
		}
	}
	return TestResult{
		Successful:   true,
		ResourceType: typs,
		TestType:     Values,
		Title:        title,
		Property:     property,
		Expected:     expectedValues,
		Found:        foundValues,
		Duration:     time.Now().Sub(startTime),
	}

}

func ValidateValue(res IDer, property string, expectedValue interface{}, method func() (interface{}, error)) TestResult {
	title := res.ID()
	typ := reflect.TypeOf(res)
	typs := strings.Split(typ.String(), ".")[1]
	startTime := time.Now()
	foundValue, err := method()
	if err != nil {
		return TestResult{
			Successful:   false,
			ResourceType: typs,
			TestType:     Value,
			Title:        title,
			Property:     property,
			Err:          err,
			Duration:     time.Now().Sub(startTime),
		}
	}

	if expectedValue == foundValue {
		return TestResult{
			Successful:   true,
			ResourceType: typs,
			TestType:     Value,
			Title:        title,
			Property:     property,
			Expected:     []string{interfaceToString(expectedValue)},
			Found:        []string{interfaceToString(foundValue)},
			Duration:     time.Now().Sub(startTime),
		}
	}

	return TestResult{
		Successful:   false,
		ResourceType: typs,
		TestType:     Value,
		Title:        title,
		Property:     property,
		Expected:     []string{interfaceToString(expectedValue)},
		Found:        []string{interfaceToString(foundValue)},
		Duration:     time.Now().Sub(startTime),
	}
}

// FIXME: Not sure I like this, perhaps move to more statically typed validations..
func interfaceToString(i interface{}) string {
	switch t := i.(type) {
	case string:
		return fmt.Sprintf("%s", t)
	case bool:
		return fmt.Sprintf("%t", t)
	case int:
		return fmt.Sprintf("%d", t)
	default:
		return fmt.Sprintf("Unexpected Type")
	}
}

type patternMatcher interface {
	Match(string) bool
	Pattern() string
	Inverse() bool
}

type stringPattern struct {
	pattern      string
	cleanPattern string
	inverse      bool
}

func newStringPattern(str string) *stringPattern {
	var inverse bool
	if strings.HasPrefix(str, "!") {
		inverse = true
	}
	cleanPattern := strings.TrimLeft(str, "\\/!")
	return &stringPattern{
		pattern:      str,
		cleanPattern: cleanPattern,
		inverse:      inverse,
	}
}

func (s *stringPattern) Match(str string) bool {
	return strings.Contains(str, s.cleanPattern)
}

func (s *stringPattern) Pattern() string { return s.pattern }
func (s *stringPattern) Inverse() bool   { return s.inverse }

type regexPattern struct {
	pattern string
	re      *regexp.Regexp
	inverse bool
}

func newRegexPattern(str string) *regexPattern {
	var inverse bool
	if strings.HasPrefix(str, "!") {
		inverse = true
	}
	// fixme, don't use MustCompile
	cleanStr := strings.TrimLeft(str, "\\/!")
	cleanStr = strings.TrimRight(cleanStr, "/")
	re := regexp.MustCompile(cleanStr)
	return &regexPattern{
		pattern: str,
		re:      re,
		inverse: inverse,
	}

}

func (re *regexPattern) Match(str string) bool {
	return re.re.MatchString(str)
}

func (re *regexPattern) Pattern() string { return re.pattern }
func (re *regexPattern) Inverse() bool   { return re.inverse }

func sliceToPatterns(slice []string) []patternMatcher {
	var patterns []patternMatcher
	for _, s := range slice {
		if (strings.HasPrefix(s, "/") || strings.HasPrefix(s, "!/")) && strings.HasSuffix(s, "/") {
			patterns = append(patterns, newRegexPattern(s))
		} else {
			patterns = append(patterns, newStringPattern(s))
		}
	}
	return patterns
}

func patternsToSlice(patterns []patternMatcher) []string {
	var slice []string
	for _, p := range patterns {
		slice = append(slice, p.Pattern())
	}
	return slice
}

func ValidateContains(res IDer, property string, expectedValues []string, method func() (io.Reader, error)) TestResult {
	title := res.ID()
	typ := reflect.TypeOf(res)
	typs := strings.Split(typ.String(), ".")[1]
	startTime := time.Now()
	fh, err := method()
	if err != nil {
		return TestResult{
			Successful:   false,
			ResourceType: typs,
			TestType:     Contains,
			Title:        title,
			Property:     property,
			Err:          err,
			Duration:     time.Now().Sub(startTime),
		}
	}
	scanner := bufio.NewScanner(fh)
	*scanner.Buffer(foo, 10)
	notfound := sliceToPatterns(expectedValues)
	var found []patternMatcher
	for scanner.Scan() {
		line := scanner.Text()

		i := 0
		for _, pat := range notfound {
			if pat.Match(line) {
				// Found it, but wasn't supposed to, don't mark it as found, but remove it from search
				if !pat.Inverse() {
					found = append(found, pat)
				}
				continue
			}
			notfound[i] = pat
			i++
		}
		notfound = notfound[:i]
		if len(notfound) == 0 {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return TestResult{
			Successful:   false,
			ResourceType: typs,
			TestType:     Contains,
			Title:        title,
			Property:     property,
			Err:          err,
			Duration:     time.Now().Sub(startTime),
		}
	}

	for _, pat := range notfound {
		// Didn't find it, but we didn't want to.. so we mark it as found
		// Empty pattern should match even if input to scanner is empty
		if pat.Inverse() || pat.Pattern() == "" {
			found = append(found, pat)
		}
	}

	if len(expectedValues) != len(found) {
		return TestResult{
			Successful:   false,
			ResourceType: typs,
			TestType:     Contains,
			Title:        title,
			Property:     property,
			Expected:     expectedValues,
			Found:        patternsToSlice(found),
			Duration:     time.Now().Sub(startTime),
		}
	}
	return TestResult{
		Successful:   true,
		ResourceType: typs,
		TestType:     Contains,
		Title:        title,
		Property:     property,
		Expected:     expectedValues,
		Found:        patternsToSlice(found),
		Duration:     time.Now().Sub(startTime),
	}

}
