package problem

import (
	"strings"
	"testing"
)

func TestPlainTextIsShownAsReceived(t *testing.T) {
	for _, received := range []string{"5.125", "-5.00", "0.4370", "", " ", "   ", "active ", "1,200.50"} {
		if got := PlainText(received).Field(); got != received {
			t.Errorf("Field() of %q = %q", received, got)
		}
	}
}

func TestTextOfTwoHundredCharactersIsNotCut(t *testing.T) {
	received := strings.Repeat("a", 200)
	if got := PlainText(received).Field(); got != received {
		t.Fatalf("a text of 200 characters was changed: %d characters", len(got))
	}
}

func TestTextOverTwoHundredCharactersIsCutAtTwoHundredWithDots(t *testing.T) {
	received := strings.Repeat("b", 201)
	want := strings.Repeat("b", 200) + "..."
	if got := PlainText(received).Field(); got != want {
		t.Fatalf("Field() = %d characters, want the first 200 and ...", len(got))
	}
}

func TestTextIsCutByCharactersNotBytes(t *testing.T) {
	received := strings.Repeat("é", 201)
	want := strings.Repeat("é", 200) + "..."
	if got := PlainText(received).Field(); got != want {
		t.Fatalf("a text of 201 characters was cut by bytes")
	}
}

func TestQuotedTextIsShownWithoutQuotesInTheFieldAndWithQuotesInASentence(t *testing.T) {
	v := QuotedText("paused")
	if v.Field() != "paused" {
		t.Errorf("Field() = %q", v.Field())
	}
	if v.Sentence() != `"paused"` {
		t.Errorf("Sentence() = %q", v.Sentence())
	}
}

func TestQuotedTextKeepsTrailingSpaces(t *testing.T) {
	if got := QuotedText("active ").Sentence(); got != `"active "` {
		t.Fatalf("Sentence() = %q", got)
	}
}

func TestQuotedEmptyTextIsTwoQuotes(t *testing.T) {
	if got := QuotedText("").Sentence(); got != `""` {
		t.Fatalf("Sentence() = %q", got)
	}
}

func TestPlainTextIsNotQuotedInASentence(t *testing.T) {
	if got := PlainText("abc").Sentence(); got != "abc" {
		t.Fatalf("Sentence() = %q", got)
	}
}

func TestListIsShownBetweenBracketsSeparatedByCommaAndSpace(t *testing.T) {
	got := List([]string{"2400.50", "3100.25"}).Field()
	if got != "[2400.50, 3100.25]" {
		t.Fatalf("Field() = %q", got)
	}
}

func TestEmptyListIsTwoBrackets(t *testing.T) {
	if got := List(nil).Field(); got != "[]" {
		t.Fatalf("Field() = %q", got)
	}
	if got := List([]string{}).Field(); got != "[]" {
		t.Fatalf("Field() = %q", got)
	}
}

func TestListOfOneItem(t *testing.T) {
	if got := List([]string{"45.25"}).Field(); got != "[45.25]" {
		t.Fatalf("Field() = %q", got)
	}
}

func TestListOfTwentyItemsIsShownInFull(t *testing.T) {
	items := repeated("70", 20)
	want := "[" + strings.Join(items, ", ") + "]"
	if got := List(items).Field(); got != want {
		t.Fatalf("Field() = %q", got)
	}
}

func TestListOfTwentyOneItemsShowsTwentyThenDots(t *testing.T) {
	items := repeated("70", 21)
	want := "[" + strings.Join(items[:20], ", ") + ", ...]"
	if got := List(items).Field(); got != want {
		t.Fatalf("Field() = %q", got)
	}
}

func TestListOfThirteenItemsWithTextIsShownInFull(t *testing.T) {
	items := append(repeated("70", 12), "abc")
	want := "[70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, abc]"
	if got := List(items).Field(); got != want {
		t.Fatalf("Field() = %q", got)
	}
}

func TestListIsTheSameInASentence(t *testing.T) {
	v := List([]string{"420.00", "380.50"})
	if v.Field() != v.Sentence() {
		t.Fatalf("Field() = %q, Sentence() = %q", v.Field(), v.Sentence())
	}
}

func repeated(item string, count int) []string {
	items := make([]string, count)
	for i := range items {
		items[i] = item
	}
	return items
}
