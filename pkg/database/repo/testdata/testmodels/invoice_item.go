package testmodels

type InvoiceItem struct {
	InvoiceItemId int `db:"InvoiceItemId,pk,table=invoice_items"`
	InvoiceId     int `db:"InvoiceId"`
	TrackId       int `db:"TrackId"`
	UnitPrice     int `db:"UnitPrice"`
	Quantity      int `db:"Quantity"`
}
