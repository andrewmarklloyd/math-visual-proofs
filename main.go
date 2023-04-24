package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	mqttC "github.com/eclipse/paho.mqtt.golang"

	"github.com/andrewmarklloyd/math-visual-proofs/internal/pkg/aws"
	"github.com/andrewmarklloyd/math-visual-proofs/internal/pkg/git"
	"github.com/andrewmarklloyd/math-visual-proofs/pkg/mqtt"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

var messageClient mqtt.MqttClient

var awsClient aws.Client

const (
	clientID  = "math-visual-proofs-server"
	clonePath = "/tmp/working"
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

	mqttAddr := os.Getenv("CLOUDMQTT_SERVER_URL")

	if mqttAddr == "" {
		logger.Fatal("CLOUDMQTT_SERVER_URL env var must be set")
	}

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

		logger.Info("received request to render: ", renderMessage)

		if os.Getenv("MOCK_MODE") != "" {
			return
		}

		err = subscribeHandler(renderMessage)
		if err != nil {
			handleError(err)
			return
		}

		logger.Info("successfully rendered and uploaded: ", renderMessage)

		err = messageClient.Publish(mqtt.RenderSuccessTopic, "successfully finished render, video is uploaded to storage")
		if err != nil {
			logger.Errorf("publishing success message: %w", err)
		}
	})

	for {
		time.Sleep(time.Hour)
	}
}

func subscribeHandler(renderMessage mqtt.RenderMessage) error {
	defer os.RemoveAll(clonePath)

	err := os.RemoveAll(clonePath)
	if err != nil {
		return fmt.Errorf("removing existing cloned repository: %s", err.Error())
	}

	err = git.Clone(renderMessage.RepoURL, clonePath)
	if err != nil {
		return fmt.Errorf("cloning repository %s: %s", renderMessage.RepoURL, err.Error())
	}

	err = messageClient.Publish(mqtt.RenderAckTopic, "successfully cloned repo and started render")
	if err != nil {
		return fmt.Errorf("publishing ack message: %w", err)
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s", clonePath, renderMessage.FileNames)); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %s not found, cannot render", renderMessage.FileNames)
	}

	err = render(renderMessage)
	if err != nil {
		return fmt.Errorf("error rendering: %s", err.Error())
	}

	path := fmt.Sprintf("%s/media/videos/%s/720p30/%s.mp4", clonePath, "", "")
	err = awsClient.UploadFile(context.Background(), path, fmt.Sprintf("%s.mp4", ""))
	if err != nil {
		return fmt.Errorf("error uploading to s3: %w", err)
	}

	return nil
}

func render(renderMessage mqtt.RenderMessage) error {
	c := fmt.Sprintf(`docker run --rm --user="$(id -u):$(id -g)" -v "%s":/manim manimcommunity/manim:stable manim %s -qm --progress_bar none`, clonePath, "")
	cmd := exec.Command("bash", "-c", c)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(string(out))
		return fmt.Errorf("%w, %s", err, string(out))
	}

	return nil
}

func handleError(err error) {
	logger.Error(err)
	pubErr := messageClient.Publish(mqtt.RenderErrTopic, fmt.Sprintf("error during render: %s", err.Error()))
	if pubErr != nil {
		logger.Errorf("error publishing to renderErrTopic: %s", pubErr)
	}
}
