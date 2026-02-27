package models

import (
	"codeberg.org/miekg/dns"
)

// Record is a DNS record that has been queried from a DNS resolver as part of
// a Lookup request.
type Record struct {
	// TTL is the maximum time, in seconds, resolvers are expected to cache a
	// DNS record for.
	TTL int `json:"ttl"`

	// Priority indicates the preference of a MX or SRV DNS record.
	Priority int `json:"priority,omitempty"`

	// Weight is used to balance between SRV DNS records of equal priority.
	Weight int `json:"weight,omitempty"`

	// Port is the network port of a service exposed with an SRV DNS record.
	Port int `json:"port,omitempty"`

	// Tag is used by CAA records to define the type of certificate.
	Tag string `json:"tag,omitempty"`

	// Content is the configuration of a DNS record, such as an IP Address for
	// an A/AAAA record or another name for a CNAME record.
	Content []string `json:"content"`
}

// RecordFromRR converts a records returned by miekg/dns into a Record model.
// If the record isn't supported, nil is returned.
func RecordFromRR(rr dns.RR) *Record {
	switch rr := rr.(type) {
	case *dns.A:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Content: []string{rr.A.Addr.String()},
		}
	case *dns.AAAA:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Content: []string{rr.AAAA.Addr.String()},
		}
	case *dns.CAA:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Tag:     rr.Tag,
			Content: []string{rr.CAA.Value},
		}
	case *dns.CNAME:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Content: []string{rr.CNAME.Target},
		}
	case *dns.DNSKEY:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Content: []string{rr.DNSKEY.String()},
		}
	case *dns.MX:
		return &Record{
			TTL:      int(rr.Hdr.TTL),
			Priority: int(rr.MX.Preference),
			Content:  []string{rr.MX.Mx},
		}
	case *dns.NS:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Content: []string{rr.NS.Ns},
		}
	case *dns.PTR:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Content: []string{rr.PTR.Ptr},
		}
	case *dns.SOA:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Content: []string{rr.SOA.String()},
		}
	case *dns.SRV:
		return &Record{
			TTL:      int(rr.Hdr.TTL),
			Priority: int(rr.SRV.Priority),
			Weight:   int(rr.SRV.Weight),
			Port:     int(rr.SRV.Port),
			Content:  []string{rr.SRV.Target},
		}
	case *dns.SVCB:
		return &Record{
			TTL:      int(rr.Hdr.TTL),
			Priority: int(rr.SVCB.Priority),
			Content:  []string{rr.SVCB.Target},
		}
	case *dns.TXT:
		return &Record{
			TTL:     int(rr.Hdr.TTL),
			Content: rr.TXT.Txt,
		}

	default:
		return nil
	}
}
