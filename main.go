package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	mqttC "github.com/eclipse/paho.mqtt.golang"

	"github.com/andrewmarklloyd/math-visual-proofs/internal/pkg/aws"
	"github.com/andrewmarklloyd/math-visual-proofs/pkg/mqtt"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

var messageClient mqtt.MqttClient

var awsClient aws.Client

const (
	clientID = "math-visual-proofs-server"
)

func main() {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Error creating logger:", err)
	}
	logger = l.Sugar().Named("math-visual-proofs-server")
	defer logger.Sync()

	awsClient, err = aws.NewClient()
	if err != nil {
		logger.Fatal("error creating aws client: %s", err.Error())
	}

	user := os.Getenv("CLOUDMQTT_MATH_PROOFS_SERVER_USER")
	pw := os.Getenv("CLOUDMQTT_MATH_PROOFS_SERVER_PASSWORD")
	url := os.Getenv("CLOUDMQTT_URL")

	if user == "" || pw == "" || url == "" {
		logger.Fatal("CLOUDMQTT_MATH_PROOFS_SERVER_USER CLOUDMQTT_MATH_PROOFS_SERVER_PASSWORD CLOUDMQTT_URL env vars must be set")
	}

	mqttAddr := fmt.Sprintf("mqtt://%s:%s@%s", user, pw, strings.Split(url, "@")[1])

	messageClient = mqtt.NewMQTTClient(mqttAddr, clientID, func(client mqttC.Client) {
		logger.Info("Connected to MQTT server")
	}, func(client mqttC.Client, err error) {
		logger.Errorf("Connection to MQTT server lost: %s", err)
	})

	err = messageClient.Connect()
	if err != nil {
		logger.Fatalf("connecting to mqtt: %s", err)
	}

	messageClient.Subscribe(mqtt.RenderStartTopic, func(message string) {
		renderMessage := mqtt.RenderMessage{}
		err := json.Unmarshal([]byte(message), &renderMessage)
		if err != nil {
			handleError(fmt.Errorf("unmarshalling render message: %w", err))
			return
		}

		logger.Infof("received request to render: ", renderMessage)

		if os.Getenv("MOCK_MODE") != "" {
			return
		}

		err = subscribeHandler(renderMessage)
		if err != nil {
			handleError(err)
		}
	})

	for {
		time.Sleep(time.Hour)
	}
}

func subscribeHandler(renderMessage mqtt.RenderMessage) error {
	err := render(renderMessage)
	if err != nil {
		return fmt.Errorf("error rendering: %s", err.Error())
	}

	path := fmt.Sprintf("/root/media/videos/%s/720p30/%s.mp4", renderMessage.ClassName, renderMessage.ClassName)
	err = awsClient.UploadFile(context.Background(), path, fmt.Sprintf("%s.mp4", renderMessage.ClassName))
	if err != nil {
		return fmt.Errorf("error uploading to s3: %w", err)
	}

	return nil
}

func render(renderMessage mqtt.RenderMessage) error {
	c := fmt.Sprintf(`docker run --rm --user="$(id -u):$(id -g)" -v "$(pwd)":/manim manimcommunity/manim:stable manim %s -qm`, renderMessage.FileName)
	cmd := exec.Command("bash", "-c", c)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}

	return nil
}

func handleError(err error) {
	logger.Error(err)
	// todo: provide user feedback
}
