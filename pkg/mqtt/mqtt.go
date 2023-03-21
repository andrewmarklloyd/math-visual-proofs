package mqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type fn func(string)

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

func (c MqttClient) Subscribe(topic string, subscribeHandler fn) error {
	if token := c.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		subscribeHandler(string(msg.Payload()))
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
