package plan

type BillingInterval string

const (
	BillingIntervalMonth BillingInterval = "month"
	BillingIntervalYear  BillingInterval = "year"
)

const FreePlanSlug = "free"

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)
