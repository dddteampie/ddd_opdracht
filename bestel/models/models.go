package models

type Product struct {
}

type Order struct {
	//TODO: fix id GUID
	ID        uint
	ProductID uint //TODO: fix Gorm connecties

}
