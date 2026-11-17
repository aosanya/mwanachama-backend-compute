package exactmoney_test

import (
	"fmt"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

type splitRow struct {
	id       string
	total    int64
	weights  []int64
	shares   []int64
	leftOver int64
}

func equalWeights(members int) []int64 {
	weights := make([]int64, members)
	for i := range weights {
		weights[i] = 1
	}
	return weights
}

func repeat(count int, value int64) []int64 {
	out := make([]int64, count)
	for i := range out {
		out[i] = value
	}
	return out
}

var splitRows = []splitRow{
	{"EM-30", 10000, equalWeights(3), repeat(3, 3333), 1},
	{"EM-31", 10000, equalWeights(7), repeat(7, 1428), 4},
	{"EM-32", 1, equalWeights(3), repeat(3, 0), 1},
	{"EM-33", 1000, equalWeights(4), repeat(4, 250), 0},
	{"EM-34", 100000, equalWeights(6), repeat(6, 16666), 4},
	{"EM-35", 10000, []int64{1, 1, 1}, repeat(3, 3333), 1},
	{"EM-36", 5000, []int64{3, 2}, []int64{3000, 2000}, 0},
	{"EM-37", 1000, []int64{1, 2}, []int64{333, 666}, 1},
	{"EM-38", 0, equalWeights(3), repeat(3, 0), 0},
	{"EM-39", 10000, equalWeights(1), []int64{10000}, 0},
}

func TestSplitCases_C1_AtEveryPlaceCount(t *testing.T) {
	for _, row := range splitRows {
		t.Run(row.id, func(t *testing.T) {
			perPlaces(t, func(t *testing.T, places int) {
				result, err := number.Split(units(row.total, places), row.weights, places)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Shares) != len(row.shares) {
					t.Fatalf("%d shares, want %d", len(result.Shares), len(row.shares))
				}
				for i, want := range row.shares {
					if result.Shares[i].String() != units(want, places).String() {
						t.Errorf("share %d is %s, want %s", i, result.Shares[i], units(want, places))
					}
				}
				wantText(t, result.LeftOver, units(row.leftOver, places).String())
			})
		})
	}
}

func TestSplitCases_C1_WrittenAsCurrency(t *testing.T) {
	result, err := number.EqualSplit(dec(t, "100.00"), 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, share := range result.Shares {
		wantText(t, share, "33.33")
	}
	wantText(t, result.LeftOver, "0.01")
	result, err = number.EqualSplit(dec(t, "1000.00"), 6, 2)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, result.Shares[0], "166.66")
	wantText(t, result.LeftOver, "0.04")
}

func TestSplitInvariants_C2(t *testing.T) {
	for _, row := range splitRows {
		t.Run(row.id, func(t *testing.T) {
			perPlaces(t, func(t *testing.T, places int) {
				total := units(row.total, places)
				first, err := number.Split(total, row.weights, places)
				if err != nil {
					t.Fatal(err)
				}
				paid := sumOf(first.Shares...)
				if !paid.Add(first.LeftOver).Equal(total) {
					t.Errorf("EM-40: shares %s plus left-over %s is not %s", paid, first.LeftOver, total)
				}
				if first.LeftOver.Sign() < 0 {
					t.Errorf("EM-41: left-over %s is negative", first.LeftOver)
				}
				if paid.Cmp(total) > 0 {
					t.Errorf("the payout %s is over the total %s", paid, total)
				}
				for i, share := range first.Shares {
					if share.Sign() < 0 {
						t.Errorf("EM-42: share %d is negative", i)
					}
				}
				second, _ := number.Split(total, row.weights, places)
				for i := range first.Shares {
					if first.Shares[i].String() != second.Shares[i].String() {
						t.Errorf("EM-44: share %d differs between two runs", i)
					}
				}
				if first.LeftOver.String() != second.LeftOver.String() {
					t.Errorf("EM-44: left-over differs between two runs")
				}
			})
		})
	}
}

