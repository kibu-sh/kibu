package testmodels

type Track struct {
	TrackId      int        `db:"TrackId,pk,table=tracks"`
	AlbumId      int        `db:"AlbumId"`
	MediaTypeId  int        `db:"MediaTypeId"`
	GenreId      int        `db:"GenreId"`
	Composer     string     `db:"Composer"`
	Milliseconds int        `db:"Milliseconds"`
	Bytes        int        `db:"Bytes"`
	UnitPrice    int        `db:"UnitPrice"`
	Name         string     `db:"Name"`
	MediaType    *MediaType `db:"-"`
	Genre        *Genre     `db:"-"`
	Album        *Album     `db:"-"`
}
