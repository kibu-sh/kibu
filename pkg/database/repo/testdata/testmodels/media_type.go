package testmodels

type MediaType struct {
	MediaTypeId int `db:"MediaTypeId,pk,table=media_types"`
	Name        string
}
