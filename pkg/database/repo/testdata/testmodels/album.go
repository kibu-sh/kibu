package testmodels

type Album struct {
	AlbumID  int     `db:"AlbumId,pk,table=albums"`
	ArtistID int     `db:"ArtistId"`
	Title    string  `db:"Title"`
	Artist   *Artist `db:"-"`
	Omitted  string  `db:"-"`
}