func TestSplitInvariants_EqualSplitHasNoExtraUnit(t *testing.T) {
	for _, row := range splitRows[:5] {
		t.Run(row.id, func(t *testing.T) {
			perPlaces(t, func(t *testing.T, places int) {
				members := len(row.weights)
				result, err := number.EqualSplit(units(row.total, places), members, places)
				if err != nil {
					t.Fatal(err)
				}
				for i, share := range result.Shares {
					if !share.Equal(result.Shares[0]) {
						t.Errorf("EM-43: share %d is %s and share 0 is %s", i, share, result.Shares[0])
					}
				}
				if result.LeftOver.Cmp(units(int64(members), places)) >= 0 {
					t.Errorf("EM-41: left-over %s is not fewer units than %d members", result.LeftOver, members)
				}
			})
		})
	}
}

func TestEM45_MemberOrderChangesNoShareAndNoTotal(t *testing.T) {
	for _, row := range []splitRow{splitRows[0], splitRows[1], splitRows[7]} {
		t.Run(row.id, func(t *testing.T) {
			weights := row.weights
			flipped := reversed(weights)
			total := units(row.total, 2)
			forward, err := number.Split(total, weights, 2)
			if err != nil {
				t.Fatal(err)
			}
			backward, err := number.Split(total, flipped, 2)
			if err != nil {
				t.Fatal(err)
			}
			count := len(weights)
			for i := range forward.Shares {
				if forward.Shares[i].String() != backward.Shares[count-1-i].String() {
					t.Errorf("share %d moved with the order", i)
				}
			}
			if !sumOf(forward.Shares...).Equal(sumOf(backward.Shares...)) {
				t.Errorf("the total of shares changed with the order")
			}
		})
	}
}

func TestEM50_TheLeftOverStaysWhereThePayoutCameFrom(t *testing.T) {
	cases := []struct {
		members      int
		share        string
		payout       string
		leftOver     string
		accountTotal string
	}{
		{3, "33.33", "99.99", "0.01", "100.00"},
		{7, "14.28", "99.96", "0.04", "100.00"},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("100.00 among %d", c.members), func(t *testing.T) {
			result, err := number.EqualSplit(dec(t, "100.00"), c.members, 2)
			if err != nil {
				t.Fatal(err)
			}
			for _, share := range result.Shares {
				wantText(t, share, c.share)
			}
			payout := sumOf(result.Shares...)
			wantEqual(t, payout, c.payout)
			wantText(t, result.LeftOver, c.leftOver)
			wantEqual(t, payout.Add(result.LeftOver), c.accountTotal)
		})
	}
}

func TestEM92_AZeroTotalSplitsCleanly(t *testing.T) {
	result, err := number.EqualSplit(dec(t, "0.00"), 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, share := range result.Shares {
		wantText(t, share, "0.00")
	}
	wantText(t, result.LeftOver, "0.00")
}

func TestEM102_ANegativeTotalCutsTowardZero_OpenQuestion2(t *testing.T) {
	result, err := number.EqualSplit(dec(t, "-100.00"), 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, share := range result.Shares {
		wantText(t, share, "-33.33")
	}
	wantText(t, result.LeftOver, "-0.01")
	wantEqual(t, sumOf(result.Shares...).Add(result.LeftOver), "-100.00")
}

func TestSplitRefusals(t *testing.T) {
	if _, err := number.EqualSplit(dec(t, "10.00"), 0, 2); err == nil {
		t.Error("a split among no members was accepted")
	}
	if _, err := number.Split(dec(t, "10.00"), []int64{0, 0}, 2); err == nil {
		t.Error("weights that are all zero were accepted")
	}
	if _, err := number.Split(dec(t, "10.00"), []int64{1, -1}, 2); err == nil {
		t.Error("a negative weight was accepted")
	}
}
