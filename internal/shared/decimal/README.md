# Decimal for Money

Elite4Print handles prices, taxes, discounts, shipping costs, and refunds. Using
`float64` for money is dangerous because binary floating point cannot represent
decimal values like `0.10` exactly, leading to rounding errors.

## What `shopspring/decimal` does

`shopspring/decimal` provides an arbitrary-precision decimal type backed by
`math/big`. It stores values as integers with a scale, so:

```go
a := decimal.NewFromFloat(0.1)
b := decimal.NewFromFloat(0.2)
fmt.Println(a.Add(b)) // 0.3 exactly
```

## Where to use it

- Product prices and pricing tiers
- Cart totals, taxes, discounts
- Order totals, payments, refunds
- Shipping rates
- Coupons and loyalty points

## Database storage

Store decimals as `NUMERIC(19, 4)` in PostgreSQL. `sqlx` can scan into
`decimal.Decimal` directly.

## Example domain value object

```go
package domain

import "github.com/shopspring/decimal"

type Money struct {
    Amount   decimal.Decimal
    Currency string
}

func NewMoney(amount decimal.Decimal, currency string) Money {
    return Money{Amount: amount.Truncate(2), Currency: currency}
}
```

## Why not float64?

The current Django backend uses `FloatField` for some money fields, which is one
of the known financial-precision bugs. The Go rewrite fixes this by making
`decimal.Decimal` the only money type.
