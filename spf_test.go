package spf

import (
	"fmt"
	"net"
	"testing"
)

var ip1110 = net.ParseIP("1.1.1.0")
var ip1111 = net.ParseIP("1.1.1.1")
var ip6666 = net.ParseIP("2001:db8::68")
var ip6660 = net.ParseIP("2001:db8::0")

func TestBasic(t *testing.T) {
	dns = NewDNS()

	cases := []struct {
		txt string
		res Result
		err error
	}{
		{"", None, nil},
		{"blah", None, nil},
		{"v=spf1", Neutral, nil},
		{"v=spf1 ", Neutral, nil},
		{"v=spf1 -", PermError, errUnknownField},
		{"v=spf1 all", Pass, errMatchedAll},
		{"v=spf1  +all", Pass, errMatchedAll},
		{"v=spf1 -all ", Fail, errMatchedAll},
		{"v=spf1 ~all", SoftFail, errMatchedAll},
		{"v=spf1 ?all", Neutral, errMatchedAll},
		{"v=spf1 a ~all", SoftFail, errMatchedAll},
		{"v=spf1 a/24", Neutral, nil},
		{"v=spf1 a:d1110/24", Pass, errMatchedA},
		{"v=spf1 a:d1110/montoto", PermError, errInvalidMask},
		{"v=spf1 a:d1110/99", PermError, errInvalidMask},
		{"v=spf1 a:d1110/32", Neutral, nil},
		{"v=spf1 a:d1110", Neutral, nil},
		{"v=spf1 a:d1111", Pass, errMatchedA},
		{"v=spf1 a:nothing/24", Neutral, nil},
		{"v=spf1 mx", Neutral, nil},
		{"v=spf1 mx/24", Neutral, nil},
		{"v=spf1 mx:a/montoto ~all", PermError, errInvalidMask},
		{"v=spf1 mx:d1110/24 ~all", Pass, errMatchedMX},
		{"v=spf1 mx:d1110/99 ~all", PermError, errInvalidMask},
		{"v=spf1 ip4:1.2.3.4 ~all", SoftFail, errMatchedAll},
		{"v=spf1 ip6:12 ~all", PermError, errInvalidIP},
		{"v=spf1 ip4:1.1.1.1 -all", Pass, errMatchedIP},
		{"v=spf1 ip4:1.1.1.1/24 -all", Pass, errMatchedIP},
		{"v=spf1 ip4:1.1.1.1/lala -all", PermError, errInvalidMask},
		{"v=spf1 include:doesnotexist", PermError, errNoResult},
		{"v=spf1 ptr -all", Pass, errMatchedPTR},
		{"v=spf1 ptr:d1111 -all", Pass, errMatchedPTR},
		{"v=spf1 ptr:lalala -all", Pass, errMatchedPTR},
		{"v=spf1 ptr:doesnotexist -all", Fail, errMatchedAll},
		{"v=spf1 blah", PermError, errUnknownField},
	}

	dns.ip["d1111"] = []net.IP{ip1111}
	dns.ip["d1110"] = []net.IP{ip1110}
	dns.mx["d1110"] = []*net.MX{{"d1110", 5}, {"nothing", 10}}
	dns.addr["1.1.1.1"] = []string{"lalala.", "domain.", "d1111."}

	for _, c := range cases {
		dns.txt["domain"] = []string{c.txt}
		res, err := CheckHost(ip1111, "domain")
		if (res == TempError || res == PermError) && (err == nil) {
			t.Errorf("%q: expected error, got nil", c.txt)
		}
		if res != c.res {
			t.Errorf("%q: expected %q, got %q", c.txt, c.res, res)
		}
		if err != c.err {
			t.Errorf("%q: expected error [%v], got [%v]", c.txt, c.err, err)
		}
	}
}

