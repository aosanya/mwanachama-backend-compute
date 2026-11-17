package number

import "testing"

func TestDefaultsAreHalfUpTwoPlacesAndTwentyDigits(t *testing.T) {
	s := DefaultSettings()
	if s.Mode != RoundHalfUp || s.Places != 2 || s.DivisionDigits != 20 {
		t.Fatalf("defaults = %+v, want half-up, 2 places, 20 digits", s)
	}
}

func TestSettingsRoundAndDivideByTheirOwnValues(t *testing.T) {
	s := DefaultSettings()
	wantText(t, s.Round(dec(t, "2.345")), "2.35")
	third, err := s.Divide(dec(t, "1"), dec(t, "3"))
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, third, "0.33333333333333333333")
	short := Settings{Places: 2, Mode: RoundHalfUp, DivisionDigits: 6}
	sixth, err := short.Divide(dec(t, "1"), dec(t, "6"))
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, sixth, "0.166667")
}

func TestSettingsWithOtherPlacesAndModes(t *testing.T) {
	zeroPlaces := Settings{Places: 0, Mode: RoundHalfEven, DivisionDigits: 20}
	wantText(t, zeroPlaces.Round(dec(t, "2.5")), "2")
	wantText(t, zeroPlaces.Round(dec(t, "3.5")), "4")
	fourPlaces := Settings{Places: 4, Mode: RoundDown, DivisionDigits: 20}
	wantText(t, fourPlaces.Round(dec(t, "0.00019")), "0.0001")
}

func TestTheSameInputGivesTheSameFigureEveryRun(t *testing.T) {
	first, _ := dec(t, "2").Div(dec(t, "7"), 20, RoundHalfUp)
	for i := 0; i < 50; i++ {
		again, _ := dec(t, "2").Div(dec(t, "7"), 20, RoundHalfUp)
		if again.String() != first.String() {
			t.Fatal("the same division gave two figures")
		}
	}
}
