package publisher

// Publisher gather all the necessary fields to publish a message
type Publisher struct {
	// Project is the GCP project ID
	Project string
	// DestinationTopic is the GCP Pub/Sub topic name where the message will be published
	DestinationTopic string
	// SourceTopic is the GCP Pub/Sub topic name where the message will be read
	SourceTopic string
}

// NewPublisher creates a new publisher
func NewPublisher(project, destinationTopic, sourceTopic string) *Publisher {
	return &Publisher{
		Project:          project,
		DestinationTopic: destinationTopic,
		SourceTopic:      sourceTopic,
	}
}