func TestIPv6(t *testing.T) {
	dns = NewDNS()

	cases := []struct {
		txt string
		res Result
		err error
	}{
		{"v=spf1 all", Pass, errMatchedAll},
		{"v=spf1 a ~all", SoftFail, errMatchedAll},
		{"v=spf1 a/24", Neutral, nil},
		{"v=spf1 a:d6660/24", Pass, errMatchedA},
		{"v=spf1 a:d6660", Neutral, nil},
		{"v=spf1 a:d6666", Pass, errMatchedA},
		{"v=spf1 a:nothing/24", Neutral, nil},
		{"v=spf1 mx:d6660/24 ~all", Pass, errMatchedMX},
		{"v=spf1 ip6:2001:db8::68 ~all", Pass, errMatchedIP},
		{"v=spf1 ip6:2001:db8::1/24 ~all", Pass, errMatchedIP},
		{"v=spf1 ip6:2001:db8::1/100 ~all", Pass, errMatchedIP},
		{"v=spf1 ptr -all", Pass, errMatchedPTR},
		{"v=spf1 ptr:d6666 -all", Pass, errMatchedPTR},
		{"v=spf1 ptr:sonlas6 -all", Pass, errMatchedPTR},
	}

	dns.ip["d6666"] = []net.IP{ip6666}
	dns.ip["d6660"] = []net.IP{ip6660}
	dns.mx["d6660"] = []*net.MX{{"d6660", 5}, {"nothing", 10}}
	dns.addr["2001:db8::68"] = []string{"sonlas6.", "domain.", "d6666."}

	for _, c := range cases {
		dns.txt["domain"] = []string{c.txt}
		res, err := CheckHost(ip6666, "domain")
		if (res == TempError || res == PermError) && (err == nil) {
			t.Errorf("%q: expected error, got nil", c.txt)
		}
		if res != c.res {
			t.Errorf("%q: expected %q, got %q", c.txt, c.res, res)
		}
		if err != c.err {
			t.Errorf("%q: expected error [%v], got [%v]", c.txt, c.err, err)
		}
	}
}

func TestNotSupported(t *testing.T) {
	cases := []struct {
		txt string
		err error
	}{
		{"v=spf1 exists:blah -all", errExistsNotSupported},
		{"v=spf1 exp=blah -all", errExpNotSupported},
		{"v=spf1 a:%{o} -all", errMacrosNotSupported},
		{"v=spf1 redirect=_spf.%{d}", errMacrosNotSupported},
	}

	for _, c := range cases {
		dns.txt["domain"] = []string{c.txt}
		res, err := CheckHost(ip1111, "domain")
		if res != Neutral || err != c.err {
			t.Errorf("%q: expected neutral/%q, got %v/%q", c.txt, c.err, res, err)
		}
	}
}

func TestInclude(t *testing.T) {
	// Test that the include is doing a recursive lookup.
	// If we got a match on 1.1.1.1, is because include:domain2 did not match.
	dns = NewDNS()
	dns.txt["domain"] = []string{"v=spf1 include:domain2 ip4:1.1.1.1"}

	cases := []struct {
		txt string
		res Result
		err error
	}{
		{"", PermError, errNoResult},
		{"v=spf1 all", Pass, errMatchedAll},

		// domain2 did not pass, so continued and matched parent's ip4.
		{"v=spf1", Pass, errMatchedIP},
		{"v=spf1 -all", Pass, errMatchedIP},
	}

	for _, c := range cases {
		dns.txt["domain2"] = []string{c.txt}
		res, err := CheckHost(ip1111, "domain")
		if res != c.res || err != c.err {
			t.Errorf("%q: expected [%v/%v], got [%v/%v]",
				c.txt, c.res, c.err, res, err)
		}
	}
}

func TestRecursionLimit(t *testing.T) {
	dns = NewDNS()
	dns.txt["domain"] = []string{"v=spf1 include:domain ~all"}

	res, err := CheckHost(ip1111, "domain")
	if res != PermError || err != errLookupLimitReached {
		t.Errorf("expected permerror, got %v (%v)", res, err)
	}
}

func TestRedirect(t *testing.T) {
	dns = NewDNS()
	dns.txt["domain"] = []string{"v=spf1 redirect=domain2"}
	dns.txt["domain2"] = []string{"v=spf1 ip4:1.1.1.1 -all"}

	res, err := CheckHost(ip1111, "domain")
	if res != Pass {
		t.Errorf("expected pass, got %v (%v)", res, err)
	}
}

