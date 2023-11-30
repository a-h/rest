package models

type Topic struct {
	// Namespace of the topic
	// Example: home
	Namespace string `json:"namespace"`
	// Topic of the topic
	// Example: 'cushions'
	Topic string `json:"topic"`
	// Private is true if the topic is private.
	Private bool `json:"private"`
	// ViewCount is the number of times the topic has been viewed.
	// Example: "1234"
	ViewCount int64 `json:"viewCount"`
	// Keywords is a list of keywords for the topic. This is optional.
	// Example: ["Growing pains", "Tech"]
	Keywords []string `json:"keywords,omitempty"`
}

// TopicsPostRequest is the request to POST /topics.
type TopicsPostRequest struct {
	Topic
}

type TopicsPostResponse struct {
	ID string `json:"id"`
}

// TopicsGetResponse is the response to GET /topics.
type TopicsGetResponse struct {
	Topics []TopicRecord `json:"topics"`
}

type TopicRecord struct {
	// ID of the topic record.
	// Example: 1234
	ID string `json:"id"`
	Topic
}
