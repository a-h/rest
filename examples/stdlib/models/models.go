package models

type Topic struct {
	Namespace string `json:"namespace"`
	Topic     string `json:"topic"`
	Private   bool   `json:"private"`
	ViewCount int64  `json:"viewCount"`
}
