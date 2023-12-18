package detections

import (
	"fmt"
	"net"
	"regexp"

	"github.com/ossf/package-analysis/internal/utils"
)

// digits0_255 matches decimal numbers from 0-255
var digits0_255 = regexp.MustCompile(`(?:25[0-5]|(?:2[0-4]|1[0-9]|[1-9]|)[0-9])`)

// TODO sometimes octal (leading 0) or hexadecimal (leading 0x) may be used
// https://stackoverflow.com/a/36760050/
var ipv4Regexp = regexp.MustCompile(fmt.Sprintf(`%s(?:\.%s){3}`, digits0_255, digits0_255))

// hex1_4 matches between 1 and 4 hex digits
var hex1_4 = regexp.MustCompile(`[[:xdigit:]]{1,4}`)

// ipv6Regexp matches all valid IPv6 address strings, covering
// the various representations of IPv6 addresses.
// Note that it has some false positive matches: specifically,
// compressed IPv6 addresses that have "too many segments" will
// also be matched.
var ipv6Regexp = utils.CombineRegexp(
	// matches uncompressed IPv6 addresses of the form
	// 123:fe:4567:dc:89ab:a9:cdef:87 (normal ipv6) or fedc:1:ba98:23:7654:45:123.54.89.7 (dual ipv6/ipv4)
	regexp.MustCompile(fmt.Sprintf(`%s(?::%s){5}(?:(?::%s){2}|:%s)`, hex1_4, hex1_4, hex1_4, ipv4Regexp)),

	// matches compressed dual IPv6 addresses of the form
	// fedc:1:ba98:23::123.54.89.7 or ::123.54.89.7 (i.e. compressed just before the IPv4 part
	regexp.MustCompile(fmt.Sprintf(`(?:(?:%s:){1,4}|:):%s`, hex1_4, ipv4Regexp)),

	// matches compressed dual IPv6 addresses of the form
	// fedc:1::ba98:23:123.54.89.7 (i.e. compressed in the middle)
	// NOTE: it also incorrectly matches IPv6-like strings with 'too many segments', as
	// it is difficult to ensure that the number of segments on each side of the '::'
	// adds up to a valid number
	regexp.MustCompile(fmt.Sprintf(`(?:(?:%s:){1,4}|:)(?::%s){1,4}:%s`, hex1_4, hex1_4, ipv4Regexp)),

	// matches compressed normal IPv6 addresses of the form
	// fedc:1:ba98:23:: or ::89ab or :: or 123:fe::89ab:a9:cdef:87
	// NOTE: it also incorrectly matches IPv6-like strings with 'too many segments', as
	// it is difficult to ensure that the number of segments on each side of the '::'
	// adds up to a valid number
	regexp.MustCompile(fmt.Sprintf(`(?:(?:%s:){1,6}|:)(?:(?::%s){1,6}|:)`, hex1_4, hex1_4)),
)

var urlSchemes = regexp.MustCompile(`(?i:https?|blob|chrome|content|file|ftp|gs|s3|ws)`)
var hostChars = regexp.MustCompile(`[\p{L}\p{N}\p{S}_-]`) // any unicode letter, number or symbol

var port = regexp.MustCompile(`(?::\d+)`)

// includes host and optional port
var urlAuthority = regexp.MustCompile(fmt.Sprintf(`%s+(?:\.%s+)*\.\p{L}+%s?`, hostChars, hostChars, port))

// urlPathAndQuery represents the and query string (e.g. /search?q=hello) in a URL
var urlPathAndQuery = regexp.MustCompile(`(?:/(?:[^.]\S*)?)?(?:\?\S*)?`)

// urlRegex is a fairly permissive url regex. Parts: scheme, subdomains, TLD, port, url query
var urlRegexp = regexp.MustCompile(fmt.Sprintf(`%s:(?://%s)%s`, urlSchemes, urlAuthority, urlPathAndQuery))

var ipv4URLRegexp = regexp.MustCompile(fmt.Sprintf(`%s://(?:%s)%s?%s`, urlSchemes, ipv4Regexp, port, urlPathAndQuery))
var ipv6URLRegexp = regexp.MustCompile(fmt.Sprintf(`%s://\[(?:%s)]%s?%s`, urlSchemes, ipv6Regexp, port, urlPathAndQuery))

func FindURLs(s string) []string {
	var urls []string
	urls = append(urls, urlRegexp.FindAllString(s, -1)...)
	urls = append(urls, ipv4URLRegexp.FindAllString(s, -1)...)
	urls = append(urls, ipv6URLRegexp.FindAllString(s, -1)...)
	return urls
}

func findIPv4Addresses(s string) []string {
	return ipv4Regexp.FindAllString(s, -1)
}

func findIPv6Addresses(s string) []string {
	var addresses []string
	// need extra validation since IPv6 regexes are hard
	for _, candidate := range ipv6Regexp.FindAllString(s, -1) {
		if net.ParseIP(candidate) != nil {
			addresses = append(addresses, candidate)
		}
	}
	return addresses
}

func FindIPAddresses(s string) []string {
	return append(findIPv4Addresses(s), findIPv6Addresses(s)...)
}
