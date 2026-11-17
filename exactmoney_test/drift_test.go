package exactmoney_test

import "testing"

func TestEM01_PointOnePlusPointTwoIsPointThree(t *testing.T) {
	sum := dec(t, "0.1").Add(dec(t, "0.2"))
	wantText(t, sum, "0.3")
	if !sum.Equal(dec(t, "0.3")) {
		t.Fatalf("0.1 + 0.2 is not equal to 0.3")
	}
}

func TestEM02_AThousandCentsAreTen(t *testing.T) {
	total := dec(t, "0.00")
	cent := dec(t, "0.01")
	for i := 0; i < 1000; i++ {
		total = total.Add(cent)
	}
	wantText(t, total, "10.00")
}

func TestEM03_OrderDoesNotChangeTheSum(t *testing.T) {
	orders := map[string][]string{
		"listed":  {"19.99", "0.01", "5.50", "0.10", "0.20"},
		"reverse": {"0.20", "0.10", "5.50", "0.01", "19.99"},
		"mixed":   {"5.50", "0.20", "19.99", "0.10", "0.01"},
	}
	for name, order := range orders {
		t.Run(name, func(t *testing.T) {
			total := dec(t, "0")
			for _, text := range order {
				total = total.Add(dec(t, text))
			}
			wantEqual(t, total, "25.80")
		})
	}
}

func TestEM04_AThousandAmountsSumTheSameInEveryOrder(t *testing.T) {
	cents := make([]int64, 1000)
	for i := range cents {
		cents[i] = int64((37*(i+1))%1000) + 1
	}
	orders := map[string][]int64{
		"forward":  cents,
		"reverse":  reversed(cents),
		"stride7":  strided(cents, 7),
		"stride9":  strided(cents, 9),
		"stride11": strided(cents, 11),
	}
	for name, order := range orders {
		t.Run(name, func(t *testing.T) {
			total := units(0, 2)
			for _, c := range order {
				total = total.Add(units(c, 2))
			}
			wantText(t, total, "5005.00")
		})
	}
}

func reversed(in []int64) []int64 {
	out := make([]int64, len(in))
	for i, v := range in {
		out[len(in)-1-i] = v
	}
	return out
}

func strided(in []int64, step int) []int64 {
	out := make([]int64, 0, len(in))
	seen := make([]bool, len(in))
	for start := 0; start < step; start++ {
		for i := start; i < len(in); i += step {
			if !seen[i] {
				seen[i] = true
				out = append(out, in[i])
			}
		}
	}
	return out
}