func TestInvalidRedirect(t *testing.T) {
	// Redirect to a non-existing host; the inner check returns None, but due
	// to the redirection, this lookup should return PermError.
	// https://tools.ietf.org/html/rfc7208#section-6.1
	dns = NewDNS()
	dns.txt["domain"] = []string{"v=spf1 redirect=doesnotexist"}

	res, err := CheckHost(ip1111, "doesnotexist")
	if res != None {
		t.Errorf("expected none, got %v (%v)", res, err)
	}

	res, err = CheckHost(ip1111, "domain")
	if res != PermError || err != nil {
		t.Errorf("expected permerror, got %v (%v)", res, err)
	}
}

func TestRedirectOrder(t *testing.T) {
	// We should only check redirects after all mechanisms, even if the
	// redirect modifier appears before them.
	dns = NewDNS()
	dns.txt["faildom"] = []string{"v=spf1 -all"}

	dns.txt["domain"] = []string{"v=spf1 redirect=faildom"}
	res, err := CheckHost(ip1111, "domain")
	if res != Fail || err != errMatchedAll {
		t.Errorf("expected fail, got %v (%v)", res, err)
	}

	dns.txt["domain"] = []string{"v=spf1 redirect=faildom all"}
	res, err = CheckHost(ip1111, "domain")
	if res != Pass || err != errMatchedAll {
		t.Errorf("expected pass, got %v (%v)", res, err)
	}
}

func TestNoRecord(t *testing.T) {
	dns = NewDNS()
	dns.txt["d1"] = []string{""}
	dns.txt["d2"] = []string{"loco", "v=spf2"}
	dns.errors["nospf"] = fmt.Errorf("no such domain")

	for _, domain := range []string{"d1", "d2", "d3", "nospf"} {
		res, err := CheckHost(ip1111, domain)
		if res != None {
			t.Errorf("expected none, got %v (%v)", res, err)
		}
	}
}

func TestDNSTemporaryErrors(t *testing.T) {
	dns = NewDNS()
	dnsError := &net.DNSError{
		Err:         "temporary error for testing",
		IsTemporary: true,
	}

	// Domain "tmperr" will fail resolution with a temporary error.
	dns.errors["tmperr"] = dnsError
	dns.errors["1.1.1.1"] = dnsError
	dns.mx["tmpmx"] = []*net.MX{{"tmperr", 10}}

	cases := []struct {
		txt string
		res Result
	}{
		{"v=spf1 include:tmperr", TempError},
		{"v=spf1 a:tmperr", TempError},
		{"v=spf1 mx:tmperr", TempError},
		{"v=spf1 ptr:tmperr", TempError},
		{"v=spf1 mx:tmpmx", TempError},
	}

	for _, c := range cases {
		dns.txt["domain"] = []string{c.txt}
		res, err := CheckHost(ip1111, "domain")
		if res != c.res {
			t.Errorf("%q: expected %v, got %v (%v)",
				c.txt, c.res, res, err)
		}
	}
}

func TestDNSPermanentErrors(t *testing.T) {
	dns = NewDNS()
	dnsError := &net.DNSError{
		Err:         "permanent error for testing",
		IsTemporary: false,
	}

	// Domain "tmperr" will fail resolution with a temporary error.
	dns.errors["tmperr"] = dnsError
	dns.errors["1.1.1.1"] = dnsError
	dns.mx["tmpmx"] = []*net.MX{{"tmperr", 10}}

	cases := []struct {
		txt string
		res Result
	}{
		{"v=spf1 include:tmperr", PermError},
		{"v=spf1 a:tmperr", Neutral},
		{"v=spf1 mx:tmperr", Neutral},
		{"v=spf1 ptr:tmperr", Neutral},
		{"v=spf1 mx:tmpmx", Neutral},
	}

	for _, c := range cases {
		dns.txt["domain"] = []string{c.txt}
		res, err := CheckHost(ip1111, "domain")
		if res != c.res {
			t.Errorf("%q: expected %v, got %v (%v)",
				c.txt, c.res, res, err)
		}
	}
}
