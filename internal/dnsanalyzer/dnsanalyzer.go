package dnsanalyzer

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type rawLookup struct {
	query *layers.DNS
	reply *layers.DNS
}

type DNSAnalyzer struct {
	lookups     map[uint16]*rawLookup
	ipHostnames map[string]map[string]struct{}
}

func New() *DNSAnalyzer {
	return &DNSAnalyzer{
		lookups:     make(map[uint16]*rawLookup),
		ipHostnames: make(map[string]map[string]struct{}),
	}
}

func (d *DNSAnalyzer) LayerTypes() []gopacket.LayerType {
	return []gopacket.LayerType{layers.LayerTypeDNS}
}

func (d *DNSAnalyzer) addIPHostnames(l *layers.DNS) {
	// The layer must be a reply.
	if !l.QR {
		return
	}

	// Collect all the hostnames who's IP address are being queried.
	hostnames := make([]string, 0)
	for _, q := range l.Questions {
		if q.Type != layers.DNSTypeA && q.Type != layers.DNSTypeAAAA {
			continue
		}
		hostnames = append(hostnames, string(q.Name))
	}

	// Iterate over all the answers and associate each hostname above with the
	// IP discovered below.
	for _, a := range l.Answers {
		// We only care about IP address lookups.
		if a.Type != layers.DNSTypeA && a.Type != layers.DNSTypeAAAA {
			continue
		}
		// No IP address is present, so continue.
		if a.IP == nil {
			continue
		}
		ip := a.IP.String()
		if _, exists := d.ipHostnames[ip]; !exists {
			d.ipHostnames[ip] = make(map[string]struct{})
		}
		for _, h := range hostnames {
			d.ipHostnames[ip][h] = struct{}{}
		}
	}
}

func (d *DNSAnalyzer) Receive(l gopacket.Layer, p gopacket.Packet) {
	// The layer must be DNS.
	dns, ok := l.(*layers.DNS)
	if !ok {
		return
	}
	if len(dns.Questions) == 0 {
		// skip, no questions
		return
	}

	if _, exists := d.lookups[dns.ID]; !exists {
		d.lookups[dns.ID] = &rawLookup{}
	}
	if dns.QR {
		d.lookups[dns.ID].reply = dns
		d.addIPHostnames(dns)
	} else {
		d.lookups[dns.ID].query = dns
	}
}

// Hostname returns the hostname used to obtain the given IP address.
//
// Returns the hostnames found. Otherwise it returns an empty slice.
func (d *DNSAnalyzer) Hostname(address string) []string {
	// We parse the IP to ensure that it is valid.
	ip := net.ParseIP(address)
	if ip == nil {
		return []string{}
	}

	if hs, exists := d.ipHostnames[ip.String()]; exists {
		hostnames := make([]string, 0)
		for h := range hs {
			hostnames = append(hostnames, h)
		}
		return hostnames
	}

	return []string{}
}
