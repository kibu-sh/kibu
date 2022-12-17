package testmodels

type Genre struct {
	GenreId int `db:"GenreId,pk,table=genres"`
	Name    string
}
