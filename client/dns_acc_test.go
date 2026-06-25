package client

import (
	"strings"
	"testing"
)

// Live DNS acceptance tests. They create and delete records on
// SPACESHIP_TEST_DOMAIN and are skipped unless credentials + a test domain are
// set (see acc_test.go). These verify the client behaviors the unit tests can
// only mock: real wire serialization, the API's case rules, upsert-as-update
// semantics, and the documented empirical API constraints.

// TestAccDNSRecord_RoundTrip is the core lifecycle proof: create a record,
// read it back identical, delete it, confirm it is gone. Exercises
// CreateDNSRecord's wire payload, GetDNSRecords deserialization, the matching
// signature, and DeleteDNSRecord — end to end against the real API.
func TestAccDNSRecord_RoundTrip(t *testing.T) {
	c := testAccClient(t)
	domain := testAccDomain(t)
	ctx := t.Context()

	rec := DNSRecord{Type: "A", Name: testAccRecordName("a", "roundtrip"), TTL: 3600, Address: "203.0.113.10"}
	testAccCleanupRecords(t, c, domain, rec)

	if err := c.CreateDNSRecord(ctx, domain, rec); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := c.FindDNSRecord(ctx, domain, rec.Type, rec.Name, RecordValueSignature(rec))
	if err != nil {
		t.Fatalf("find after create: %v", err)
	}
	if got.Address != rec.Address {
		t.Errorf("address: got %q want %q", got.Address, rec.Address)
	}
	if got.TTL != rec.TTL {
		t.Errorf("ttl: got %d want %d", got.TTL, rec.TTL)
	}
	// Records created via the external API must land in the custom group
	// (or be ungrouped). This is the boundary filterCustomDNSRecords enforces.
	if got.Group != nil && got.Group.Type != DNSGroupCustom {
		t.Errorf("expected custom group, got %q", got.Group.Type)
	}

	if err := c.DeleteDNSRecord(ctx, domain, rec); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := c.FindDNSRecord(ctx, domain, rec.Type, rec.Name, RecordValueSignature(rec)); err != ErrRecordNotFound {
		t.Errorf("after delete: expected ErrRecordNotFound, got %v", err)
	}
}

