package models

type AddressItem struct {
	ID         string   `json:"id" firestore:"id"`
	Name       string   `json:"name" firestore:"name"`
	BriefIntro string   `json:"briefIntro" firestore:"briefIntro"`
	CreatedAt  int64    `json:"createdAt" firestore:"createdAt"`
	UpdatedAt  int64    `json:"updatedAt" firestore:"updatedAt"`
	Tags       []string `json:"tags" firestore:"tags"`
}
