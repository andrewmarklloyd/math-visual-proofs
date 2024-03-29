package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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
	clientID      = "math-visual-proofs-server"
	baseClonePath = "/tmp"
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

	messageClient.Subscribe(mqtt.RenderStartTopic, logger, func(renderMessage mqtt.RenderMessage) {
		err = subscribeHandler(renderMessage)
		if err != nil {
			handleError(err, renderMessage.RepoURL, renderMessage.GithubSHA)
			return
		}

		logger.Info("successfully rendered and uploaded: ", renderMessage)

		err = messageClient.PublishRenderFeedbackMessage(mqtt.RenderSuccessTopic, mqtt.RenderFeedbackMessage{
			Status:    mqtt.StatusSucceess,
			RepoURL:   renderMessage.RepoURL,
			GithubSHA: renderMessage.GithubSHA,
			Message:   "successfully finished render, video is uploaded to storage",
		})
		if err != nil {
			logger.Errorf("publishing success message: %w", err)
		}
	})

	for {
		time.Sleep(time.Hour)
	}
}

func subscribeHandler(renderMessage mqtt.RenderMessage) error {
	clonePath := fmt.Sprintf("%s/%s", baseClonePath, renderMessage.GithubSHA)
	defer os.RemoveAll(clonePath)

	err := os.RemoveAll(clonePath)
	if err != nil {
		return fmt.Errorf("removing existing cloned repository: %s", err.Error())
	}

	err = git.Clone(renderMessage.RepoURL, clonePath)
	if err != nil {
		return fmt.Errorf("cloning repository %s: %s", renderMessage.RepoURL, err.Error())
	}

	for _, f := range renderMessage.FileNames {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", clonePath, f)); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file %s not found, cannot render", renderMessage.FileNames)
		}

		err = render(f, clonePath)
		if err != nil {
			return fmt.Errorf("error rendering: %s", err.Error())
		}

		name := strings.Trim(f, ".py")
		path := fmt.Sprintf("%s/media/videos/%s/720p30/%s.mp4", clonePath, name, name)
		// TODO: get org and repo as args
		tmpOrgRepoPath := strings.Replace(renderMessage.RepoURL, "https://github.com/", "", -1)
		orgRepoPath := strings.Replace(tmpOrgRepoPath, ".git", "", -1)
		key := fmt.Sprintf("%s/%s.mp4", orgRepoPath, name)
		err = awsClient.UploadFile(context.Background(), path, key, map[string]string{
			"x-amz-meta-sha":     renderMessage.GithubSHA,
			"x-amz-meta-repourl": renderMessage.RepoURL,
		})
		if err != nil {
			return fmt.Errorf("error uploading to s3: %w", err)
		}
	}

	return nil
}

func render(fileName, clonePath string) error {
	c := fmt.Sprintf(`docker run --rm --user="$(id -u):$(id -g)" -v "%s":/manim manimcommunity/manim:stable manim %s -qm --progress_bar none`, clonePath, fileName)
	cmd := exec.Command("bash", "-c", c)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(string(out))
		return fmt.Errorf("%w, %s", err, string(out))
	}

	return nil
}

func handleError(err error, repoURL, githubSHA string) {
	logger.Error(err)
	pubErr := messageClient.PublishRenderFeedbackMessage(mqtt.RenderErrTopic, mqtt.RenderFeedbackMessage{
		Status:    mqtt.StatusSucceess,
		RepoURL:   repoURL,
		GithubSHA: githubSHA,
		Message:   fmt.Sprintf("error during render: %s", err.Error()),
	})
	if pubErr != nil {
		logger.Errorf("error publishing to renderErrTopic: %s", pubErr)
	}
}
