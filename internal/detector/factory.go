package detector

import (
	"github.com/capture-env-analyzer/internal/types"
)

// DetectorFactory creates language-specific detectors based on file extension
type DetectorFactory struct{}

// NewDetectorFactory creates a new DetectorFactory
func NewDetectorFactory() *DetectorFactory {
	return &DetectorFactory{}
}

// Create returns the appropriate detector for the given file extension
func (f *DetectorFactory) Create(extension string) types.LanguageDetector {
	switch extension {
	case ".js", ".ts":
		return NewJSDetector()
	case ".go":
		return NewGoDetector()
	case ".py":
		return NewPythonDetector()
	default:
		return nil
	}
}
