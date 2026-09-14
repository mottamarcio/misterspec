package contextengine

import "testing"

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
