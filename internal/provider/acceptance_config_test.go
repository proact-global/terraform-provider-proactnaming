// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Acceptance tests run against a live Azure Naming Tool, and the component
// values it accepts differ from instance to instance: short names, organisation
// codes, locations and environments are all part of each deployment's own
// configuration. Hard-coding them makes the suite pass on one instance and fail
// on every other, with errors ("ResourceType value is invalid.") that read like
// provider bugs.
//
// So the values are read from the environment, with the defaults matching the
// instance these tests were originally written against. Override whichever ones
// your naming tool disagrees with.
const (
	envTestOrg             = "PROACTNAMING_TEST_ORG"
	envTestLocation        = "PROACTNAMING_TEST_LOCATION"
	envTestEnvironment     = "PROACTNAMING_TEST_ENVIRONMENT"
	envTestResourceType    = "PROACTNAMING_TEST_RESOURCE_TYPE"
	envTestResourceTypeAlt = "PROACTNAMING_TEST_RESOURCE_TYPE_ALT"
	envTestInstanceWidth   = "PROACTNAMING_TEST_INSTANCE_WIDTH"
)

// testAccOrg returns the organisation component to generate names with.
func testAccOrg() string { return envOrDefault(envTestOrg, "man") }

// testAccLocation returns the location component to generate names with.
func testAccLocation() string { return envOrDefault(envTestLocation, "euw") }

// testAccEnvironment returns the environment component to generate names with.
func testAccEnvironment() string { return envOrDefault(envTestEnvironment, "dev") }

// testAccResourceType returns a resource type short name that exists on the
// target instance. The naming tool matches short names exactly and is case
// sensitive, so this must be the instance's own spelling.
func testAccResourceType() string { return envOrDefault(envTestResourceType, "st") }

// testAccResourceTypeAlt returns a second, different resource type short name,
// used by tests that generate more than one name at once.
func testAccResourceTypeAlt() string { return envOrDefault(envTestResourceTypeAlt, "rg") }

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// testAccInstanceWidth returns the exact number of characters the naming tool
// requires for the instance component. The tool validates this strictly (its
// message is of the form "must be between 3 and 3 characters"), so a value that
// is merely numeric is not enough -- it has to be the right length.
func testAccInstanceWidth() int {
	if v := strings.TrimSpace(os.Getenv(envTestInstanceWidth)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 3
}

// testAccInstance returns an instance value derived from the clock, so repeated
// runs do not collide, zero-padded to the width the naming tool requires.
//
// Padding is the point. Deriving the value with a plain modulo yields 1 to 3
// characters, which the tool rejects whenever the number happens to be small --
// intermittently for a modulo of 1000, and always for a modulo of 100.
//
// Pass distinct offsets when a single test needs several instance values.
func testAccInstance(offset int64) string {
	width := testAccInstanceWidth()

	mod := int64(1)
	for i := 0; i < width; i++ {
		mod *= 10
	}

	return fmt.Sprintf("%0*d", width, (time.Now().Unix()+offset)%mod)
}

// testAccNameRegexp builds a case-insensitive pattern asserting that the given
// components appear in order in a generated name, whatever delimiter the
// instance is configured to use. Components are quoted, so values containing
// regex metacharacters are matched literally.
func testAccNameRegexp(components ...string) *regexp.Regexp {
	quoted := make([]string, 0, len(components))
	for _, c := range components {
		quoted = append(quoted, regexp.QuoteMeta(c))
	}
	return regexp.MustCompile(`(?i)` + strings.Join(quoted, ".*"))
}

// TestTestAccInstanceIsAlwaysPaddedToWidth guards the defect these helpers were
// introduced to fix: instance values derived with a plain modulo are 1 to 3
// characters long, and the naming tool rejects anything but the exact width.
// That failed intermittently for some tests and unconditionally for others, and
// it was only ever visible against a live instance. Checking it here means the
// unit job catches a regression instead.
func TestTestAccInstanceIsAlwaysPaddedToWidth(t *testing.T) {
	for _, width := range []int{2, 3, 4} {
		t.Run(fmt.Sprintf("width=%d", width), func(t *testing.T) {
			t.Setenv(envTestInstanceWidth, strconv.Itoa(width))

			// Sweep offsets so small values -- the ones that used to slip
			// through unpadded -- are covered too.
			for offset := int64(0); offset < 250; offset++ {
				got := testAccInstance(offset)
				if len(got) != width {
					t.Fatalf("testAccInstance(%d) = %q, want %d characters", offset, got, width)
				}
				if _, err := strconv.Atoi(got); err != nil {
					t.Fatalf("testAccInstance(%d) = %q, want digits only", offset, got)
				}
			}
		})
	}
}

func TestTestAccInstanceDefaultsToThreeCharacters(t *testing.T) {
	t.Setenv(envTestInstanceWidth, "")
	if got := testAccInstance(0); len(got) != 3 {
		t.Errorf("testAccInstance(0) = %q, want 3 characters by default", got)
	}
}

func TestTestAccInstanceDistinctOffsetsDiffer(t *testing.T) {
	t.Setenv(envTestInstanceWidth, "3")
	if a, b := testAccInstance(0), testAccInstance(1); a == b {
		t.Errorf("offsets 0 and 1 both produced %q, want distinct values", a)
	}
}

func TestEnvOrDefault(t *testing.T) {
	const key = "PROACTNAMING_TEST_SOME_VALUE"

	t.Run("falls back when unset", func(t *testing.T) {
		t.Setenv(key, "")
		if got := envOrDefault(key, "fallback"); got != "fallback" {
			t.Errorf("got %q, want %q", got, "fallback")
		}
	})
	t.Run("falls back when only whitespace", func(t *testing.T) {
		t.Setenv(key, "   ")
		if got := envOrDefault(key, "fallback"); got != "fallback" {
			t.Errorf("got %q, want %q", got, "fallback")
		}
	})
	t.Run("prefers the environment and trims it", func(t *testing.T) {
		t.Setenv(key, "  override  ")
		if got := envOrDefault(key, "fallback"); got != "override" {
			t.Errorf("got %q, want %q", got, "override")
		}
	})
}

func TestTestAccNameRegexp(t *testing.T) {
	re := testAccNameRegexp("man", "st", "webapp")

	for _, name := range []string{"man-st-webapp-001", "MANSTWEBAPP", "x_man_st_webapp_y"} {
		if !re.MatchString(name) {
			t.Errorf("%q should match %v", name, re)
		}
	}
	// Order matters, and unrelated names must not match.
	for _, name := range []string{"st-man-webapp", "man-st", "unrelated"} {
		if re.MatchString(name) {
			t.Errorf("%q should not match %v", name, re)
		}
	}
	// Metacharacters in a component are matched literally, not as a pattern.
	if testAccNameRegexp("a.c").MatchString("abc") {
		t.Error(`component "a.c" should not match "abc"`)
	}
}
