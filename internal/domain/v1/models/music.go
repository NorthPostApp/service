package models

type Music struct {
	Filename     string  `json:"filename" firestore:"filename"`
	Title        string  `json:"title" firestore:"title"`
	Genre        string  `json:"genre" firestore:"genre"`
	Size         float64 `json:"size" firestore:"size"`
	LastModified int64   `json:"lastModified" firestore:"lastModified"`
	DurationSec  int64   `json:"durationSec" firestore:"durationSec"`
}
