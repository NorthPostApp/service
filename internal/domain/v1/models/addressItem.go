package models

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
	Line2        string `json:"line2,omitempty" firestore:"line2,omitempty"`
	BuildingName string `json:"buildingName,omitempty" firestore:"buildingName,omitempty"`
	PostalCode   string `json:"postalCode" firestore:"postalCode"`
	Region       string `json:"region" firestore:"region"`
}
