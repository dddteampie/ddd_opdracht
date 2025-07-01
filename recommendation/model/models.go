package model

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Product struct {
	EAN          int64  `json:"ean"`
	Naam         string `json:"naam"`
	Omschrijving string `json:"omschrijving"`
}

type Aanbeveling struct {
	gorm.Model
	ClientID              string `gorm:"uniqueIndex"`
	Versie                int
	AanmaakDatum          time.Time
	PassendeCategorieënID *uint                     `gorm:"index"`
	PassendeCategorieën   *PassendeCategorieënLijst `gorm:"foreignKey:PassendeCategorieënID"`
	OplossingenLijstID    *uint                     `gorm:"index"`
	OplossingenLijst      *OplossingenLijst         `gorm:"foreignKey:OplossingenLijstID"`
}

type PassendeCategorieënLijst struct {
	gorm.Model
	CategoryIDs pq.Int64Array `gorm:"type:integer[]" json:"-"`
	Categories  []Category    `gorm:"-" json:"categories"`
}

type OplossingenLijst struct {
	gorm.Model
	ProductEANs pq.Int64Array `gorm:"type:bigint[]" json:"-"`
	Products    []Product     `gorm:"-" json:"products"`
}

type Category struct {
	ID   int    `json:"id"`
	Naam string `json:"naam"`
}

type Tag struct {
	ID   int    `json:"id"`
	Naam string `json:"naam"`
}

func ConvertIntSliceToPQInt64Array(slice []int) pq.Int64Array {
	arr := make(pq.Int64Array, len(slice))
	for i, v := range slice {
		arr[i] = int64(v)
	}
	return arr
}

func ConvertPQInt64ArrayToIntSlice(arr pq.Int64Array) []int {
	slice := make([]int, len(arr))
	for i, v := range arr {
		slice[i] = int(v)
	}
	return slice
}

func ConvertInt64SliceToPQInt64Array(slice []int64) pq.Int64Array {
	arr := make(pq.Int64Array, len(slice))
	copy(arr, slice)
	return arr
}
