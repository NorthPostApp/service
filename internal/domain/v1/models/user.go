package models

type AdminUser struct {
	Email       string `json:"email" firestore:"email"`
	DisplayName string `json:"displayName" firestore:"displayName"`
	CreatedAt   int64  `json:"createdAt" firestore:"createdAt"`
	LastLogin   int64  `json:"lastLogin" firestore:"lastLogin"`
	ImageUrl    string `json:"imageUrl,omitempty" firestore:"imageUrl"`
}

type AppUser struct {
	Email       string `json:"email" firestore:"email"`
	DisplayName string `json:"displayName" firestore:"displayName"`
	CreatedAt   int64  `json:"createdAt" firestore:"createdAt"`
	LastLogin   int64  `json:"lastLogin" firestore:"lastLogin"`
	ImageUrl    string `json:"imageUrl,omitempty" firestore:"imageUrl"`
}

type SavedAddresses struct {
	IDs []string `json:"ids" firestore:"ids"`
}

type UserAddressRequests struct {
	IDs                []string `json:"ids" firestore:"ids"`
	ActiveRequestCount int64    `json:"activeRequestCount" firestore:"activeRequestCount"`
}

type AddressBook struct {
	SavedAddresses     map[Language][]string `json:"savedAddresses" firestore:"savedAddresses"`
	Requests           map[Language][]string `json:"requests" firestore:"requests"`
	ActiveRequestCount map[Language]int64    `json:"activeRequestCount" firestore:"activeRequestCount"`
}
