package jsonform

import (
	"reflect"
	"testing"
)

func TestQuoteKeepsMarkupAndEscapesWhatJSONNeeds(t *testing.T) {
	cases := map[string]string{
		`a<b>&c`:   `"a<b>&c"`,
		`say "hi"`: `"say \"hi\""`,
		"line\nbr": `"line\nbr"`,
		``:         `""`,
	}
	for text, want := range cases {
		if got := string(Quote(text)); got != want {
			t.Errorf("Quote(%q) = %s, want %s", text, got, want)
		}
	}
}

func TestDecodeTakesExactlyOneValue(t *testing.T) {
	good := []string{`{}`, ` [1, 2] `, `null`, `"x"`, `12`}
	for _, text := range good {
		if _, ok := Decode([]byte(text)); !ok {
			t.Errorf("%q was refused", text)
		}
	}
	bad := []string{``, `{`, `{} {}`, `{} x`, `[1,]`}
	for _, text := range bad {
		if _, ok := Decode([]byte(text)); ok {
			t.Errorf("%q was accepted", text)
		}
	}
}

func TestPlainWritesAValueTheWayAPersonReadsIt(t *testing.T) {
	value, _ := Decode([]byte(`{"a": ["x", 1.50, true, null], "b": "t"}`))
	object := value.(map[string]any)
	cases := []struct {
		value any
		want  string
	}{
		{object["b"], "t"},
		{object["a"].([]any)[1], "1.50"},
		{object["a"].([]any)[2], "true"},
		{object["a"].([]any)[3], "null"},
		{object["a"], `["x",1.50,true,null]`},
	}
	for _, c := range cases {
		if got := Plain(c.value); got != c.want {
			t.Errorf("Plain(%#v) = %s, want %s", c.value, got, c.want)
		}
	}
}

func TestWholeNumberRefusesFractionsExponentsAndText(t *testing.T) {
	value, _ := Decode([]byte(`[3, 2.5, 1e2, "4", true]`))
	items := value.([]any)
	want := []bool{true, false, false, false, false}
	for i, item := range items {
		if _, ok := WholeNumber(item); ok != want[i] {
			t.Errorf("item %d: %v, want %v", i, ok, want[i])
		}
	}
	if number, _ := WholeNumber(items[0]); number != 3 {
		t.Errorf("got %d, want 3", number)
	}
}

func TestUnknownKeysAreSorted(t *testing.T) {
	object := map[string]any{"b": 1, "kind": 2, "a": 3, "z": 4}
	if got := UnknownKeys(object, "kind", "z"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("got %v", got)
	}
	if got := UnknownKeys(object, "a", "b", "kind", "z"); got != nil {
		t.Errorf("got %v, want none", got)
	}
}
