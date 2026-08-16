// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/proact-global/azurenamingtool-client-go"
)

// Acceptance tests write to a real, shared Azure Naming Tool. Until now they
// left no account of what they did there: which records they created, which
// they replaced, and whether the ones they were responsible for were actually
// removed at the end.
//
// That matters here more than it might elsewhere. Records have already been
// observed surviving a run, because the tool answers a wrong admin password
// with a success status, and because two runs writing at once lose each other's
// changes. Both were found by accident. This makes them visible on purpose: the
// ledger records what each test touched, and the run ends by asking the tool
// which of those records still exist.

type ledgerAction string

const (
	ledgerCreated  ledgerAction = "created"
	ledgerReplaced ledgerAction = "replaced"
)

type ledgerEntry struct {
	test   string
	action ledgerAction
	id     int64
	name   string
}

type namingLedger struct {
	mu      sync.Mutex
	entries []ledgerEntry
}

var ledger = &namingLedger{}

func (l *namingLedger) record(test string, action ledgerAction, id int64, name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, ledgerEntry{test: test, action: action, id: id, name: name})
}

func (l *namingLedger) snapshot() []ledgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]ledgerEntry(nil), l.entries...)
}

// testAccRecord returns a check that notes what a resource holds at this point
// in a test. Used after each step, it captures both the original record and the
// one replacing it, so a replacement leaves both in the account.
func testAccRecord(t *testing.T, resourceName string, action ledgerAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		id, err := strconv.ParseInt(rs.Primary.Attributes["id"], 10, 64)
		if err != nil {
			return fmt.Errorf("could not read id of %s: %w", resourceName, err)
		}
		name := rs.Primary.Attributes["resource_name"]

		ledger.record(t.Name(), action, id, name)
		t.Logf("naming tool: %s id=%d name=%q (%s)", action, id, name, resourceName)
		return nil
	}
}

// TestMain reports the ledger once the tests have finished, and asks the naming
// tool which of the recorded records survive.
//
// Every record a test creates should be gone by the end: Terraform destroys what
// it created, and a replacement destroys the record it replaced. Anything left
// is either a cleanup that silently failed or the work of something else writing
// to the same tool. Both are worth knowing, and neither is visible otherwise.
func TestMain(m *testing.M) {
	code := m.Run()

	entries := ledger.snapshot()
	if len(entries) == 0 {
		os.Exit(code)
	}

	fmt.Fprintf(os.Stderr, "\n=== naming tool records touched by this run ===\n")

	byTest := map[string][]ledgerEntry{}
	for _, e := range entries {
		byTest[e.test] = append(byTest[e.test], e)
	}
	names := make([]string, 0, len(byTest))
	for n := range byTest {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, n := range names {
		fmt.Fprintf(os.Stderr, "  %s\n", n)
		for _, e := range byTest[n] {
			fmt.Fprintf(os.Stderr, "    %-9s id=%-6d %s\n", e.action, e.id, e.name)
		}
	}

	survivors := checkForSurvivingRecords(entries)
	switch {
	case survivors == nil:
		fmt.Fprintf(os.Stderr, "\n  (not checking for survivors: no credentials)\n")
	case len(survivors) == 0:
		fmt.Fprintf(os.Stderr, "\n  all %d records were removed\n", len(entries))
	default:
		fmt.Fprintf(os.Stderr, "\n  %d of %d records SURVIVED and remain in the naming tool:\n",
			len(survivors), len(entries))
		for _, e := range survivors {
			fmt.Fprintf(os.Stderr, "    id=%-6d %-28s (%s)\n", e.id, e.name, e.test)
		}
		fmt.Fprintf(os.Stderr,
			"\n  Records outliving the run mean a cleanup failed, most often an incorrect\n"+
				"  admin_password, or that something else wrote to the tool during the run.\n")
	}
	fmt.Fprintln(os.Stderr)

	os.Exit(code)
}

// checkForSurvivingRecords asks the tool which recorded records still exist.
// A nil result means the check could not run for want of credentials, which is
// different from finding nothing.
func checkForSurvivingRecords(entries []ledgerEntry) []ledgerEntry {
	host := os.Getenv("PROACTNAMING_HOST")
	apiKey := os.Getenv("PROACTNAMING_APIKEY")
	if host == "" || apiKey == "" {
		return nil
	}

	client, err := azurenamingtool.NewClient(&host, &apiKey, nil)
	if err != nil {
		return nil
	}

	// A record may legitimately appear twice, as created and then replaced.
	seen := map[int64]bool{}
	survivors := []ledgerEntry{}
	for _, e := range entries {
		if seen[e.id] {
			continue
		}
		seen[e.id] = true

		if _, err := client.GetName(e.id); err == nil {
			survivors = append(survivors, e)
		} else if !errors.Is(err, azurenamingtool.ErrNotFound) {
			// The tool could not answer. Reporting this as a survivor would be a
			// guess, so say nothing rather than invent a leak.
			fmt.Fprintf(os.Stderr, "    (could not check id=%d: %v)\n", e.id, err)
		}
	}
	return survivors
}
