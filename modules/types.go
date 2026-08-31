// Package modules contains the lint rule implementations for Specula.
//
// Each module scans project files and produces Findings with a severity level.
// Modules accept a ModuleInput and return a ModuleResult. They never call
// os.Exit or write to stdout directly.
package modules

// Severity levels for findings.
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarn
	SeverityFail
)

// String returns the display string for a severity.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarn:
		return "WARN"
	case SeverityFail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// Finding is a single issue discovered by a module.
type Finding struct {
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	File     string   `json:"file,omitempty"`   // relative path
	Line     int      `json:"line,omitempty"`   // 1-indexed
	Detail   string   `json:"detail,omitempty"` // extra context
}

// ModuleInput is the data provided to each lint module.
type ModuleInput struct {
	RootDir string      // absolute path to the project root
	Files   []FileEntry // scanned files from the walker
}

// FileEntry is a minimal representation of a file for module consumption.
type FileEntry struct {
	Path    string // relative path from root
	AbsPath string // absolute path
	Size    int64
}

// ModuleResult is the output from a single lint module.
type ModuleResult struct {
	Name     string    `json:"name"`
	Findings []Finding `json:"findings"`
}

// CountBySeverity returns counts of findings grouped by severity.
func (r ModuleResult) CountBySeverity() (fails, warns, infos int) {
	for _, f := range r.Findings {
		switch f.Severity {
		case SeverityFail:
			fails++
		case SeverityWarn:
			warns++
		case SeverityInfo:
			infos++
		}
	}
	return
}
