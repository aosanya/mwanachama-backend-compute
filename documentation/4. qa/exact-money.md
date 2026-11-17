# Exact money: test design

Status: draft. A test design for owner review. Nothing here is built, and nothing here chooses an open rule.

All names and figures are invented. Amounts are shown with 2 places. That is a working assumption for this page only (open question 3).

## What this protects

No balance, interest or share is ever off by a fraction. Any drift is caught by a test before a person sees it. Every case below has an id, its inputs, the exact expected result, the working beside the figure, and the rule it protects. Where the right answer depends on the owner's decision, the case checks the part that does not depend on it and says which part waits.

## Open questions for the owner

1. Rounding modes. Which modes does the formula language offer, which are allowed in an agency's formula, and which is the default? The eight candidates are defined under group B. The tests run all eight and choose none.
2. Remainder of a split. When a total does not divide evenly, where do the leftover cents go? Candidates are listed under group C. Related: the base share of a negative total (this page assumes truncation toward zero), and the tie-break when two members have equal fractions.
3. Places per currency. The page assumes 2. If a currency can have 0, 3 or 4 places, every case repeats per place count.
4. Non-terminating division. 1 divided by 3 has no exact decimal. Candidates: refuse; round to a stated number of places with a named mode; carry a rational number.
5. Negative inputs. May an input be negative? Cases that need it are marked "only if negative inputs are allowed".
6. Error wording. The exact sentence for mixing decimal and float is not decided. Candidate wordings are not fixed here.
7. Fast-path boundary. The size of the largest whole number of smallest units held on the fast path is set by the engine choice (the engine spike). The page calls it B.
8. Reading a float on conversion. Does an explicit float-to-decimal conversion read the float's exact binary value or its shortest printed form? For float 1.005 to 2 places under half-up the answers are 1.00 and 1.01.

## A. Drift

Rule: decimal arithmetic is exact. No step adds, drops or moves a fraction.

| Id | Inputs | Expected | Working | Protects |
|---|---|---|---|---|
| EM-01 | 0.1 + 0.2 | 0.3 exactly, equal to the literal 0.3 | 1 tenth + 2 tenths = 3 tenths. Binary float gives 0.30000000000000004, which is not equal to 0.3 | no float in the sum |
| EM-02 | 0.01 added 1,000 times, starting from 0.00 | 10.00 exactly | 1,000 x 1 cent = 1,000 cents = 10.00. Binary float gives 9.999999999999831 | repeated addition does not drift |
| EM-03 | 19.99, 0.01, 5.50, 0.10, 0.20 summed in listed order, reverse order, and order 5.50, 0.20, 19.99, 0.10, 0.01 | 25.80 in all three orders | In cents: 1,999 + 1 + 550 + 10 + 20 = 2,580 = 25.80. Binary float gives 25.8 forward and 25.799999999999997 in reverse | order does not change the sum |
| EM-04 | 1,000 amounts: amount i is ((37 x i) mod 1,000) + 1 cents, for i = 1 to 1,000. Sum in forward order, reverse order and 3 fixed shuffles | 5,005.00 in all five orders | 37 shares no factor with 1,000, so 37 x i mod 1,000 visits each of 0 to 999 once. The amounts are 1 to 1,000 cents once each: 1,000 x 1,001 / 2 = 500,500 cents = 5,005.00. Binary float, reverse order: 5004.999999999999 | order does not change a long sum |

## B. Rounding

Rule: for a named mode and a stated number of places, the result is one exact value. A figure at the half rounds the same way every run.

The eight candidate modes. Each is defined by what it does at the half, or, for the last four, always.

