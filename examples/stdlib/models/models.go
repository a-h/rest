package models

// Topic of a thread.
type Topic struct {
	Namespace string   `json:"namespace" validate:"required"`
	Topic     string   `json:"topic"  validate:"omitempty,oneof=general politics health"`
	Private   bool     `json:"private"`
	ViewCount int64    `json:"viewCount"`
	Roles     []string `json:"roles" validate:"gt=0,dive,oneof=admin user staff"`
	CreatedAt string   `json:"createdAt" validate:"datetime=2006-01-02T15:04:05Z07:00"`
}
