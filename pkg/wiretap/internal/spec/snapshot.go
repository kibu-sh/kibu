package spec

import (
	"github.com/discernhq/devx/pkg/messaging"
	"github.com/discernhq/devx/pkg/messaging/multichannel"
	"time"
)

type SnapshotMessageTopic messaging.Topic[Snapshot]

// Snapshot a heavy object that contains a reconstructed http request and response pair
type Snapshot struct {
	ID       string
	Secure   bool
	Duration time.Duration
	Request  *SerializableRequest
	Response *SerializableResponse
}

// SnapshotRef a lightweight reference to a snapshot
type SnapshotRef struct {
	ID string `json:"id"`
}

func NewSnapshotMessageTopic() SnapshotMessageTopic {
	return multichannel.NewTopicWithDefaults[Snapshot]()
}

func NewSnapshot(req RequestCloner, res ResponseCloner, elapsed time.Duration) (*Snapshot, error) {
	sReq, err := SerializeRequest(req)
	if err != nil {
		return nil, err
	}

	sRes, err := SerializeResponse(res)
	if err != nil {
		return nil, err
	}

	reqC := req.Clone()

	return &Snapshot{
		ID:       defaultSnapshotIDFunc(reqC.Method, reqC.URL),
		Request:  sReq,
		Response: sRes,
		Duration: elapsed,
		// TODO: lock this in with a test
		Secure: reqC.URL.Scheme == "https",
	}, nil
}

// Ref generates a lightweight reference to a snapshot
func (s Snapshot) Ref() *SnapshotRef {
	return NewSnapshotRef(s.ID)
}

func NewSnapshotRef(s string) *SnapshotRef {
	return &SnapshotRef{ID: s}
}