| Column | Meaning |
|---|---|
| half-up | at the half, away from zero (2.5 gives 3, -2.5 gives -3) |
| half-down | at the half, toward zero (2.5 gives 2, -2.5 gives -2) |
| half-even | at the half, to the neighbour whose last kept digit is even (banker's) |
| half-+inf | at the half, toward positive infinity (2.5 gives 3, -2.5 gives -2): half-up on the number line |
| up | any fraction, away from zero |
| down | any fraction, toward zero (truncate) |
| ceiling | any fraction, toward positive infinity |
| floor | any fraction, toward negative infinity |

half-up and half-+inf give the same result for every positive input and differ only at the half of a negative input. Libraries use the name "half-up" for both, so the tests name the behaviour, not only the label.

### B1. Inputs, to 0 places

| Id | Input | Lies between | Position |
|---|---|---|---|
| EM-10 | 2.5 | 2 and 3 | at the half |
| EM-11 | 1.5 | 1 and 2 | at the half |
| EM-12 | 0.5 | 0 and 1 | at the half |
| EM-13 | -0.5 | -1 and 0 | at the half |
| EM-14 | -1.5 | -2 and -1 | at the half |
| EM-15 | -2.5 | -3 and -2 | at the half |
| EM-16 | 2.4 | 2 and 3 | below the half |
| EM-17 | 2.6 | 2 and 3 | above the half |
| EM-18 | -2.4 | -3 and -2 | below the half (nearer -2) |
| EM-19 | -2.6 | -3 and -2 | above the half (nearer -3) |

Expected results, to 0 places:

| Id | half-up | half-down | half-even | half-+inf | up | down | ceiling | floor |
|---|---|---|---|---|---|---|---|---|
| EM-10 | 3 | 2 | 2 | 3 | 3 | 2 | 3 | 2 |
| EM-11 | 2 | 1 | 2 | 2 | 2 | 1 | 2 | 1 |
| EM-12 | 1 | 0 | 0 | 1 | 1 | 0 | 1 | 0 |
| EM-13 | -1 | 0 | 0 | 0 | -1 | 0 | 0 | -1 |
| EM-14 | -2 | -1 | -2 | -1 | -2 | -1 | -1 | -2 |
| EM-15 | -3 | -2 | -2 | -2 | -3 | -2 | -2 | -3 |
| EM-16 | 2 | 2 | 2 | 2 | 3 | 2 | 3 | 2 |
| EM-17 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 2 |
| EM-18 | -2 | -2 | -2 | -2 | -3 | -2 | -2 | -3 |
| EM-19 | -3 | -3 | -3 | -3 | -3 | -2 | -2 | -3 |

Negative inputs in this group run only if negative inputs are allowed (open question 5).

### B2. Inputs, to 2 places

Float trap: the number is not held exactly in binary float, and the float's true value sits on the other side of the half, so a float rounds it the wrong way.

| Id | Input | Lies between | Position | Float trap |
|---|---|---|---|---|
| EM-20 | 1.005 | 1.00 and 1.01 | at the half | yes. Float holds 1.00499999999999989... Float rounding gives 1.0 |
| EM-21 | 1.015 | 1.01 and 1.02 | at the half | yes. Float holds 1.01499999999999990... Float rounding gives 1.01 |
| EM-22 | 1.025 | 1.02 and 1.03 | at the half | yes. Float holds 1.02499999999999991... Float rounding gives 1.02 |
| EM-23 | 1.004 | 1.00 and 1.01 | below the half | no |
| EM-24 | 1.006 | 1.00 and 1.01 | above the half | no |
| EM-25 | -1.005 | -1.01 and -1.00 | at the half | yes. Float holds -1.00499999999999989... |
| EM-26 | -1.025 | -1.03 and -1.02 | at the half | yes. Float holds -1.02499999999999991... |

Expected results, to 2 places:

| Id | half-up | half-down | half-even | half-+inf | up | down | ceiling | floor |
|---|---|---|---|---|---|---|---|---|
| EM-20 | 1.01 | 1.00 | 1.00 | 1.01 | 1.01 | 1.00 | 1.01 | 1.00 |
| EM-21 | 1.02 | 1.01 | 1.02 | 1.02 | 1.02 | 1.01 | 1.02 | 1.01 |
| EM-22 | 1.03 | 1.02 | 1.02 | 1.03 | 1.03 | 1.02 | 1.03 | 1.02 |
| EM-23 | 1.00 | 1.00 | 1.00 | 1.00 | 1.01 | 1.00 | 1.01 | 1.00 |
| EM-24 | 1.01 | 1.01 | 1.01 | 1.01 | 1.01 | 1.00 | 1.01 | 1.00 |
| EM-25 | -1.01 | -1.00 | -1.00 | -1.00 | -1.01 | -1.00 | -1.00 | -1.01 |
| EM-26 | -1.03 | -1.02 | -1.02 | -1.02 | -1.03 | -1.02 | -1.02 | -1.03 |

Working for half-even: 1.015 sits between 1.01 and 1.02, and 1.02 ends in the even digit 2, so it is chosen. 1.025 sits between 1.02 and 1.03, and 1.02 is even. 1.005 sits between 1.00 and 1.01, and 1.00 ends in 0, which is even.

### B3. Zero has no sign

| Id | Input | Expected | Working | Protects |
|---|---|---|---|---|
| EM-27 | -0.004 to 2 places, under half-up, half-down, half-even, half-+inf, down and ceiling | 0.00, printed without a minus sign | -0.004 is below the half of 0.01, so it lands on zero. A decimal type keeps the sign and prints -0.00 | a balance never shows -0.00 |

Only if negative inputs are allowed. Under up and floor the result is -0.01.

## C. Splitting a total

Rule: the shares add back to the total. The remainder is never lost and never invented.

Working assumption for the tables: the base share is the total divided by the number of members, cut off after 2 places toward zero. The remainder is the total minus the sum of the base shares, counted in cents.

### C1. Cases

| Id | Total, members | Base share | Working | Remainder | Check: shares plus remainder |
|---|---|---|---|---|---|
| EM-30 | 100.00 among 3 | 33.33 | 10,000 cents / 3 = 3,333 each, 3 x 3,333 = 9,999 | 1 cent | 99.99 + 0.01 = 100.00 |
| EM-31 | 100.00 among 7 | 14.28 | 10,000 / 7 = 1,428 each, 7 x 1,428 = 9,996 | 4 cents | 99.96 + 0.04 = 100.00 |
| EM-32 | 0.01 among 3 | 0.00 | 1 cent / 3 = 0 each, 3 x 0 = 0 | 1 cent | 0.00 + 0.01 = 0.01 |
| EM-33 | 10.00 among 4 | 2.50 | 1,000 / 4 = 250 each, 4 x 250 = 1,000 | 0 cents | 10.00 + 0.00 = 10.00 |
| EM-34 | 1,000.00 among 6 | 166.66 | 100,000 / 6 = 16,666 each, 6 x 16,666 = 99,996 | 4 cents | 999.96 + 0.04 = 1,000.00 |
| EM-35 | 100.00, weights 1:1:1 | 33.33, 33.33, 33.33 | weights sum to 3, each is 10,000 x 1 / 3 = 3,333, sum 9,999 | 1 cent | 99.99 + 0.01 = 100.00 |
| EM-36 | 50.00, weights 3:2 | 30.00, 20.00 | weights sum to 5: 5,000 x 3 / 5 = 3,000 and 5,000 x 2 / 5 = 2,000 | 0 cents | 50.00 + 0.00 = 50.00 |
| EM-37 | 10.00, weights 1:2 | 3.33, 6.66 | weights sum to 3: 1,000 x 1 / 3 = 333 and 1,000 x 2 / 3 = 666, sum 999 | 1 cent | 9.99 + 0.01 = 10.00 |
| EM-38 | 0.00 among 3 | 0.00 each | 0 / 3 = 0 | 0 cents | 0.00 |
| EM-39 | 100.00 among 1 | 100.00 | 10,000 / 1 = 10,000 | 0 cents | 100.00 |

EM-37 is added beyond the issue's list: it is the only weighted case with a remainder. The split of a negative total is EM-102.

### C2. Invariant checks, apart from where the remainder goes

These hold under every candidate placement below and are tested without choosing one.

| Id | Check | Applies to |
|---|---|---|
| EM-40 | Sum of shares, plus any amount held outside the members, equals the total exactly | every case in C1 and EM-102 |
| EM-41 | Remainder is 0 or more cents and fewer cents than there are members | EM-30 to EM-39 |
| EM-42 | No share is negative when the total is not negative | EM-30 to EM-39 |
| EM-43 | Each share differs from the base share by 0 or 1 cent | candidates 2, 3, 5 and 6 below only. Candidates 1 and 4 give one member the whole remainder, so EM-31 shows a 4 cent difference. Under candidate 5 members hold exactly the base share |
| EM-44 | Running the same split twice gives the same shares | every case, with the seed recorded for candidate 6 |
| EM-45 | Member order in the list does not change the total of shares, only who holds which cent | EM-30, EM-31 |

### C3. Remainder placement: depends on the owner's answer to open question 2

The candidates:

1. The first member gets all of it.
2. One cent each to the first r members in list order (r is the remainder in cents).
3. One cent each to the r members with the largest fractional parts (largest remainder method). A tie-break is then needed (open question 2).
4. The last member absorbs it.
5. A named account, such as a rounding account, holds it.
6. Member order randomised with a recorded seed, then candidate 2.

Shown for two cases only. Each row adds to 100.00.

| Id | Candidate | 100.00 among 3 (r = 1) | 100.00 among 7 (r = 4) |
|---|---|---|---|
| EM-46 | 1 first member | 33.34, 33.33, 33.33 | 14.32, then six at 14.28. Check: 14.32 + 6 x 14.28 = 14.32 + 85.68 = 100.00 |
| EM-47 | 2 first r members | 33.34, 33.33, 33.33 | four at 14.29, then three at 14.28. Check: 4 x 14.29 + 3 x 14.28 = 57.16 + 42.84 = 100.00 |
| EM-48 | 3 largest fractions | all three fractions are equal (one third), so the tie-break decides. With list order as tie-break: 33.34, 33.33, 33.33 | all seven fractions are equal (four sevenths), so the tie-break decides. With list order as tie-break: same as candidate 2 |
| EM-49 | 4 last member | 33.33, 33.33, 33.34 | six at 14.28, then 14.32 |
| EM-50 | 5 account | members 33.33 each, account 0.01. Check: 99.99 + 0.01 = 100.00 | members 14.28 each, account 0.04. Check: 99.96 + 0.04 = 100.00 |
| EM-51 | 6 seeded order | any 1 of 3 members holds the extra cent: 3 possible results. Same seed gives the same one | any 4 of 7 members hold an extra cent: 35 possible results (7 choose 4). Same seed gives the same one |

Candidate 3 only differs from candidate 2 when fractions differ, as in EM-37: shares 3.33 and 6.66 plus 1 cent. The member with weight 2 has the larger fraction (two thirds of a cent against one third), so candidate 3 gives 3.33 and 6.67, and candidate 2 gives 3.34 and 6.66.

## D. Interest

Rule: interest figures match the hand-worked figures to the cent. Rounding steps are explicit.

| Id | Case | Expected | Working | Protects |
|---|---|---|---|---|
| EM-52 | Simple interest, 1,000.00 at 5% a year for 3 years | interest 150.00, total 1,150.00 | 1,000.00 x 0.05 = 50.00 a year. 3 x 50.00 = 150.00. 1,000.00 + 150.00 = 1,150.00 | simple interest has no rounding |
| EM-53 | Simple interest, 2,500.00 at 4.5% a year for 2 years | interest 225.00, total 2,725.00 | 2,500.00 x 0.045 = 112.50 a year. 2 x 112.50 = 225.00. 2,500.00 + 225.00 = 2,725.00 | a rate with one decimal is exact |
| EM-54 | Compound annually, 1,000.00 at 5% for 3 years, rounding only at the end | unrounded 1,157.625 | see table below | compound figures carry every digit until the end |
| EM-55 | Compound per period, 500.00 at 1.5% per period for 4 periods | unrounded 530.6817753125. To 2 places under half-up, half-down and half-even: 530.68 | see table below | rounding once at the end |
| EM-56 | Compound per period, 120.00 at 2% per period for 3 periods | rounding each period gives 127.35. Rounding once at the end gives 127.34. Half-up used for both | see table below | the two approaches differ, so a formula must say which one it uses |

### EM-54: 1,000.00 at 5%, 3 years

| Year | Start | x 1.05 | End |
|---|---|---|---|
| 1 | 1,000.00 | 1,000.00 + 50.00 | 1,050.00 |
| 2 | 1,050.00 | 1,050.00 + 52.50 | 1,102.50 |
| 3 | 1,102.50 | 1,102.50 + 55.125 | 1,157.625 |

The unrounded figure is exactly 1,157.625. It sits exactly at the half of the third place, between 1,157.62 and 1,157.63. Rounded to 2 places:

| Mode | Result | Working |
|---|---|---|
| half-up | 1,157.63 | at the half, away from zero |
| half-down | 1,157.62 | at the half, toward zero |
| half-even | 1,157.62 | 1,157.62 ends in 2, which is even |
| half-+inf | 1,157.63 | at the half, toward positive infinity |
| up | 1,157.63 | any fraction, away from zero |
| down | 1,157.62 | any fraction, toward zero |
| ceiling | 1,157.63 | any fraction, upward |
| floor | 1,157.62 | any fraction, downward |

The four modes that give 1,157.62 and the four that give 1,157.63 split evenly, so this case shows the choice of mode changes a visible figure.

### EM-55: 500.00 at 1.5% per period, 4 periods

Working: each period multiplies by 1.015.

| Period | Start, rounded each period (half-up) | x 1.015 | Rounded | Unrounded running value |
|---|---|---|---|---|
| 1 | 500.00 | 507.50 | 507.50 | 507.5 |
| 2 | 507.50 | 515.1125 | 515.11 | 515.1125 |
| 3 | 515.11 | 522.83665 | 522.84 | 522.8391875 |
| 4 | 522.84 | 530.6826 | 530.68 | 530.6817753125 |

Rounding each period gives 530.68. Rounding once at the end gives 530.68. The cents agree for this case under half-up, half-down and half-even. They differ under the other modes:

| Mode | Each period | Once at the end |
|---|---|---|
| up | 530.70 | 530.69 |
| ceiling | 530.70 | 530.69 |
| down | 530.67 | 530.68 |
| floor | 530.67 | 530.68 |

### EM-56: 120.00 at 2% per period, 3 periods

Found by search over principals from 100.00 to 1,000.00 and rates from 2% to 6%. The search also finds 105.00 and 365.00 at 2%; 120.00 is picked for easy hand working.

| Period | Start | x 1.02 | Rounded each period (half-up) |
|---|---|---|---|
| 1 | 120.00 | 122.4000 | 122.40 |
| 2 | 122.40 | 124.8480 | 124.85 |
| 3 | 124.85 | 127.3470 | 127.35 |

Rounding once at the end: 120.00 x 1.02 x 1.02 x 1.02 = 127.344960, which rounds to 127.34 (below the half of the third place). The difference is 1 cent, from rounding 124.848 up to 124.85 in period 2.

Under half-up, half-down, half-even and half-+inf the pair is the same, 127.35 each period and 127.34 at the end. Under up and ceiling both approaches give 127.35. Under down and floor, each period gives 127.33 and the end gives 127.34.

## E. Decimal and float never mix

Rule: a formula that combines a decimal with a float is refused, not computed. The refusal is one plain sentence that names both kinds and says to convert one explicitly. The exact wording is an open question (6); the tests check the behaviour and that the message is one sentence naming both kinds, not a fixed string.

| Id | Formula | Expected |
|---|---|---|
| EM-60 | decimal amount 12.50 + float literal 0.1 | refused |
| EM-61 | decimal amount 12.50 x float variable rate (value 0.1) | refused |
| EM-62 | compare decimal 0.3 with float 0.3 | refused. The float 0.3 is really 0.299999999999999988898..., so an answer of "equal" or "not equal" would both mislead |
| EM-63 | explicit conversion of float 0.1 to a decimal at 2 places under a half mode | accepted, result 0.10, then 12.50 + 0.10 = 12.60 |
| EM-64 | explicit conversion of float 0.30000000000000004 (the result of 0.1 + 0.2 in float) to 2 places under a half mode | accepted, result 0.30 |
| EM-65 | the same conversion as EM-63 under up and ceiling | 0.11. Float 0.1 is 0.1000000000000000055...; any fraction above 0.10 goes up. Open question 8 decides whether this is the answer |
| EM-66 | conversion with no number of places stated | refused. A conversion states its places |

EM-66 is added beyond the issue's list because the accepted case needs a stated number of places.

## F. Overflow

Rule: amounts are held as whole numbers of the smallest unit on a fast path. Past the boundary B the engine hands over to big decimals. The result is the same on both sides of B. No result may wrap around, saturate or lose digits.

B is to be set by the engine choice. The cases are relative to B.

Expected values come from an independent big-integer calculation, not from the engine. Each case runs twice, with the fast path allowed and with only the big path, and both runs must match the independent value. How to force the big path is a test-hook decision for the engine spike.

| Id | Operation | Inputs, in smallest units | Expected |
|---|---|---|---|
| EM-70 | addition | (B - 2) + 1 | B - 1 |
| EM-71 | addition | (B - 1) + 1 | B |
| EM-72 | addition | B + 1 | B + 1 |
| EM-73 | multiplication | (B - 1) x 1 | B - 1 |
| EM-74 | multiplication | B x 1 | B |
| EM-75 | multiplication | (B + 1) x 1 | B + 1 |
| EM-76 | multiplication | B x 2 | 2B |
| EM-77 | multiplication | B x B | B squared |
| EM-78 | sum of many | 10 amounts: nine of q = floor(B / 10), and a tenth of B - 9q - 1 | B - 1, forward and reverse order |
| EM-79 | sum of many | nine of q, and a tenth of B - 9q | B, forward and reverse order |
| EM-80 | sum of many | nine of q, and a tenth of B - 9q + 1 | B + 1, forward and reverse order |

For EM-78 to EM-80 the running total crosses B only on the tenth addition in EM-80, and the engine hands over mid-sum in that case.

### Illustration only: the 64-bit signed limit

The numbers below use 9,223,372,036,854,775,807 (2 to the power 63, minus 1). It is an illustration. It is not the chosen boundary. At 2 places it is 92,233,720,368,547,758.07 in currency.

| Id | Operation | Expected, in cents | Expected, as currency | Working |
|---|---|---|---|---|
| EM-81 | 9,223,372,036,854,775,807 cents + 1 cent | 9,223,372,036,854,775,808 | 92,233,720,368,547,758.08 | the last digit 7 + 1 = 8; no carry |
| EM-82 | 9,223,372,036,854,775,807 x 2 | 18,446,744,073,709,551,614 | 184,467,440,737,095,516.14 | 2 x 9,223,372,036,854,775,807 |
| EM-83 | sum of 10 amounts of 1,000,000,000,000,000,000 cents | 10,000,000,000,000,000,000 | 100,000,000,000,000,000.00 | 10 x 10^18 = 10^19. After 9 additions the total is 9,000,000,000,000,000,000, still below the limit. The 10th passes it |

A wrapped result of EM-81 would be negative. A saturated result would stay at ...807. Either fails the test.

## G. Extremes

| Id | Case | Expected | Working | Protects |
|---|---|---|---|---|
| EM-90 | 0.00 + 0.00 | 0.00 | 0 + 0 = 0 | zero is a figure like any other |
| EM-91 | 123.45 x 0 | 0.00 | 12,345 cents x 0 = 0 | zero times anything |
| EM-92 | 0.00 among 3 | 0.00 each, remainder 0 | see EM-38 | zero total splits cleanly |
| EM-93 | interest at 0%: 1,000.00 for 3 years, simple | interest 0.00, total 1,000.00 | 1,000.00 x 0 = 0.00 | zero rate adds nothing |
| EM-94 | smallest amount: 0.01 + 0.01 | 0.02 | 1 + 1 = 2 cents | one smallest unit is not lost |
| EM-95 | 0.01 x 0.5 to 2 places | 0.005 before rounding. Rounds as the half row of EM-12, scaled: half-up 0.01, half-down 0.00, half-even 0.00 | 1 cent x 0.5 = 0.5 cent | a half cent is a rounding case, not a loss |
| EM-96 | very large: 999,999,999,999,999,999.99 + 0.01 | 1,000,000,000,000,000,000.00 | 99,999,999,999,999,999,999 cents + 1 = 100,000,000,000,000,000,000 cents. This is above the 64-bit illustration (about 1.8 x 10^19 even unsigned), so it runs on the big path whichever B is chosen | carry through every digit |
| EM-97 | 1 divided by 3 | open question 4. Candidates: refuse; 0.33 to 2 places under half-up; the rational 1/3 | 1/3 has no exact decimal. Under the rational candidate, 1/3 x 3 = 1 exactly. Under the rounding candidate, 0.33 x 3 = 0.99 | a division never returns a quiet wrong figure |
| EM-98 | 2 divided by 3 to 2 places, if the rounding candidate is chosen | half-up 0.67, up 0.67, ceiling 0.67, down 0.66, floor 0.66 | 2/3 = 0.6666..., which is above the half of the third place, so every half mode gives 0.67 | the mode is named in the formula |
| EM-99 | 0.0001 at 4 places, only if places per currency can be 4 | 0.0001 + 0.0001 = 0.0002; 0.0001 x 0.5 = 0.00005 before rounding | 1 unit + 1 unit = 2 units | the smallest unit follows the currency's places |

Only if negative inputs are allowed (open question 5):

| Id | Case | Expected | Working |
|---|---|---|---|
| EM-100 | -10.00 + 4.00 | -6.00 | -1,000 + 400 = -600 cents |
| EM-101 | -0.5 rounded to 0 places, under each mode | the EM-13 row | see B1 |
| EM-102 | -100.00 among 3 | base share -33.33 each, remainder -1 cent, check -99.99 + -0.01 = -100.00 | 10,000 cents / 3 = 3,333 each, cut toward zero; 3 x 3,333 = 9,999; remainder -1. Where the -1 cent goes: open question 2 |

Cut toward floor instead, the base share would be -33.34 with a remainder of +2 cents. That is a different convention and part of open question 2.

## How the figures were checked

Every figure on this page was worked by hand first. Each was then re-checked with Python's decimal module at 100 digits of precision, with the rounding mode named on every rounding step: half-up as ROUND_HALF_UP, half-down as ROUND_HALF_DOWN, half-even as ROUND_HALF_EVEN, up as ROUND_UP, down as ROUND_DOWN, ceiling as ROUND_CEILING and floor as ROUND_FLOOR. Half toward positive infinity has no constant of its own, so it was computed as the floor of the value plus one half, and cross-checked against ROUND_HALF_UP for non-negative inputs and ROUND_HALF_DOWN for negative inputs. The big-number figures were checked with Python's whole-number arithmetic. The float figures were produced with ordinary binary floats to show the drift the library must not have.
