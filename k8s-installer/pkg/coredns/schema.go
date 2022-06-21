package coredns

type TopLevelDomain struct {
	Domain       string `json:"domain" validate:"required,fqdn" description:"domain name unique"`
	Description  string `json:"description,omitempty"`
	DomainCounts int    `json:"domain_counts" description:"auto generator, do not input"`
}

// only use for api docs
type DNSDomainUpdateDocSchema struct {
	TopLevelDomain string                   `json:"top_level_domain" validate:"required,fqdn" description:"top level domain name"`
	Domain         string                   `json:"domain" validate:"required,fqdn" description:"domain name unique"`
	DomainResolve  []DomainResolveDocSchema `json:"domain_resolve"  validate:"required" description:"ip address list of this domain resolve to"`
	Description    string                   `json:"description,omitempty"`
}

// only use for api docs
type DomainResolveDocSchema struct {
	RecordType string `json:"record_type" validate:"omitempty,oneof=A AAAA" enum:"A|AAAA" description:"dns record type"`
	IpAddress  string `json:"ip_address" validate:"ipv4|ipv6" description:"ip address of the domain resolve to"`
}

type DNSDomain struct {
	TopLevelDomain string          `json:"top_level_domain" validate:"required,fqdn" description:"top level domain name"`
	Domain         string          `json:"domain" validate:"required,fqdn" description:"domain name unique"`
	DomainResolve  []DomainResolve `json:"domain_resolve"  validate:"required" description:"ip address list of this domain resolve to"`
	Description    string          `json:"description,omitempty"`
	Action         string          `json:"action,omitempty" description:"auto generator, do not input"`
}

type DomainResolve struct {
	ResolveDomain string `json:"resolve_domain,omitempty" description:"auto generator, do not input"`
	RecordType    string `json:"record_type" validate:"omitempty,oneof=A AAAA" enum:"A|AAAA" description:"dns record type"`
	IpAddress     string `json:"ip_address" validate:"ipv4|ipv6" description:"ip address of the domain resolve to"`
}
