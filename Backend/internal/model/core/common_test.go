package core

import (
	"testing"
	"time"
)

func TestAgeCivilBirthdayAndUnknown(t *testing.T) {
	birth := time.Date(2008, 10, 28, 0, 0, 0, 0, time.UTC)
	a := Actor{BirthDate: &birth}
	if a.Age(time.Date(2026, 10, 27, 23, 59, 0, 0, time.UTC)) != 17 || a.Age(time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)) != 18 {
		t.Fatal("birthday boundary wrong")
	}
	if (Actor{}).Age(time.Now()) != 0 {
		t.Fatal("unknown age must be zero")
	}
}
func TestURLsRejectExecutableSchemesAndCredentials(t *testing.T) {
	for _, u := range []string{"javascript:alert(1)", "data:text/plain,hi", "https://user:secret@example.test/", "//example.test/a"} {
		if ValidateURL(u) == nil {
			t.Fatalf("accepted %s", u)
		}
	}
	if e := ValidateURL("https://example.test/cover.jpg"); e != nil {
		t.Fatal(e)
	}
}
