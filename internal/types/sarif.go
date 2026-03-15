package types

// SARIFDocument is the top-level SARIF 2.1.0 envelope.
type SARIFDocument struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SARIFRun `json:"runs"`
}

// SARIFRun represents a single analysis invocation.
type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

// SARIFTool describes the analysis tool.
type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

// SARIFDriver contains tool metadata and rule definitions.
type SARIFDriver struct {
	Name           string                     `json:"name"`
	InformationURI string                     `json:"informationUri"`
	Rules          []SARIFReportingDescriptor `json:"rules"`
}

// SARIFReportingDescriptor defines a single rule.
type SARIFReportingDescriptor struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	ShortDescription SARIFMessage `json:"shortDescription"`
	HelpURI          string       `json:"helpUri"`
}

// SARIFResult represents a single finding.
type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	RuleIndex int             `json:"ruleIndex"`
	Level     string          `json:"level"`
	Message   SARIFMessage    `json:"message"`
	Locations []SARIFLocation `json:"locations,omitempty"`
}

// SARIFMessage holds a human-readable text message.
type SARIFMessage struct {
	Text string `json:"text"`
}

// SARIFLocation wraps a physical location.
type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

// SARIFPhysicalLocation contains file and region info.
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

// SARIFArtifactLocation identifies the file.
type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// SARIFRegion identifies the line within the file.
type SARIFRegion struct {
	StartLine int `json:"startLine"`
}
