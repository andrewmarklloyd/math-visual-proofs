package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mqttC "github.com/eclipse/paho.mqtt.golang"

	"github.com/andrewmarklloyd/math-visual-proofs/internal/pkg/mqtt"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

var messageClient mqtt.MqttClient

func main() {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Error creating logger:", err)
	}
	logger = l.Sugar().Named("math-visual-proofs")
	defer logger.Sync()

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
		fmt.Println(message)
	})

	for {
		time.Sleep(time.Hour)
	}

}
