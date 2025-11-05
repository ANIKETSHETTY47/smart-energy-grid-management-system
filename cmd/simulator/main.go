package main

import (
	"encoding/json"
	"math/rand"
	"time"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"smart/internal/config"
)

type Reading struct {
	MeterID   string    `json:"meter_id"`
	Timestamp time.Time `json:"timestamp"`
	Voltage   float64   `json:"voltage"`
	Current   float64   `json:"current"`
	PowerKW   float64   `json:"power_kw"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := config.Load(); err != nil {
		log.Fatal().Err(err).Msg("config load failed")
	}
	opts := mqtt.NewClientOptions().AddBroker(config.MQTTBroker())
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("mqtt connect")
	}
	defer client.Disconnect(250)

	for i := 0; i < 100; i++ {
		r := Reading{
			MeterID:   "meter-001",
			Timestamp: time.Now(),
			Voltage:   220 + rand.Float64()*10,
			Current:   5 + rand.Float64()*2,
			PowerKW:   1 + rand.Float64(),
		}
		payload, _ := json.Marshal(r)
		token := client.Publish("energy/readings", 0, false, payload)
		token.Wait()
		time.Sleep(500 * time.Millisecond)
	}
	log.Info().Msg("simulation done")
}
