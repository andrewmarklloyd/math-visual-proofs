package mqtt

const (
	// prefix represents the subscriber. example: math-visual-proofs-server is
	// for messages intended to be sent to the server. math-visual-proofs-agent
	// is for messages intended to be sent to the agent. this is important for
	// ACL configured in mqtt
	RenderStartTopic   = "math-visual-proofs-server/render/start"
	RenderAckTopic     = "math-visual-proofs-agent/render/ack"
	RenderErrTopic     = "math-visual-proofs-agent/render/error"
	RenderSuccessTopic = "math-visual-proofs-agent/render/success"
)

type RenderMessage struct {
	FileNames []string `json:"fileNames"`
	RepoURL   string   `json:"repoURL"`
}
