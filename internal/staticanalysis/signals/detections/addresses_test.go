package detections

import (
	"testing"
)

// source: https://mathiasbynens.be/demo/url-regex
var validURLs = []string{
	"https://foo.com",
	"http://foo.com/blah_blah",
	"http://foo.com/blah_blah/",
	"http://foo.com/blah_blah_(wikipedia)",
	"http://foo.com/blah_blah_(wikipedia)_(again)",
	"http://www.example.com/wpstyle/?p=364",
	"https://www.example.com/foo/?bar=baz&inga=42&quux",
	"http://✪df.ws/123",
	"http://142.42.1.1/",
	"http://142.42.1.1:8080/",
	"http://➡.ws/䨹",
	"http://⌘.ws",
	"http://⌘.ws/",
	"http://foo.com/blah_(wikipedia)#cite-1",
	"http://foo.com/blah_(wikipedia)_blah#cite-1",
	"http://foo.com/unicode_(✪)_in_parens",
	"http://foo.com/(something)?after=parens",
	"http://☺.damowmow.com/",
	"http://code.google.com/events/#&product=browser",
	"http://j.mp",
	"ftp://foo.bar/baz",
	"http://foo.bar/?q=Test%20URL-encoded%20stuff",
	"http://مثال.إختبار",
	"http://例子.测试",
	"http://1337.net",
	"http://a.b-c.de",
	"http://223.255.255.254",
	"http://en.example.org/abcd/Apples_on_the_Trees_(film)",
	"https://www.example.co.uk/maps/place/Hello+World/@33.1338131,27.0154284,17z/data=!3m1!4b1!4m2!3m1!1s0x1bae1351dbe4e30b:0x23e2fa2705b4c315",
	"http://www.qwertyuiop.com/?Mamas&papa&resources-for-papas&Id=123",
	"http://[2001:db8::1]:80",
	"http://[2001:db8::1]:80/anz?val=42",
	"ws://[2600:3c00::f03c:91ff:fe73:2b08]:31333",
	"http://[1080:0:0:0:8:800:200C:417A]/index.html",
	"HTTPS://GITHUB.COM/",
	"http://GITHUB.COM/",
	"https://GITHUB.com:443",
	// note, these IP addresses were specified as invalid in the source test cases
	"http://0.0.0.0",
	"http://10.1.1.0",
	"http://10.1.1.255",
	"http://224.1.1.1",
	"http://10.1.1.1",
	// TODO data URLs
	// "data:,Hello%2C%20World%21",
	// "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==",
	// "data:text/html,%3Ch1%3EHello%2C%20World%21%3C%2Fh1%3E",
	// "data:text/html,%3Cscript%3Ealert%28%27hi%27%29%3B%3C%2Fscript%3E",
	// Unsupported but still technically valid URLs
	//"http://उदाहरण.परीक्षा",
	//"http://userid:password@example.com:8080",
	//"http://userid:password@example.com:8080/",
	//"http://userid@example.com",
	//"http://userid@example.com/",
	//"http://userid@example.com:8080",
	//"http://userid@example.com:8080/",
	//"http://userid:password@example.com",
	//"http://userid:password@example.com/",
	//"http://-.~_!$&'()*+,;=:%40:80%2f::::::@example.com",
}

var invalidURLs = []string{
	"foo.com", // this one is debatable
	"http://",
	"http://.",
	"http://..",
	"http://../",
	"http://?",
	"http://??",
	"http://??/",
	"http://#",
	"http://##",
	"http://##/",
	"http://foo.bar?q=Spaces should be encoded",
	"//",
	"//a",
	"///a",
	"///",
	"http:///a",
	"rdar://1234",
	"h://test",
	"http:// shouldfail.com",
	":// should fail",
	"http://foo.bar/foo(bar)baz quux",
	"ftps://foo.bar/",
	"http://3628126748",
	"http://.www.foo.bar/",
	"http://www.foo.bar./",
	"http://.www.foo.bar./",
	"http://1.1.1.1.1",
	"http://123.123.123",
	// TODO unsupported
	//"http://a.b--c.de/",
	//"http://-a.b.co",
	//"http://a.b-.co",
	//"http://-error-.invalid/",
}

// https://stackoverflow.com/a/33618635
// https://stackoverflow.com/a/2596900/
var validIPv4Addresses = []string{
	"127.0.0.1",
	"192.168.1.1",
	"192.168.1.255",
	"255.255.255.255",
	"0.0.0.0",
	"1.0.1.0",
	"8.8.8.8",
	"100.1.2.3",
	"172.15.1.2",
	"172.32.1.2",
	"192.167.1.2",
	// below are "private" addresses or not technically valid
	// depending on the context, but syntactically correct
	"0.1.2.3",
	"10.1.2.3",
	"172.16.1.2",
	"172.31.1.2",
	"192.168.1.2",
	"255.255.255.255",
}

var invalidIPv4Addresses = []string{
	"1.1.1.1.1",
	"1.1.1.00",
	"30.168.1.255.1",
	"127.1",
	"192.168.1.256",
	"-1.2.3.4",
	"1.1.1.1.",
	"3...3",
	"1.1.1.01",
	".2.3.4",
	"1.2.3.",
	"1.2.3.256",
	"1.2.256.4",
	"1.256.3.4",
	"256.2.3.4",
	"1.2.3.4.5",
	"1..3.4",
}

