package numtext

import (
	"strings"
	"testing"
)

type refusedCase struct {
	text   string
	reason Reason
}

func accepted(t *testing.T, kind Type, text string) Value {
	t.Helper()
	value, refusals := Read(kind, One(text))
	if len(refusals) != 0 {
		t.Fatalf("%v %q was refused: %+v", kind, text, refusals)
	}
	return value
}

func refused(t *testing.T, kind Type, text string, reason Reason) {
	t.Helper()
	_, refusals := Read(kind, One(text))
	if len(refusals) != 1 {
		t.Fatalf("%v %q: want 1 refusal, got %+v", kind, text, refusals)
	}
	got := refusals[0]
	if got.Rule != RuleType || got.Type != kind || got.Reason != reason || got.Item != 0 || got.Received != text {
		t.Fatalf("%v %q: got %+v, want rule %q reason %q", kind, text, got, RuleType, reason)
	}
}

func TestIntegerAccepted(t *testing.T) {
	cases := map[string]int64{
		"0":                   0,
		"7":                   7,
		"-7":                  -7,
		"1237":                1237,
		"123456789012345678":  123456789012345678,
		"-123456789012345678": -123456789012345678,
	}
	for text, want := range cases {
		value := accepted(t, Integer, text)
		if value.Type != Integer || value.Integer != want || value.Text != text {
			t.Errorf("%q: got %+v", text, value)
		}
		if value.Decimal.String() != text {
			t.Errorf("%q: decimal reads %q", text, value.Decimal.String())
		}
	}
}

func TestIntegerRefused(t *testing.T) {
	cases := []refusedCase{
		{"-0", ReasonNegativeZero},
		{"+5", ReasonSign},
		{"1237.5", ReasonNotDigits},
		{"1e3", ReasonNotDigits},
		{"1,200", ReasonNotDigits},
		{" ", ReasonNotDigits},
		{"5 ", ReasonNotDigits},
		{"", ReasonEmpty},
		{"-", ReasonNotDigits},
		{"007", ReasonLeadingZero},
		{"1234567890123456789", ReasonTooManyDigits},
	}
	for _, c := range cases {
		refused(t, Integer, c.text, c.reason)
	}
}

func TestDecimalAccepted(t *testing.T) {
	cases := map[string]int{
		"0":                                0,
		"5":                                0,
		"5.125":                            3,
		"45.20":                            2,
		"45.2":                             1,
		"-5.00":                            2,
		"0.4370":                           4,
		"-0.5":                             1,
		"2400.50":                          2,
		"123456789012345678901234567890":   0,
		"1234567890123456789012345.12345":  5,
		"-1234567890123456789012345.12345": 5,
		"0.000000000000000000000000000001": 30,
	}
	for text, places := range cases {
		value := accepted(t, Decimal, text)
		if value.Type != Decimal || value.Text != text {
			t.Errorf("%q: got %+v", text, value)
		}
		if value.Decimal.String() != text {
			t.Errorf("%q: digits changed to %q", text, value.Decimal.String())
		}
		if value.Decimal.Scale() != places {
			t.Errorf("%q: places %d, want %d", text, value.Decimal.Scale(), places)
		}
	}
}

func TestDecimalRefused(t *testing.T) {
	cases := []refusedCase{
		{".5", ReasonMissingWholePart},
		{"5.", ReasonMissingFraction},
		{"1e3", ReasonNotDigits},
		{"1,5", ReasonNotDigits},
		{"+5", ReasonSign},
		{" ", ReasonNotDigits},
		{"", ReasonEmpty},
		{"0123", ReasonLeadingZero},
		{"1,200.50", ReasonNotDigits},
		{"95%", ReasonNotDigits},
		{"1 200,50", ReasonNotDigits},
		{"5.1.2", ReasonNotDigits},
		{"--5", ReasonNotDigits},
		{"-", ReasonNotDigits},
		{"123456789012345678901234567890.1", ReasonTooManyDigits},
		{"1234567890123456789012345678901", ReasonTooManyDigits},
	}
	for _, c := range cases {
		refused(t, Decimal, c.text, c.reason)
	}
}

func TestFloatReadsLikeADecimalThenApproximates(t *testing.T) {
	value := accepted(t, Float, "2400.50")
	if value.Type != Float || value.Float.Value() != 2400.5 || value.Text != "2400.50" {
		t.Fatalf("got %+v", value)
	}
	if accepted(t, Float, "0.1").Float.Value() != 0.1 {
		t.Fatal("0.1 is not read as the nearest float")
	}
	for _, c := range []refusedCase{
		{".5", ReasonMissingWholePart},
		{"5.", ReasonMissingFraction},
		{"1e3", ReasonNotDigits},
		{"+5", ReasonSign},
		{"", ReasonEmpty},
		{"1234567890123456789012345678901", ReasonTooManyDigits},
	} {
		refused(t, Float, c.text, c.reason)
	}
}

func TestTextAccepted(t *testing.T) {
	for _, text := range []string{"", "lapsed", "héllo", "a b", strings.Repeat("a", 200), strings.Repeat("é", 200)} {
		value := accepted(t, Text, text)
		if value.Type != Text || value.Text != text {
			t.Errorf("%q: got %+v", text, value)
		}
	}
}

func TestTextRefused(t *testing.T) {
	refused(t, Text, strings.Repeat("a", 201), ReasonTooLong)
	refused(t, Text, strings.Repeat("é", 201), ReasonTooLong)
	refused(t, Text, "line\nbreak", ReasonControlCharacter)
	refused(t, Text, "tab\there", ReasonControlCharacter)
	refused(t, Text, "bell\a", ReasonControlCharacter)
	refused(t, Text, "nul\x00", ReasonControlCharacter)
	refused(t, Text, "bad\xffbytes", ReasonInvalidText)
}

