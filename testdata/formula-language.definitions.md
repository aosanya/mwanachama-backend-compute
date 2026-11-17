# Definitions of the Validate rows

Each row of the first table is one input or one check of a definition. A row with the part `from` copies another definition, and the rows after it with the same part and name replace the copied ones. The second table says which definition each Validate case is checked against, and whether the case validates a row or the definition itself.

| id | part | name | detail | message |
|---|---|---|---|---|
| V1 | input | balance | decimal: required; no-negatives; decimal-places 2 | |
| V1 | input | payment | decimal: required; no-negatives; decimal-places 2 | |
| V1 | check | 1 | payment > 0 | The payment must be more than zero |
| V1 | check | 2 | payment <= balance + 1000.00 | The payment cannot be more than 1000.00 above the balance |
| V2 | input | status | text: required; allowed-values "active", "lapsed", "suspended" | |
| V2 | input | requested | decimal: required; range more than 0.00; decimal-places 2 | |
| V2 | input | contributions | list of decimal: min-items 1; no-negatives | |
| V3 | input | deposits | list of decimal: min-items 1; no-negatives | |
| V3 | input | expenses | list of decimal: no-negatives | |
| V3 | input | members | integer: required; range 1 to 500 | |
| V4 | input | kilos | integer: required; no-negatives | |
| V4 | input | price_per_kilo | decimal: required; range more than 0.000; decimal-places 3 | |
| V4 | input | credit | decimal: required; no-negatives; decimal-places 2 | |
| V5 | input | due | date: required | |
| V5 | input | filed | date: required | |
| V9 | input | principal | decimal: required; range more than 0.00; decimal-places 2 | |
| V9 | input | monthly_rate | decimal: required; range 0.000 to 0.050; decimal-places 3 | |
| V9 | input | months | integer: required; range 1 to 60 | |
| V9 | check | 1 | principal >= 100.00 | The loan must be at least 100.00 |
| V9 | check | 2 | principal * monthly_rate * months <= principal | The interest cannot be more than the amount borrowed |
| V10 | input | scores | list of integer: min-items 1; max-items 12; range 0 to 100 | |
| V16 | input | balances | list of decimal: min-items 1; decimal-places 2 | |
| V16 | input | floor | decimal: required; no-negatives; decimal-places 2 | |
| V17 | input | yields | list of decimal: min-items 1; no-negatives; decimal-places 2 | |
| V18 | input | joined | date: required | |
| V18 | input | on | date: required | |
| V18 | input | paid_up | boolean: required | |
| V18 | input | suspended | boolean: required | |
| V18 | input | role | text: required; allowed-values "member", "officer", "guest" | |
| V19 | input | kilos | integer: required; no-negatives | |
| V19 | input | bag_size | integer: required; range 1 to 100 | |
| VB | input | total | decimal: required; no-negatives; decimal-places 2 | |
| VB | input | people | integer: required | |
| VB | check | 1 | total / people >= 1.00 | Each person must get at least 1.00 |
| V19-range-reversed | from | V19 | | |
| V19-range-reversed | input | bag_size | integer: required; range 100 to 1 | |
| V1-places-too-many | from | V1 | | |
| V1-places-too-many | input | payment | decimal: required; no-negatives; decimal-places 11 | |
| V2-allowed-values-empty | from | V2 | | |
| V2-allowed-values-empty | input | status | text: required; allowed-values | |
| V1-check-unknown-name | from | V1 | | |
| V1-check-unknown-name | check | 1 | payment <= limit | The payment must be more than zero |
| V1-check-not-boolean | from | V1 | | |
| V1-check-not-boolean | check | 1 | payment + 1 | The payment must be more than zero |
| V19-input-needs-required | from | V19 | | |
| V19-input-needs-required | input | bag_size | integer: range 1 to 100 | |

| case | definition | target |
|---|---|---|
| rules-required-valid | V1 | row |
| rules-required-missing | V1 | row |
| rules-required-and-bad-together | V1 | row |
| rules-no-negatives-valid | V1 | row |
| rules-no-negatives-invalid | V1 | row |
| rules-decimal-places-valid | V1 | row |
| rules-decimal-places-invalid | V1 | row |
| rules-three-problems-at-once | V1 | row |
| rules-check-failed | V1 | row |
| rules-check-failed-second | V1 | row |
| rules-decimal-places-boundary-valid | V4 | row |
| rules-decimal-places-trailing-zero | V4 | row |
| rules-different-codes-at-once | V4 | row |
| rules-type-valid | V19 | row |
| rules-type-invalid | V19 | row |
| rules-type-other-inputs-still-tested | V19 | row |
| rules-unknown-input | V19 | row |
| rules-type-date-valid | V5 | row |
| rules-type-date-invalid | V5 | row |
| rules-required-date-missing | V5 | row |
| rules-required-boolean-missing | V18 | row |
| rules-allowed-values-valid | V2 | row |
| rules-allowed-values-invalid | V2 | row |
| rules-allowed-values-case-differs | V2 | row |
| rules-range-exclusive-valid | V2 | row |
| rules-range-exclusive-invalid | V2 | row |
| rules-no-negatives-list-item-valid | V2 | row |
| rules-no-negatives-list-item-invalid | V2 | row |
| rules-list-at-least-one-not-given | V2 | row |
| rules-list-at-least-one-valid | V17 | row |
| rules-list-at-least-one-invalid | V17 | row |
| rules-decimal-places-list-item-invalid | V17 | row |
| rules-range-valid | V3 | row |
| rules-range-edge-valid | V3 | row |
| rules-range-too-small | V3 | row |
| rules-range-too-large | V3 | row |
| rules-list-not-given-is-empty | V3 | row |
| rules-each-item-valid | V10 | row |
| rules-each-item-too-large | V10 | row |
| rules-each-item-three-problems | V10 | row |
| rules-list-at-most-valid | V10 | row |
| rules-list-too-long | V10 | row |
| rules-type-list-item-valid | V16 | row |
| rules-type-list-item-invalid | V16 | row |
| rules-check-valid | V9 | row |
| rules-check-two-failed | V9 | row |
| rules-check-skipped-input-problem | V9 | row |
| rules-check-skipped-all | V9 | row |
| rules-check-error | VB | row |
| rules-check-division-valid | VB | row |
| rules-definition-range-reversed | V19-range-reversed | definition |
| rules-definition-places-too-many | V1-places-too-many | definition |
| rules-definition-allowed-values-empty | V2-allowed-values-empty | definition |
| rules-definition-check-unknown-name | V1-check-unknown-name | definition |
| rules-definition-check-not-boolean | V1-check-not-boolean | definition |
| rules-definition-input-needs-required | V19-input-needs-required | definition |
