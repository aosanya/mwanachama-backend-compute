package number

type Settings struct {
	Places         int
	Mode           Rounding
	DivisionDigits int
}

func DefaultSettings() Settings {
	return Settings{Places: 2, Mode: RoundHalfUp, DivisionDigits: 20}
}

func (s Settings) Round(d Decimal) Decimal {
	return d.Round(s.Places, s.Mode)
}

func (s Settings) Divide(a, b Decimal) (Decimal, error) {
	return a.Div(b, s.DivisionDigits, s.Mode)
}
