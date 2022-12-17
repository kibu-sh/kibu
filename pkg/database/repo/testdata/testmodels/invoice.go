package testmodels

import "time"

type Invoice struct {
	InvoiceId      int            `db:"InvoiceId,pk,table=invoices"`
	CustomerId     int            `db:"CustomerId"`
	InvoiceDate    time.Time      `db:"InvoiceDate"`
	BillingAddress string         `db:"BillingAddress"`
	Items          []*InvoiceItem `db:"-"`
}
