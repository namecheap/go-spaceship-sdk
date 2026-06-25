package client

import (
	"context"
	"os"
	"testing"
)

// Shared helpers for live acceptance tests (TestAcc* prefix). These hit the
// real Spaceship API and are skipped unless credentials are set. They mirror
// the env-var conventions of domains_acc_test.go and the
// terraform-provider-spaceship acceptance suite, so the same .env works for
// both: SPACESHIP_API_KEY, SPACESHIP_API_SECRET, optional SPACESHIP_BASE_URL,
// SPACESHIP_TEST_DOMAIN, SPACESHIP_TEST_RECORD_PREFIX.
//
// DNS acceptance tests CREATE AND DELETE records on SPACESHIP_TEST_DOMAIN.
// Point it at a throwaway domain you control.

const defaultAccBaseURL = "https://spaceship.dev/api/v1"

// testAccBaseURL returns the API base URL, allowing override for staging.
func testAccBaseURL() string {
	if v := os.Getenv("SPACESHIP_BASE_URL"); v != "" {
		return v
	}
	return defaultAccBaseURL
}

// testAccClient builds a live client, skipping the test if credentials are absent.
func testAccClient(t *testing.T) *Client {
	t.Helper()
	key := os.Getenv("SPACESHIP_API_KEY")
	secret := os.Getenv("SPACESHIP_API_SECRET")
	if key == "" || secret == "" {
		t.Skip("set SPACESHIP_API_KEY and SPACESHIP_API_SECRET to run live acceptance tests")
	}
	c, err := NewClient(testAccBaseURL(), key, secret)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return c
}

// testAccDomain returns the domain to mutate in DNS acceptance tests, skipping
// when unset. Use a throwaway domain — these tests create and delete records.
func testAccDomain(t *testing.T) string {
	t.Helper()
	domain := os.Getenv("SPACESHIP_TEST_DOMAIN")
	if domain == "" {
		t.Skip("set SPACESHIP_TEST_DOMAIN (a throwaway domain) to run DNS acceptance tests")
	}
	return domain
}

// testAccRecordName builds a record name from SPACESHIP_TEST_RECORD_PREFIX
// (default "goacc") plus the given parts, so repeat runs against the same
// domain are namespaced and easy to spot.
func testAccRecordName(parts ...string) string {
	name := os.Getenv("SPACESHIP_TEST_RECORD_PREFIX")
	if name == "" {
		name = "goacc"
	}
	for _, p := range parts {
		name += "-" + p
	}
	return name
}

// testAccCleanupRecords registers a t.Cleanup that best-effort deletes the
// given records, so a failed assertion never leaves orphans on the live domain.
// DeleteDNSRecords already swallows 404, so deleting never-created records is safe.
func testAccCleanupRecords(t *testing.T, c *Client, domain string, records ...DNSRecord) {
	t.Helper()
	t.Cleanup(func() {
		if err := c.DeleteDNSRecords(context.Background(), domain, records); err != nil {
			t.Logf("cleanup: failed to delete %d record(s) on %s: %v", len(records), domain, err)
		}
	})
}
