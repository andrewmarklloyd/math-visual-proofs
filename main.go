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
	"github.com/andrewmarklloyd/math-visual-proofs/internal/pkg/mqtt"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

var messageClient mqtt.MqttClient

var awsClient aws.Client

var processing bool

type RenderMessage struct {
	FileName  string `json:"fileName"`
	ClassName string `json:"className"`
}

func main() {
	processing = false

	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Error creating logger:", err)
	}
	logger = l.Sugar().Named("math-visual-proofs")
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

	messageClient = mqtt.NewMQTTClient(mqttAddr, func(client mqttC.Client) {
		logger.Info("Connected to MQTT server")
	}, func(client mqttC.Client, err error) {
		logger.Fatalf("Connection to MQTT server lost: %s", err)
	})
	err = messageClient.Connect()
	if err != nil {
		logger.Fatalf("connecting to mqtt: %s", err)
	}

	messageClient.Subscribe("math-visual-proofs/render/start", func(message string) {
		logger.Infof("received message: %s", message)

		if processing {
			// refuse message somehow?
			logger.Info("cannot render, already processing another request")
			return
		}

		renderMessage := RenderMessage{}
		err := json.Unmarshal([]byte(message), &renderMessage)
		if err != nil {
			logger.Errorf("unmarshalling render message: %w", err)
			return
		}

		processing = true
		logger.Infof("rendering %s", renderMessage.FileName)
		err = render(renderMessage)
		if err != nil {
			processing = false
			logger.Errorf("error rendering: %s", err.Error())
			return
		}

		logger.Infof("uploading %s.mp4 to s3", renderMessage.ClassName)
		path := fmt.Sprintf("/root/media/videos/%s/720p30/%s.mp4", renderMessage.ClassName, renderMessage.ClassName)
		err = awsClient.UploadFile(context.Background(), path, fmt.Sprintf("%s.mp4", renderMessage.ClassName))
		if err != nil {
			processing = false
			logger.Errorf("error uploading to s3: ", err.Error())
			return
		}

		processing = false
		logger.Infof("successfully uploaded %s to s3", renderMessage.ClassName)
	})

	for {
		time.Sleep(time.Hour)
	}
}

func render(renderMessage RenderMessage) error {
	c := fmt.Sprintf(`docker run --rm --user="$(id -u):$(id -g)" -v "$(pwd)":/manim manimcommunity/manim:stable manim %s -qm`, renderMessage.FileName)
	cmd := exec.Command("bash", "-c", c)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}

	return nil
}
