package detector

import (
	"testing"
)

func TestDetectorFactory_CreateJSDetector(t *testing.T) {
	factory := NewDetectorFactory()

	detector := factory.Create(".js")
	if detector == nil {
		t.Error("Expected JSDetector for .js extension, got nil")
	}

	if _, ok := detector.(*JSDetector); !ok {
		t.Error("Expected JSDetector type for .js extension")
	}
}

func TestDetectorFactory_CreateTSDetector(t *testing.T) {
	factory := NewDetectorFactory()

	detector := factory.Create(".ts")
	if detector == nil {
		t.Error("Expected JSDetector for .ts extension, got nil")
	}

	if _, ok := detector.(*JSDetector); !ok {
		t.Error("Expected JSDetector type for .ts extension")
	}
}

func TestDetectorFactory_CreateGoDetector(t *testing.T) {
	factory := NewDetectorFactory()

	detector := factory.Create(".go")
	if detector == nil {
		t.Error("Expected GoDetector for .go extension, got nil")
	}

	if _, ok := detector.(*GoDetector); !ok {
		t.Error("Expected GoDetector type for .go extension")
	}
}

func TestDetectorFactory_CreatePythonDetector(t *testing.T) {
	factory := NewDetectorFactory()

	detector := factory.Create(".py")
	if detector == nil {
		t.Error("Expected PythonDetector for .py extension, got nil")
	}

	if _, ok := detector.(*PythonDetector); !ok {
		t.Error("Expected PythonDetector type for .py extension")
	}
}

func TestDetectorFactory_CreateUnsupportedExtension(t *testing.T) {
	factory := NewDetectorFactory()

	detector := factory.Create(".txt")
	if detector != nil {
		t.Error("Expected nil for unsupported extension, got detector")
	}
}
