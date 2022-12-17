package testmodels

type Artist struct {
	ArtistID int      `db:"ArtistId,pk,table=artists"`
	Name     string   `db:"Name"`
	Albums   []*Album `db:"-"`
}
