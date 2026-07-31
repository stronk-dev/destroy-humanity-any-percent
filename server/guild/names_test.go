package guild

import "testing"

func TestNormalizeGuildNameAndDenylist(t *testing.T) {
	cases := map[string]string{
		"  Small\tSystems  ": "small systems",
		"ＳＭＡＬＬ systems":      "small systems",
		"a_b-c":              "a_b-c",
	}
	for input, expected := range cases {
		actual, ok := NormalizeGuildName(input)
		if !ok || actual != expected {
			t.Fatalf("NormalizeGuildName(%q)=(%q,%v), want (%q,true)", input, actual, ok, expected)
		}
	}
	for _, input := range []string{"--bad", "bad_", "x", "no.dots", "café", "___"} {
		if actual, ok := NormalizeGuildName(input); ok {
			t.Fatalf("NormalizeGuildName(%q)=(%q,true)", input, actual)
		}
	}
	validator, err := NewDenylistNameValidator([]byte("admin\nmoderator\n"), []byte("house name\nadmin\n"))
	if err != nil {
		t.Fatal(err)
	}
	if validator.ValidateGuildName("small admin club") || !validator.ValidateGuildName("small systems") {
		t.Fatal("denylist substring policy mismatch")
	}
}