func TestBoolean(t *testing.T) {
	if !accepted(t, Boolean, "true").Boolean || accepted(t, Boolean, "false").Boolean {
		t.Fatal("true and false are read wrong")
	}
	for _, text := range []string{"True", "TRUE", "yes", "1", "", " true", "false "} {
		reason := ReasonNotBoolean
		if text == "" {
			reason = ReasonEmpty
		}
		refused(t, Boolean, text, reason)
	}
}

func TestDate(t *testing.T) {
	for _, text := range []string{"2026-03-31", "2028-02-29", "0001-01-01", "9999-12-31"} {
		value := accepted(t, Date, text)
		if value.Type != Date || value.Date.String() != text || value.Text != text {
			t.Errorf("%q: got %+v", text, value)
		}
	}
	cases := []refusedCase{
		{"2026-02-30", ReasonNoSuchDay},
		{"2027-02-29", ReasonNoSuchDay},
		{"2026-13-01", ReasonNoSuchDay},
		{"2026-00-10", ReasonNoSuchDay},
		{"31/03/2026", ReasonNotDate},
		{"2026-3-1", ReasonNotDate},
		{"2026-03-31 ", ReasonNotDate},
		{"", ReasonEmpty},
	}
	for _, c := range cases {
		refused(t, Date, c.text, c.reason)
	}
}

func TestListOfIntegers(t *testing.T) {
	value, refusals := Read(IntegerList, Many("1", "2", "3"))
	if len(refusals) != 0 || value.Type != IntegerList || len(value.Items) != 3 || value.Items[2].Integer != 3 {
		t.Fatalf("got %+v %+v", value, refusals)
	}
	if value.Items[0].Type != Integer {
		t.Fatalf("an item reads as the element type, got %v", value.Items[0].Type)
	}
}

func TestEmptyListIsGiven(t *testing.T) {
	value, refusals := Read(DecimalList, Many())
	if len(refusals) != 0 || value.Type != DecimalList || len(value.Items) != 0 {
		t.Fatalf("got %+v %+v", value, refusals)
	}
}

func TestListOfDecimalsKeepsDigits(t *testing.T) {
	value, refusals := Read(DecimalList, Many("2400.50", "3100.25", "45.20"))
	if len(refusals) != 0 {
		t.Fatalf("refused: %+v", refusals)
	}
	for i, want := range []string{"2400.50", "3100.25", "45.20"} {
		if value.Items[i].Decimal.String() != want {
			t.Errorf("item %d: %q", i+1, value.Items[i].Decimal.String())
		}
	}
}

func TestEveryRefusedItemIsReportedWithItsPosition(t *testing.T) {
	_, refusals := Read(DecimalList, Many("200.00", "1,200.50", "100.00", "95%", ".5"))
	if len(refusals) != 3 {
		t.Fatalf("want 3 refusals, got %+v", refusals)
	}
	want := []struct {
		item     int
		received string
		reason   Reason
	}{
		{2, "1,200.50", ReasonNotDigits},
		{4, "95%", ReasonNotDigits},
		{5, ".5", ReasonMissingWholePart},
	}
	for i, w := range want {
		got := refusals[i]
		if got.Item != w.item || got.Received != w.received || got.Reason != w.reason || got.Rule != RuleType || got.Type != Decimal {
			t.Errorf("refusal %d: got %+v", i, got)
		}
	}
}

func TestATextForAListInputIsRefused(t *testing.T) {
	_, refusals := Read(DecimalList, One("200.00"))
	if len(refusals) != 1 || refusals[0].Reason != ReasonTextForList || refusals[0].Type != DecimalList || refusals[0].Item != 0 {
		t.Fatalf("got %+v", refusals)
	}
}

func TestAListForASingleInputIsRefused(t *testing.T) {
	_, refusals := Read(Decimal, Many("200.00"))
	if len(refusals) != 1 || refusals[0].Reason != ReasonListForSingle || refusals[0].Type != Decimal || refusals[0].Item != 0 {
		t.Fatalf("got %+v", refusals)
	}
}

func TestAnUnknownTypeIsRefused(t *testing.T) {
	_, refusals := Read(Type(99), One("1"))
	if len(refusals) != 1 || refusals[0].Reason != ReasonUnknownType {
		t.Fatalf("got %+v", refusals)
	}
}

func TestTheSameTextIsReadTheSameEveryTime(t *testing.T) {
	first := accepted(t, Decimal, "45.20")
	for i := 0; i < 100; i++ {
		again := accepted(t, Decimal, "45.20")
		if !again.Decimal.Equal(first.Decimal) || again.Decimal.String() != "45.20" {
			t.Fatalf("read %d differs: %q", i, again.Decimal.String())
		}
	}
}

func TestNoNumberStyleOtherThanThePlainOneIsConverted(t *testing.T) {
	for _, text := range []string{"1,200.50", "95%", "1 200,50", "1e3", "+5", ".5", "5."} {
		for _, kind := range []Type{Integer, Decimal, Float} {
			if _, refusals := Read(kind, One(text)); len(refusals) != 1 {
				t.Errorf("%v %q was not refused: %+v", kind, text, refusals)
			}
		}
	}
}

func TestRefusalErrorNamesTheInputAndTheReason(t *testing.T) {
	_, refusals := Read(Decimal, One("1,5"))
	message := refusals[0].Error()
	if !strings.Contains(message, "1,5") || !strings.Contains(message, string(ReasonNotDigits)) {
		t.Fatalf("got %q", message)
	}
}
