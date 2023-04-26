package mqtt

import (
	"encoding/json"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type fn func(RenderMessage)

type MqttClient struct {
	client mqtt.Client
}

func NewMQTTClient(addr, clientID string, connectHandler func(client mqtt.Client), connectionLostHandler func(client mqtt.Client, err error)) MqttClient {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(addr)
	opts.SetClientID(clientID)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectionLostHandler
	opts.CleanSession = true
	opts.AutoReconnect = true
	client := mqtt.NewClient(opts)

	return MqttClient{
		client,
	}
}

func (c MqttClient) Connect() error {
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c MqttClient) Cleanup() {
	c.client.Disconnect(250)
}

func (c MqttClient) Subscribe(topic string, logger *zap.SugaredLogger, subscribeHandler fn) error {
	if token := c.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		renderMessage := RenderMessage{}
		err := json.Unmarshal(msg.Payload(), &renderMessage)
		if err != nil {
			pubErr := c.PublishRenderFeedbackMessage(RenderErrTopic, RenderFeedbackMessage{
				Status:    StatusSucceess,
				RepoURL:   UnknownRepoURL,
				GithubSHA: UnknownGithubSHA,
				Message:   fmt.Sprintf("error during render: %s", err.Error()),
			})
			if pubErr != nil {
				logger.Errorf("error publishing to renderErrTopic: %s", pubErr)
			}
			return
		}

		logger.Info("received request to render: ", renderMessage)

		if os.Getenv("MOCK_MODE") != "" {
			return
		}

		err = c.PublishRenderFeedbackMessage(RenderAckTopic, RenderFeedbackMessage{
			Status:    StatusSucceess,
			RepoURL:   renderMessage.RepoURL,
			GithubSHA: renderMessage.GithubSHA,
			Message:   "successfully cloned repo and started render",
		})
		if err != nil {
			logger.Errorf("error publishing to renderErrTopic: %s", err)
		}
		// wait group here?
		subscribeHandler(renderMessage)
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// TODO: solve for retained and qos to ensure durability
func (c MqttClient) Publish(topic, message string) error {
	token := c.client.Publish(topic, 1, false, message)
	token.Wait()
	return token.Error()
}

func (c MqttClient) PublishRenderFeedbackMessage(topic string, message RenderFeedbackMessage) error {
	m, err := json.Marshal(message)
	if err != nil {
		return err
	}

	token := c.client.Publish(topic, 1, false, string(m))
	token.Wait()
	return token.Error()
}
