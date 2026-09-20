package models

type ContentWithTags struct {
	Content Content `json:"content"`
	Tags    []Tag   `json:"tags"`
}
