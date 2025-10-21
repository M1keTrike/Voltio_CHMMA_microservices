// Message represents a message structure with a content field.
// The Content field holds the textual content of the message and is serialized as "content" in JSON.
package models

type Message struct {
	Content string `json:"content"`
}
