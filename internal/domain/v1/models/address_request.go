package models

type AddressRequestStatus string

const (
	RequestStatusPending    AddressRequestStatus = "pending"
	RequestStatusProcessing AddressRequestStatus = "processing"
	RequestStatusCompleted  AddressRequestStatus = "completed"
	RequestStatusFailed     AddressRequestStatus = "failed"
)

func (s AddressRequestStatus) IsValid() bool {
	switch s {
	case RequestStatusPending,
		RequestStatusProcessing,
		RequestStatusCompleted,
		RequestStatusFailed:
		return true
	default:
		return false
	}
}

type AddressRequest struct {
	ID        string               `json:"id" firestore:"id"`
	Content   string               `json:"content" firestore:"content"`
	RequestBy string               `json:"requestBy" firestore:"requestBy"`
	Status    AddressRequestStatus `json:"status" firestore:"status"`
	CreatedAt int64                `json:"createdAt" firestore:"createdAt"`
	UpdatedAt int64                `json:"updatedAt" firestore:"updatedAt"`

	PendingCandidates []AddressItem `json:"pendingCandidates,omitempty" firestore:"pendingCandidates"`
	FailedReason      string        `json:"failedReason,omitempty" firestore:"failedReason"`
	ResolvedID        string        `json:"resolvedID,omitempty" firestore:"resolvedID"`
	Notes             string        `json:"notes,omitempty" firestore:"notes"`
}
