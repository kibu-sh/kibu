package database

type Driver struct {
	Entity EntityDefinition
}

type DriverOption func(q *Driver) error
