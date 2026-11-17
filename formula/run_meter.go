package formula

type Meter struct {
	limit int
	spent int
	depth int
}

func NewMeter(limit int) *Meter {
	return &Meter{limit: limit}
}

func (m *Meter) Charge(units int) bool {
	m.spent += units
	return m.spent <= m.limit
}

func (m *Meter) Spent() int { return m.spent }

func (m *Meter) Limit() int { return m.limit }

func (m *Meter) Depth() int { return m.depth }

func (m *Meter) callee(declared int) *Meter {
	return &Meter{limit: m.limit - (m.spent - declared), depth: m.depth + 1}
}