// https://www.ibm.com/docs/en/ts4500-tape-library?topic=functionality-ipv4-ipv6-address-formats
var validIPv6Addresses = []string{
	// IPv6 normal/single addresses
	"2001:db8:3333:4444:5555:6666:7777:8888",
	"2001:db8:3333:4444:CCCC:DDDD:EEEE:FFFF",
	"::",
	"2001:db8::",
	"::1234:5678",
	"2001:db8::1234:5678",
	"2001:0db8:0001:0000:0000:0ab9:C0A8:0102",
	"001:db8:3333:4444:5555:6666:7777:8888",
	"2001:db8:3333:4444:CCCC:DDDD:EEEE:FFFF",
	"2001:db8::",
	"::1234:5678",
	"2001:db8::1234:5678",
	"2001:0db8:0001:0000:0000:0ab9:C0A8:0102",
	// IPv6 dual addresses
	"2001:db8:3333:4444:5555:6666:1.2.3.4",
	"::11.22.33.44",
	"2001:db8::123.123.123.123",
	"::1234:5678:91.123.4.56",
	"::1234:5678:1.2.3.4",
	"2001:db8::1234:5678:5.6.7.8",
	"2001:db8:3333:4444:5555:6666:1.2.3.4",
	"::11.22.33.44",
	"::1234:5678:91.123.4.56",
	"::1234:5678:1.2.3.4",
	"2001:db8::123.123.123.123",
	"2001:db8::1234:5678:5.6.7.8",
	"face::11.22.33.44",
}

var invalidIPv6Addresses = []string{
	"2001db8:3333:44445555:66667777:8888",
	"2001:db8:3333:4444:CCCC:DDDD:EEEE",
	"::::",
	"2001:db8::::",
	"::12345678",
	"2001:5678",
	"1:2:3:4:5.6.7.8",
	"::1:2:3:45:6:7:8",
	"12::1:2:3:45:6:7:8",
	"1:1:1:1:1::1.1.1.1",
	"::1:1:1:1:1:1.1.1.1",
	// The below IPv6 addresses are invalid, however the regexp incorrectly treats them as valid
	//"1:2:3:4::5:6:7:8",
	//"1:1:1:1::1:1:1",
	//"1:1:1::1:1:1.1.1.1",
}

var validEmailAddresses = []string{
	"simple@example.com",
	"very.common@example.com",
	"disposable.style.email.with+symbol@example.com",
	"other.email-with-hyphen@and.subdomains.example.com",
	"fully-qualified-domain@example.com",
	"user.name+tag+sorting@example.com",
	"x@example.com",
	"example-indeed@strange-example.com",
	"test/test@test.com",
	"admin@mailserver1",
	"example@s.example",
	"mailhost!username@example.org",
	"user%example.com@example.org",
	"user-@example.org",
	"postmaster@[123.123.123.123]",
	"postmaster@[IPv6:2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
	`"john..doe"@example.org`,
	"UPPERCASE_NAME@lowercasedomain.com",
	"lower_case_name@UPPERCASEDOMAIN.COM",
	// Unsupported email addresses
	// `" "@example.org`,
	// `"very.(),:;<>[]\".VERY.\"very@\\ \"very\".unusual"@strange.example.com`,
}

var invalidEmailAddresses = []string{
	"Abc.example.com",
	"A@b@c@example.com",
	`this is"not\allowed@example.com`,
	`this\ still\"not\\allowed@example.com`,
	`1234567890123456789012345678901234567890123456789012345678901234+x@example.com`,
	// These are incorrectly treated as valid email addresses
	// `a"b(c)d,e:f;g<h>i[j\k]l@example.com`,
	// `just"not"right@example.com`,
	// `i.like.underscores@but_its_not_allowed_in_this_part`,
	// `QA[icon]CHOCOLATE[icon]@test.com`,
}

// TestURLRegexp tests exact matching on single URLs
func TestURLRegexp(t *testing.T) {
	for _, url := range validURLs {
		result := FindURLs(url)
		if !(len(result) == 1 && url == result[0]) {
			t.Errorf("expected to detect valid URL %s, got %v", url, result)
		}
	}
	for _, url := range invalidURLs {
		result := FindURLs(url)
		if len(result) == 1 && url == result[0] {
			t.Errorf("expected not to detect invalid URL %s, got %v", url, result)
		}
	}
}

// TestIPv4Regexp tests exact matching on single IPv4 addresses
func TestIPv4Regexp(t *testing.T) {
	for _, addr := range validIPv4Addresses {
		result := findIPv4Addresses(addr)
		if !(len(result) == 1 && addr == result[0]) {
			t.Errorf("expected to detect valid IPv4 address %s, got %v", addr, result)
		}
	}
	for _, addr := range invalidIPv4Addresses {
		result := findIPv4Addresses(addr)
		if len(result) == 1 && addr == result[0] {
			t.Errorf("expected not to detect invalid IPv4 address %s, got, %v", addr, result)
		}
	}
}

// TestIPv6Regexp tests exact matching on single IPv6 addresses
// Note that the regexp currently gives some false positives;
// see the commented out negative test cases for specific examples.
func TestIPv6Regexp(t *testing.T) {
	for _, addr := range validIPv6Addresses {
		result := findIPv6Addresses(addr)
		if !(len(result) == 1 && addr == result[0]) {
			t.Errorf("expected to detect valid IPv6 address %s, got %v", addr, result)
		}
	}
	for _, addr := range invalidIPv6Addresses {
		result := findIPv6Addresses(addr)
		if len(result) == 1 && addr == result[0] {
			t.Errorf("expected not to detect invalid IPv6 address %s, got %v", addr, result)
		}
	}
}
