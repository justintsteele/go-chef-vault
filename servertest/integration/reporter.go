package integration

import (
	"fmt"
	"reflect"
	"slices"
)

type Assertion struct {
	Name     string
	Passed   bool
	Error    error
	Actual   any
	Expected any
}

type ScenarioResult struct {
	Assertions []Assertion
}

type Scenario struct {
	Name string
	Run  func(*IntegrationService) *ScenarioResult
}

type ExecutedScenario struct {
	Name   string
	Result *ScenarioResult
}

func (sr *ScenarioResult) passed() bool {
	for _, a := range sr.Assertions {
		if !a.Passed {
			return false
		}
	}
	return true
}

func (sr *ScenarioResult) assert(name string, ok bool, err error) {
	sr.Assertions = append(sr.Assertions, Assertion{
		Name:   name,
		Passed: ok,
		Error:  err,
	})
}

func (sr *ScenarioResult) assertEqual(name string, expected, actual any) {
	sr.assertValue(name, expected, actual, true)
}

func (sr *ScenarioResult) assertNotEqual(name string, expected, actual any) {
	sr.assertValue(name, expected, actual, false)
}

func (sr *ScenarioResult) assertValue(name string, expected, actual any, want bool) {
	normalizedExpected := normalizeForEquality(expected)
	normalizedActual := normalizeForEquality(actual)

	passed := reflect.DeepEqual(normalizedExpected, normalizedActual)

	if want {
		sr.assert(name, passed, fmt.Errorf("expected %s == %s", normalizedExpected, normalizedActual))
	} else {
		sr.assert(name, !passed, fmt.Errorf("expected %s != %s", normalizedActual, normalizedExpected))
	}
}

func (sr *ScenarioResult) assertContains(name string, container any, value string) {
	sr.assertContainsValue(name, container, value, true)
}

func (sr *ScenarioResult) assertNotContains(name string, container any, value string) {
	sr.assertContainsValue(name, container, value, false)
}

func (sr *ScenarioResult) assertContainsValue(name string, container any, value string, want bool) {
	found, err := containsString(container, value)
	if err != nil {
		sr.assertNoError(name, err)
		return
	}

	if want {
		sr.assert(name, found, fmt.Errorf("%q not found in %v", value, container))
	} else {
		sr.assert(name, !found, fmt.Errorf("%q unexpectedly found in %v", value, container))
	}
}

func containsString(container any, value string) (bool, error) {
	switch x := container.(type) {
	case []string:
		return slices.Contains(x, value), nil

	case []interface{}:
		for _, e := range x {
			s, ok := e.(string)
			if !ok {
				return false, fmt.Errorf("expected string element, got %T", e)
			}
			if s == value {
				return true, nil
			}
		}
		return false, nil

	default:
		return false, fmt.Errorf("expected slice, got %T", container)
	}
}

func normalizeForEquality(v any) any {
	switch x := v.(type) {
	case []string:
		cpy := append([]string(nil), x...) // copy
		slices.Sort(cpy)
		return cpy
	case []interface{}:
		cpy := make([]string, 0, len(x))
		for _, v := range x {
			s, ok := v.(string)
			if !ok {
				continue
			}
			cpy = append(cpy, s)
		}
		slices.Sort(cpy)
		return cpy
	default:
		return v
	}
}

func (sr *ScenarioResult) assertNoError(name string, err error) {
	sr.assert(name, err == nil, err)
}

func (sr *ScenarioResult) assertError(name string, err error) {
	sr.assert(name, err != nil, err)
}

type ScenarioReporter interface {
	Report(results []*ExecutedScenario)
}

type ConsoleReporter struct{}

func (r *ConsoleReporter) Report(results []*ExecutedScenario) {
	fmt.Println()
	fmt.Println("=== Integration Scenarios ===")

	failed := 0

	for _, es := range results {
		status := "PASS"
		if !es.Result.passed() {
			status = "FAIL"
		}

		fmt.Printf("\n[%s] %s\n", status, es.Name)

		for _, a := range es.Result.Assertions {
			if a.Passed {
				fmt.Printf("  ✔ %s\n", a.Name)
				continue
			}

			fmt.Printf("  ✘ %s\n", a.Name)

			if a.Expected != nil || a.Actual != nil {
				fmt.Printf("      expected: %v\n", a.Expected)
				fmt.Printf("      actual:   %v\n", a.Actual)
			}

			if a.Error != nil {
				fmt.Printf("      error: %v\n", a.Error)
			}
			failed++
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d scenarios, %d failed\n", len(results), failed)
	fmt.Println()
}

func AnyFailed(results []*ExecutedScenario) bool {
	for _, r := range results {
		if !r.Result.passed() {
			return true
		}
	}
	return false
}
