package mqtt

const (
	RenderStartTopic = "math-visual-proofs/render/start"
)

type RenderMessage struct {
	FileName  string `json:"fileName"`
	ClassName string `json:"className"`
}
