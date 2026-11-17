package exactmoney_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

const generatedSplits = 500

const splitSeed = 20261117

type generatedSplit struct {
	index   int
	places  int
	total   int64
	weights []int64
}

func (g generatedSplit) describe() string {
	return fmt.Sprintf("seed %d, split %d: total %d units at %d places, weights %v", splitSeed, g.index, g.total, g.places, g.weights)
}

func generate(count int) []generatedSplit {
	source := rand.New(rand.NewSource(splitSeed))
	out := make([]generatedSplit, count)
	for i := range out {
		members := 1 + source.Intn(40)
		weights := make([]int64, members)
		equal := source.Intn(2) == 0
		for j := range weights {
			weights[j] = 1
			if !equal {
				weights[j] = int64(source.Intn(9))
			}
		}
		if !equal && sumWeights(weights) == 0 {
			weights[0] = 1
		}
		out[i] = generatedSplit{
			index:   i,
			places:  source.Intn(5),
			total:   source.Int63n(1_000_000_000_000),
			weights: weights,
		}
	}
	return out
}

func sumWeights(weights []int64) int64 {
	var sum int64
	for _, w := range weights {
		sum += w
	}
	return sum
}

func TestGeneratedSplits_SharesPlusLeftOverAreTheTotalAndNeverExceedIt(t *testing.T) {
	for _, g := range generate(generatedSplits) {
		total := units(g.total, g.places)
		result, err := number.Split(total, g.weights, g.places)
		if err != nil {
			t.Fatalf("%s: %v", g.describe(), err)
		}
		if len(result.Shares) != len(g.weights) {
			t.Fatalf("%s: %d shares for %d weights", g.describe(), len(result.Shares), len(g.weights))
		}
		paid := sumOf(result.Shares...)
		if !paid.Add(result.LeftOver).Equal(total) {
			t.Fatalf("%s: shares %s plus left-over %s is not the total %s", g.describe(), paid, result.LeftOver, total)
		}
		if paid.Cmp(total) > 0 {
			t.Fatalf("%s: the payout %s is over the total %s", g.describe(), paid, total)
		}
		if result.LeftOver.Sign() < 0 {
			t.Fatalf("%s: the left-over %s is negative", g.describe(), result.LeftOver)
		}
		for i, share := range result.Shares {
			if share.Sign() < 0 {
				t.Fatalf("%s: share %d is negative", g.describe(), i)
			}
			if share.Scale() != g.places {
				t.Fatalf("%s: share %d has %d places", g.describe(), i, share.Scale())
			}
		}
	}
}

func TestGeneratedSplits_TheLeftOverIsFewerUnitsThanMembersWithShares(t *testing.T) {
	for _, g := range generate(generatedSplits) {
		active := int64(0)
		for _, w := range g.weights {
			if w > 0 {
				active++
			}
		}
		result, err := number.Split(units(g.total, g.places), g.weights, g.places)
		if err != nil {
			t.Fatalf("%s: %v", g.describe(), err)
		}
		if result.LeftOver.Cmp(units(active, g.places)) >= 0 {
			t.Fatalf("%s: the left-over %s is not fewer units than the %d members with a weight", g.describe(), result.LeftOver, active)
		}
	}
}

func TestGeneratedSplits_EqualWeightsGiveEveryMemberTheSameShare(t *testing.T) {
	checked := 0
	for _, g := range generate(generatedSplits) {
		if sumWeights(g.weights) != int64(len(g.weights)) {
			continue
		}
		result, err := number.Split(units(g.total, g.places), g.weights, g.places)
		if err != nil {
			t.Fatalf("%s: %v", g.describe(), err)
		}
		for i, share := range result.Shares {
			if !share.Equal(result.Shares[0]) {
				t.Fatalf("%s: share %d is %s and share 0 is %s", g.describe(), i, share, result.Shares[0])
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no generated split had equal weights")
	}
}

func TestGeneratedSplits_TheSameSeedGivesTheSameSharesTwice(t *testing.T) {
	first := generate(generatedSplits)
	second := generate(generatedSplits)
	for i := range first {
		a, _ := number.Split(units(first[i].total, first[i].places), first[i].weights, first[i].places)
		b, _ := number.Split(units(second[i].total, second[i].places), second[i].weights, second[i].places)
		if a.LeftOver.String() != b.LeftOver.String() {
			t.Fatalf("%s: left-over differs between two runs", first[i].describe())
		}
		for j := range a.Shares {
			if a.Shares[j].String() != b.Shares[j].String() {
				t.Fatalf("%s: share %d differs between two runs", first[i].describe(), j)
			}
		}
	}
}

func TestGeneratedSplits_MemberOrderMovesTheSharesAndNothingElse(t *testing.T) {
	for _, g := range generate(generatedSplits) {
		total := units(g.total, g.places)
		forward, err := number.Split(total, g.weights, g.places)
		if err != nil {
			t.Fatalf("%s: %v", g.describe(), err)
		}
		backward, err := number.Split(total, reversed(g.weights), g.places)
		if err != nil {
			t.Fatalf("%s: %v", g.describe(), err)
		}
		count := len(g.weights)
		for i := range forward.Shares {
			if forward.Shares[i].String() != backward.Shares[count-1-i].String() {
				t.Fatalf("%s: share %d moved with the order", g.describe(), i)
			}
		}
		if forward.LeftOver.String() != backward.LeftOver.String() {
			t.Fatalf("%s: the left-over changed with the order", g.describe())
		}
	}
}

func TestGeneratedSums_AnyOrderGivesTheSameFigure(t *testing.T) {
	source := rand.New(rand.NewSource(splitSeed))
	for run := 0; run < generatedSplits; run++ {
		count := 2 + source.Intn(200)
		places := source.Intn(5)
		amounts := make([]int64, count)
		var independent int64
		for i := range amounts {
			amounts[i] = source.Int63n(10_000_000)
			independent += amounts[i]
		}
		forward := units(0, places)
		for _, a := range amounts {
			forward = forward.Add(units(a, places))
		}
		shuffled := append([]int64(nil), amounts...)
		source.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		other := units(0, places)
		for _, a := range shuffled {
			other = other.Add(units(a, places))
		}
		want := units(independent, places)
		if forward.String() != want.String() || other.String() != want.String() {
			t.Fatalf("seed %d, run %d: forward %s, shuffled %s, want %s", splitSeed, run, forward, other, want)
		}
	}
}
