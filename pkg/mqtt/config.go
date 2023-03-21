package mqtt

const (
	// prefix represents the subscriber. example: math-visual-proofs-server is
	// for messages intended to be sent to the server. math-visual-proofs-agent
	// is for messages intended to be sent to the agent. this is important for
	// ACL configured in mqtt
	RenderStartTopic = "math-visual-proofs-server/render/start"
	RenderAckTopic   = "math-visual-proofs-agent/render/ack"
)

type RenderMessage struct {
	FileName  string `json:"fileName"`
	ClassName string `json:"className"`
	RepoURL   string `json:"repoURL"`
}
