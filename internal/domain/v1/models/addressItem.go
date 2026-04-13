package models

import "github.com/typesense/typesense-go/v4/typesense/api"

type AddressGenerationSchema struct {
	Name       string   `json:"name"`
	BriefIntro string   `json:"briefIntro"`
	Tags       []string `json:"tags"`
	Address    Address  `json:"address"`
}

type BatchAddressGenerationSchema struct {
	Addresses []AddressGenerationSchema
}

type AddressItem struct {
	ID         string   `json:"id" firestore:"id"`
	Name       string   `json:"name" firestore:"name"`
	BriefIntro string   `json:"briefIntro" firestore:"briefIntro"`
	CreatedAt  int64    `json:"createdAt" firestore:"createdAt"`
	UpdatedAt  int64    `json:"updatedAt" firestore:"updatedAt"`
	Tags       []string `json:"tags" firestore:"tags"`
	Address    Address  `json:"address" firestore:"address"`
}

type Address struct {
	City         string `json:"city" firestore:"city"`
	Country      string `json:"country" firestore:"country"`
	Line1        string `json:"line1" firestore:"line1"`
	Line2        string `json:"line2,omitempty" firestore:"line2"`
	BuildingName string `json:"buildingName,omitempty" firestore:"buildingName"`
	PostalCode   string `json:"postalCode,omitempty" firestore:"postalCode"`
	Region       string `json:"region" firestore:"region"`
}

type TagsRecord struct {
	Tags        map[string][]string `json:"tags" firestore:"tags"`
	RefreshedAt int64               `json:"refreshedAt" firestore:"refreshedAt"`
}

// The typesense collection schema should be the same as this struct
type TypesenseAddressRecord struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	BriefIntro string   `json:"briefIntro"`
	Tags       []string `json:"tags"`
}

func GetTypesenseAddressCollectionSchema(name string) *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: name,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "briefIntro", Type: "string"},
			{Name: "tags", Type: "string[]"},
		},
	}
}
