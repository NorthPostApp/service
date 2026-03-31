package models

type AdminUser struct {
	Email       string `json:"email" firestore:"email"`
	DisplayName string `json:"displayName" firestore:"displayName"`
	CreatedAt   int64  `json:"createdAt" firestore:"createdAt"`
	LastLogin   int64  `json:"lastLogin" firestore:"lastLogin"`
	ImageUrl    string `json:"imageUrl,omitempty" firestore:"imageUrl"`
}

// type AppUser struct {
// 	Email       string   `json:"email" firestore:"email"`
// 	DisplayName string   `json:"displayName" firestore:"displayName"`
// 	CreatedAt   int64    `json:"createdAt" firestore:"createdAt"`
// 	LastLogin   int64    `json:"lastLogin" firestore:"lastLogin"`
// 	ImageUrl    string   `json:"imageUrl,omitempty" firestore:"imageUrl"`
// 	LikedMusics []string `json:"likedMusics" firestore:"likedMusics"`
// 	Drafts      []string `json:"drafts" firestore:"drafts"`
// }
