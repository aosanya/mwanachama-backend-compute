package laws

import (
	"fmt"
	"math/rand"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

const (
	PlantedMissing   = "missing"
	PlantedWrongType = "wrong-type"
	MostPlanted      = 5
	MostRows         = 200
)

type Planted struct {
	Input string
	Kind  string
}

type Row struct {
	Values  map[string]formula.Value
	Planted []Planted
}

func (r Row) Plants(input string, kind string) bool {
	for _, planted := range r.Planted {
		if planted.Input == input && planted.Kind == kind {
			return true
		}
	}
	return false
}

func NewRows(seed int64, inputs []formula.Input) []Row {
	rng := rand.New(rand.NewSource(seed ^ 0x5eed5eed))
	rows := make([]Row, 1+rng.Intn(MostRows))
	for i := range rows {
		rows[i] = newRow(rng, inputs)
	}
	return rows
}

func newRow(rng *rand.Rand, inputs []formula.Input) Row {
	row := Row{Values: map[string]formula.Value{}}
	for _, input := range inputs {
		row.Values[input.Name] = valueOf(rng, input.Type)
	}
	if rng.Intn(10) < 6 {
		return row
	}
	count := 1 + rng.Intn(MostPlanted)
	if count > len(inputs) {
		count = len(inputs)
	}
	for _, index := range rng.Perm(len(inputs))[:count] {
		input := inputs[index]
		if rng.Intn(2) == 0 {
			delete(row.Values, input.Name)
			row.Planted = append(row.Planted, Planted{Input: input.Name, Kind: PlantedMissing})
			continue
		}
		row.Values[input.Name] = wrongValueOf(rng, input.Type)
		row.Planted = append(row.Planted, Planted{Input: input.Name, Kind: PlantedWrongType})
	}
	return row
}

func wrongValueOf(rng *rand.Rand, kind formula.Type) formula.Value {
	if kind == formula.TypeText {
		return formula.IntegerValue(int64(rng.Intn(1000)))
	}
	return formula.TextValue(fmt.Sprintf("abc%d", rng.Intn(100)))
}

func valueOf(rng *rand.Rand, kind formula.Type) formula.Value {
	switch kind {
	case formula.TypeInteger:
		return formula.IntegerValue(int64(rng.Intn(20000)))
	case formula.TypeDecimal:
		return formula.DecimalValue(number.NewDecimal(int64(rng.Intn(2000000)), 2))
	case formula.TypeFloat:
		value, _ := number.NewFloat(float64(rng.Intn(100000)) / 997)
		return formula.FloatValue(value)
	case formula.TypeText:
		return formula.TextValue(textPool[rng.Intn(len(textPool))])
	case formula.TypeBoolean:
		return formula.BooleanValue(rng.Intn(2) == 0)
	case formula.TypeDate:
		date, _ := datefn.Parse(fmt.Sprintf("%04d-%02d-%02d", 2020+rng.Intn(8), 1+rng.Intn(12), 1+rng.Intn(28)))
		return formula.DateValue(date)
	case formula.TypeListInteger:
		items := make([]int64, 1+rng.Intn(5))
		for i := range items {
			items[i] = int64(rng.Intn(5000))
		}
		return formula.IntegerListValue(items)
	}
	items := make([]number.Decimal, 1+rng.Intn(5))
	for i := range items {
		items[i] = number.NewDecimal(int64(rng.Intn(500000)), 2)
	}
	return formula.DecimalListValue(items)
}