// TestAccDNSRecord_TXTCaseSensitiveCoexist verifies the client's marquee
// matching rule against the real API: TXT values are case-sensitive, so two
// TXT records on the same host differing only in case are DISTINCT records and
// must coexist. For any other record type these would collide as duplicates.
func TestAccDNSRecord_TXTCaseSensitiveCoexist(t *testing.T) {
	c := testAccClient(t)
	domain := testAccDomain(t)
	ctx := t.Context()

	host := testAccRecordName("txt", "case")
	lower := DNSRecord{Type: "TXT", Name: host, TTL: 3600, Value: "spaceship-verify=abc"}
	upper := DNSRecord{Type: "TXT", Name: host, TTL: 3600, Value: "spaceship-verify=ABC"}
	testAccCleanupRecords(t, c, domain, lower, upper)

	if err := c.UpsertDNSRecords(ctx, domain, false, []DNSRecord{lower, upper}); err != nil {
		t.Fatalf("upsert two case-distinct TXT records: %v", err)
	}

	records, err := c.GetDNSRecords(ctx, domain)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	count := 0
	for _, r := range records {
		if strings.EqualFold(r.Type, "TXT") && strings.EqualFold(r.Name, host) {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 case-distinct TXT records on %s, got %d (TXT case-sensitivity broken)", host, count)
	}

	// Exact-case lookup must resolve to the matching record, not its twin.
	got, err := c.FindDNSRecord(ctx, domain, "TXT", host, RecordValueSignature(upper))
	if err != nil {
		t.Fatalf("find upper-case TXT: %v", err)
	}
	if got.Value != upper.Value {
		t.Errorf("case-sensitive match returned wrong record: got %q want %q", got.Value, upper.Value)
	}
}

// TestAccDNSRecord_UpsertUpdatesTTLInPlace verifies the upsert-as-update
// semantics documented on CreateDNSRecord: a second upsert with the same
// type+name+data but a different TTL updates the existing record rather than
// creating a duplicate (matching is on identity, not TTL).
func TestAccDNSRecord_UpsertUpdatesTTLInPlace(t *testing.T) {
	c := testAccClient(t)
	domain := testAccDomain(t)
	ctx := t.Context()

	name := testAccRecordName("a", "ttl")
	rec := DNSRecord{Type: "A", Name: name, TTL: 3600, Address: "203.0.113.20"}
	testAccCleanupRecords(t, c, domain, rec)

	if err := c.UpsertDNSRecords(ctx, domain, false, []DNSRecord{rec}); err != nil {
		t.Fatalf("initial upsert: %v", err)
	}

	updated := rec
	updated.TTL = 600
	if err := c.UpsertDNSRecords(ctx, domain, true, []DNSRecord{updated}); err != nil {
		t.Fatalf("update upsert: %v", err)
	}

	records, err := c.GetDNSRecords(ctx, domain)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	matches, ttl := 0, 0
	for _, r := range records {
		if strings.EqualFold(r.Type, "A") && strings.EqualFold(r.Name, name) && r.Address == rec.Address {
			matches++
			ttl = r.TTL
		}
	}
	if matches != 1 {
		t.Fatalf("expected exactly 1 A record (update, not duplicate), got %d", matches)
	}
	if ttl != 600 {
		t.Errorf("ttl after update: got %d want 600", ttl)
	}
}

// TestAccDNSRecord_DeleteAbsentIsNoop verifies the rate-limit-friendly delete
// contract: deleting a record that does not exist returns nil, because
// DeleteDNSRecords swallows the API's not-found response (IsNotFoundError).
func TestAccDNSRecord_DeleteAbsentIsNoop(t *testing.T) {
	c := testAccClient(t)
	domain := testAccDomain(t)
	ctx := t.Context()

	rec := DNSRecord{Type: "A", Name: testAccRecordName("a", "absent"), TTL: 3600, Address: "203.0.113.30"}
	if err := c.DeleteDNSRecord(ctx, domain, rec); err != nil {
		t.Fatalf("deleting an absent record should be a no-op, got: %v", err)
	}
}

// TestAccDNSRecord_SVCBPortWithoutSchemeRejected guards an empirical API
// constraint encoded in the SVCB/HTTPS validators: the real API returns 422
// when a port is set without a scheme, even though the published spec marks
// scheme optional. The client does not pre-validate on upsert, so this sends
// the raw record and asserts the API rejects it. If the API ever starts
// accepting it, this test fails loudly — a signal the validator comment is stale.
func TestAccDNSRecord_SVCBPortWithoutSchemeRejected(t *testing.T) {
	c := testAccClient(t)
	domain := testAccDomain(t)
	ctx := t.Context()

	svcPriority := 1
	rec := DNSRecord{
		Type:        "SVCB",
		Name:        testAccRecordName("svcb", "portnoscheme"),
		TTL:         3600,
		SvcPriority: &svcPriority,
		TargetName:  "svc.example.com",
		// Port must be a valid SVCB port string ("_N" or "*") so the ONLY
		// constraint left to violate is the missing scheme. An integer port
		// here would be rejected on port format instead, making the test pass
		// for the wrong reason.
		Port: NewStringPortValue("_443"),
	}

	// No pre-registered cleanup: this is a negative test, so the record should
	// never be created. Cleaning up unconditionally would fire a DELETE that
	// the API also rejects on the missing scheme (delete payloads are
	// validated too), producing misleading "failed to delete" noise. We only
	// clean up in the unexpected case where the API accepts the record.
	err := c.CreateDNSRecord(ctx, domain, rec)
	if err == nil {
		testAccCleanupRecords(t, c, domain, rec)
		t.Fatal("expected API to reject SVCB with port and no scheme (422), but it was accepted")
	}
	// Guard against passing for the wrong reason: the rejection must be about
	// the scheme, not the port format. If this assertion fails on "port", the
	// port value regressed; if it accepts the record, the empirical constraint
	// is stale.
	if !strings.Contains(strings.ToLower(err.Error()), "scheme") {
		t.Fatalf("expected a missing-scheme rejection, got a different error: %v", err)
	}
	t.Logf("API rejected port-without-scheme as expected: %v", err)
}

// TestAccGetDomainInfo verifies the read path for a single domain returns the
// queried domain. GetDomainInfo falls back to the list endpoint on HTTP 429;
// this exercises the primary path and the response shape.
func TestAccGetDomainInfo(t *testing.T) {
	c := testAccClient(t)
	domain := testAccDomain(t)
	ctx := t.Context()

	info, err := c.GetDomainInfo(ctx, domain)
	if err != nil {
		t.Fatalf("get domain info: %v", err)
	}
	if !strings.EqualFold(info.Name, domain) {
		t.Errorf("domain name: got %q want %q", info.Name, domain)
	}
}
