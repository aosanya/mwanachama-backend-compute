package number

import "testing"

func TestEM30To39_EachShareIsCutDownAndTheLeftOverStays(t *testing.T) {
	cases := []struct {
		id       string
		total    string
		weights  []int64
		shares   []string
		leftOver string
	}{
		{"EM-30", "100.00", []int64{1, 1, 1}, []string{"33.33", "33.33", "33.33"}, "0.01"},
		{"EM-31", "100.00", []int64{1, 1, 1, 1, 1, 1, 1}, []string{"14.28", "14.28", "14.28", "14.28", "14.28", "14.28", "14.28"}, "0.04"},
		{"EM-32", "0.01", []int64{1, 1, 1}, []string{"0.00", "0.00", "0.00"}, "0.01"},
		{"EM-33", "10.00", []int64{1, 1, 1, 1}, []string{"2.50", "2.50", "2.50", "2.50"}, "0.00"},
		{"EM-34", "1000.00", []int64{1, 1, 1, 1, 1, 1}, []string{"166.66", "166.66", "166.66", "166.66", "166.66", "166.66"}, "0.04"},
		{"EM-36", "50.00", []int64{3, 2}, []string{"30.00", "20.00"}, "0.00"},
		{"EM-37", "10.00", []int64{1, 2}, []string{"3.33", "6.66"}, "0.01"},
		{"EM-38", "0.00", []int64{1, 1, 1}, []string{"0.00", "0.00", "0.00"}, "0.00"},
		{"EM-39", "100.00", []int64{1}, []string{"100.00"}, "0.00"},
		{"EM-102", "-100.00", []int64{1, 1, 1}, []string{"-33.33", "-33.33", "-33.33"}, "-0.01"},
	}
	for _, c := range cases {
		result, err := Split(dec(t, c.total), c.weights, 2)
		if err != nil {
			t.Fatalf("%s: %v", c.id, err)
		}
		if len(result.Shares) != len(c.shares) {
			t.Fatalf("%s: %d shares, want %d", c.id, len(result.Shares), len(c.shares))
		}
		sum := dec(t, "0.00")
		for i, want := range c.shares {
			if result.Shares[i].String() != want {
				t.Errorf("%s share %d = %s, want %s", c.id, i, result.Shares[i], want)
			}
			sum = sum.Add(result.Shares[i])
		}
		if result.LeftOver.String() != c.leftOver {
			t.Errorf("%s left-over = %s, want %s", c.id, result.LeftOver, c.leftOver)
		}
		if !sum.Add(result.LeftOver).Equal(dec(t, c.total)) {
			t.Errorf("%s: shares plus left-over = %s, want %s", c.id, sum.Add(result.LeftOver), c.total)
		}
	}
}

func TestEM35_EqualWeightsGiveTheEqualSplit(t *testing.T) {
	weighted, err := Split(dec(t, "100.00"), []int64{1, 1, 1}, 2)
	if err != nil {
		t.Fatal(err)
	}
	equal, err := EqualSplit(dec(t, "100.00"), 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	for i := range weighted.Shares {
		if !weighted.Shares[i].Equal(equal.Shares[i]) {
			t.Fatalf("share %d differs between weighted and equal split", i)
		}
	}
	if !weighted.LeftOver.Equal(equal.LeftOver) {
		t.Fatal("left-over differs between weighted and equal split")
	}
}

func TestEM40To45_SplitInvariants(t *testing.T) {
	totals := []string{"100.00", "0.01", "10.00", "1000.00", "37.77", "0.00"}
	for _, total := range totals {
		for members := 1; members <= 9; members++ {
			result, err := EqualSplit(dec(t, total), members, 2)
			if err != nil {
				t.Fatal(err)
			}
			sum := dec(t, "0.00")
			for _, share := range result.Shares {
				sum = sum.Add(share)
				if share.Sign() < 0 {
					t.Fatalf("%s among %d: negative share", total, members)
				}
				if !share.Equal(result.Shares[0]) {
					t.Fatalf("%s among %d: shares differ, nobody gets an extra cent", total, members)
				}
			}
			if !sum.Add(result.LeftOver).Equal(dec(t, total)) {
				t.Fatalf("%s among %d: shares plus left-over do not add back", total, members)
			}
			limit := NewDecimal(int64(members), 2)
			if result.LeftOver.Sign() < 0 || result.LeftOver.Cmp(limit) >= 0 {
				t.Fatalf("%s among %d: left-over %s is not below %d cents", total, members, result.LeftOver, members)
			}
			again, _ := EqualSplit(dec(t, total), members, 2)
			for i := range again.Shares {
				if !again.Shares[i].Equal(result.Shares[i]) {
					t.Fatalf("%s among %d: a second run gave other shares", total, members)
				}
			}
		}
	}
}

func TestSplitFollowsThePlacesOfTheCurrency(t *testing.T) {
	zero, err := EqualSplit(dec(t, "100"), 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, zero.Shares[0], "33")
	wantText(t, zero.LeftOver, "1")
	three, err := EqualSplit(dec(t, "100.000"), 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, three.Shares[0], "33.333")
	wantText(t, three.LeftOver, "0.001")
}

func TestSplitRefusesNoMembersAndBadWeights(t *testing.T) {
	if _, err := EqualSplit(dec(t, "10.00"), 0, 2); err == nil {
		t.Fatal("no members must be refused")
	}
	if _, err := Split(dec(t, "10.00"), []int64{1, -1}, 2); err == nil {
		t.Fatal("a negative weight must be refused")
	}
	if _, err := Split(dec(t, "10.00"), []int64{0, 0}, 2); err == nil {
		t.Fatal("weights that sum to zero must be refused")
	}
	if _, err := Split(dec(t, "10.00"), nil, 2); err == nil {
		t.Fatal("no weights must be refused")
	}
}
