package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTT client instance
var mqttClient mqtt.Client

func initMQTT() {
	// Configure MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(os.Getenv("MQTT_BROKER"))
	opts.SetClientID(os.Getenv("MQTT_CLIENT_ID"))
	_, ok := os.LookupEnv("MQTT_USERNAME")
	if ok {
		opts.SetUsername(os.Getenv("MQTT_USERNAME"))
		opts.SetPassword(os.Getenv("MQTT_PASSWORD"))
	}
	// Create and connect MQTT client
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}
	fmt.Println("Connected to MQTT broker")
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the payload from the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Publish the payload to the MQTT topic
	token := mqttClient.Publish(os.Getenv("MQTT_TOPIC"), 0, false, body)
	token.Wait()

	if token.Error() != nil {
		http.Error(w, "Failed to publish to MQTT", http.StatusInternalServerError)
		return
	}

	// Respond to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payload published to MQTT topic"))
}

func main() {
	// Initialize MQTT client
	initMQTT()

	// Set up HTTP server
	http.HandleFunc("/publish", handlePostRequest)

	fmt.Printf("HTTP server listening on port %s\n", os.Getenv("HTTP_PORT"))
	log.Fatal(http.ListenAndServe(os.Getenv("HTTP_PORT"), nil))
}
