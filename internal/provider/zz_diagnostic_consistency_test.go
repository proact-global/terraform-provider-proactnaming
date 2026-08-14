// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

// TEMPORARY DIAGNOSTIC -- remove once the acceptance failures are understood.
//
// Three acceptance tests fail in contradictory directions: two report a name as
// missing moments after it was created, while the drift test reports a name as
// present moments after it was deleted. Both cannot be explained by the same
// simple cause, and the logs do not distinguish between them.
//
// This probes the naming tool directly through the client -- the same calls the
// provider's Read makes -- and records what it observes over time, so the next
// CI run answers the question instead of inviting more speculation.

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/proact-global/azurenamingtool-client-go"
)

// diagnosticProbeDelays are the points after a write at which the read is
// retried. If a write becomes visible only after some delay, the transition
// shows up here.
var diagnosticProbeDelays = []time.Duration{
	0,
	250 * time.Millisecond,
	1 * time.Second,
	3 * time.Second,
	8 * time.Second,
}

func TestAccDiagnoseNamingToolConsistency(t *testing.T) {
	// resource.Test applies the TF_ACC guard for the other acceptance tests.
	// This one talks to the API directly, so it has to check for itself --
	// without this the unit job would try to reach the network and, having no
	// credentials under CI, fail.
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for the consistency diagnostic")
	}
	testAccPreCheckWithAdmin(t)

	host := os.Getenv("PROACTNAMING_HOST")
	apiKey := os.Getenv("PROACTNAMING_APIKEY")
	adminPwd := os.Getenv("PROACTNAMING_ADMIN_PASSWORD")

	client, err := azurenamingtool.NewClient(&host, &apiKey, &adminPwd)
	if err != nil {
		t.Fatalf("could not construct client: %v", err)
	}

	req := azurenamingtool.GenerateNameRequest{
		ResourceOrg:         testAccOrg(),
		ResourceType:        testAccResourceType(),
		ResourceEnvironment: testAccEnvironment(),
		ResourceLocation:    testAccLocation(),
		ResourceInstance:    testAccInstance(t, 500),
		CustomComponents:    map[string]string{"application": "probe"},
	}

	t.Log("=========== READ AFTER CREATE ===========")

	created, err := client.GenerateName(req)
	if err != nil {
		t.Fatalf("GenerateName failed: %v", err)
	}
	id := created.ResourceNameDetails.ID
	t.Logf("created id=%d name=%q", id, created.ResourceName)

	probeVisibility(t, client, id, "after create", true)
	logGeneratedNamesLog(t, host, apiKey, id, "after create")

	t.Log("=========== READ AFTER DELETE ===========")

	if _, err := client.DeleteName(azurenamingtool.DeleteGeneratedNameRequest{ID: id}); err != nil {
		t.Fatalf("DeleteName failed: %v", err)
	}
	t.Logf("deleted id=%d", id)

	probeVisibility(t, client, id, "after delete", false)
	logGeneratedNamesLog(t, host, apiKey, id, "after delete")

	t.Log("=========== ID REUSE ===========")

	// The plan-time preview creates an entry and deletes it again, so whether
	// the next create reuses that id decides if a stored id can later refer to
	// a different record.
	second, err := client.GenerateName(req)
	if err != nil {
		t.Fatalf("second GenerateName failed: %v", err)
	}
	t.Logf("second create id=%d name=%q (first was id=%d)", second.ResourceNameDetails.ID, second.ResourceName, id)
	if second.ResourceNameDetails.ID == id {
		t.Logf(">>> id was REUSED after the first was deleted")
	} else {
		t.Logf(">>> id was not reused (delta %d)", second.ResourceNameDetails.ID-id)
	}

	// Clean up so the probe does not leave records behind.
	if _, err := client.DeleteName(azurenamingtool.DeleteGeneratedNameRequest{ID: second.ResourceNameDetails.ID}); err != nil {
		t.Logf("cleanup of id=%d failed: %v", second.ResourceNameDetails.ID, err)
	}
}

// probeVisibility reports whether GetName can see id at each delay, and flags
// the point at which the answer stops matching what the write implies.
func probeVisibility(t *testing.T, client *azurenamingtool.Client, id int64, phase string, wantFound bool) {
	t.Helper()

	start := time.Now()
	previous := ""

	for _, delay := range diagnosticProbeDelays {
		if delay > 0 {
			time.Sleep(delay - time.Since(start))
		}

		outcome := "found"
		details, err := client.GetName(id)
		switch {
		case err == nil && details != nil:
			outcome = fmt.Sprintf("found (name=%q)", details.ResourceName)
		case err != nil:
			outcome = fmt.Sprintf("NOT FOUND (%v)", err)
		}

		marker := ""
		found := err == nil
		if found != wantFound {
			marker = "   <<< disagrees with the write"
		}
		if previous != "" && previous != outcome {
			marker += "   <<< CHANGED from the previous probe"
		}
		previous = outcome

		t.Logf("  %-6s %s: id=%d -> %s%s", time.Since(start).Round(time.Millisecond), phase, id, outcome, marker)
	}
}

// logGeneratedNamesLog checks the same question against the Admin list endpoint
// rather than the by-id lookup. If the two disagree, the problem is in how a
// single record is read, not in whether the write landed.
func logGeneratedNamesLog(t *testing.T, host, apiKey string, id int64, phase string) {
	t.Helper()

	req, err := http.NewRequest("GET", host+"/api/Admin/GetGeneratedNamesLog", nil)
	if err != nil {
		t.Logf("  list %s: could not build request: %v", phase, err)
		return
	}
	req.Header.Set("APIKey", apiKey)

	res, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		t.Logf("  list %s: request failed: %v", phase, err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Logf("  list %s: could not read body: %v", phase, err)
		return
	}

	// Only the presence of the id matters, so avoid unmarshalling a payload
	// whose shape may differ between versions. Both spacings are checked because
	// the serialiser's formatting is not guaranteed.
	haystack := strings.ToLower(string(body))
	present := strings.Contains(haystack, fmt.Sprintf(`"id":%d,`, id)) ||
		strings.Contains(haystack, fmt.Sprintf(`"id": %d,`, id))

	t.Logf("  list %s: status=%d bytes=%d contains id=%d -> %t",
		phase, res.StatusCode, len(body), id, present)
}
