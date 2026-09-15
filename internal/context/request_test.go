package contextengine

import (
	"errors"
	"testing"
)

func TestValidateIntent_EmptyIsValid(t *testing.T) {
	if err := validateIntent(""); err != nil {
		t.Errorf("validateIntent(\"\") unexpected error: %v", err)
	}
}

func TestValidateIntent_EachRecognizedValueIsValid(t *testing.T) {
	for _, i := range []Intent{IntentPlanning, IntentTasks, IntentImplementation, IntentValidation, IntentAnalysis} {
		if err := validateIntent(i); err != nil {
			t.Errorf("validateIntent(%q) unexpected error: %v", i, err)
		}
	}
}

func TestValidateIntent_UnrecognizedValueIsRejected(t *testing.T) {
	if err := validateIntent(Intent("not-a-real-intent")); err == nil {
		t.Error("validateIntent(\"not-a-real-intent\") expected an error, got nil")
	}
}

func TestValidateIntent_UnrecognizedValueMatchesErrUnsupportedIntent(t *testing.T) {
	err := validateIntent(Intent("not-a-real-intent"))
	if !errors.Is(err, ErrUnsupportedIntent) {
		t.Errorf("validateIntent(\"not-a-real-intent\") = %v, want errors.Is(err, ErrUnsupportedIntent)", err)
	}
}

func TestValidateIntent_RecognizedValuesNeverMatchErrUnsupportedIntent(t *testing.T) {
	for _, i := range []Intent{"", IntentPlanning, IntentTasks, IntentImplementation, IntentValidation, IntentAnalysis} {
		if err := validateIntent(i); errors.Is(err, ErrUnsupportedIntent) {
			t.Errorf("validateIntent(%q) unexpectedly matches ErrUnsupportedIntent", i)
		}
	}
}
