package main

import (
	"time"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/config"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/database"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal().Err(err).Msg("config load failed")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal().Err(err).Msg("db connect failed")
	}
	defer db.Close()

	svcs := service.New(db)

	opts := mqtt.NewClientOptions().AddBroker(config.MQTTBroker())
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("mqtt connect")
	}
	defer client.Disconnect(250)

	handler := func(_ mqtt.Client, msg mqtt.Message) {
		if err := svcs.Readings.FromMQTT(msg.Topic(), msg.Payload()); err != nil {
			log.Error().Err(err).Msg("ingest failed")
		}
	}

	if token := client.Subscribe("energy/readings", 0, handler); token.Wait() && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("subscribe failed")
	}

	log.Info().Msg("ingestor running; Ctrl+C to stop")
	for {
		time.Sleep(10 * time.Second)
	}
}
