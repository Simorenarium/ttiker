package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	log.Printf("Getting env vars: TG_API_KEY, TANKERKOENIG_API_KEY, TARGET_CHAT_ID")

	tgApiKey := os.Getenv("TG_API_KEY")
	apiKey := os.Getenv("TANKERKOENIG_API_KEY")
	chatId, _ := strconv.ParseInt(os.Getenv("TARGET_CHAT_ID"), 10, 64)

	if tgApiKey == "" || apiKey == "" || chatId == 0 {
		log.Fatal("Missing env vars: TG_API_KEY, TANKERKOENIG_API_KEY, TARGET_CHAT_ID")
		return
	}

	gasStations := map[string]string{
		"df44694b-e38d-4b5c-8323-1c27afba3d0b": "Shell Holzappel Hauptstr. 104",
		"20947e99-074e-4fb3-9e16-3813330beaac": "Shell Goergeshausen In der Neuwiese 1",
		"5ee86c30-2760-4a35-b582-f492cb6fee82": "ED Diez",
		"dbc2d764-3330-400e-a0b6-b8ca113149e5": "Shell Diez Wilhelmstr. 58 A",
	}

	bot, err := tgbotapi.NewBotAPI(tgApiKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	messages := make([]string, 0)
	for id, name := range gasStations {
		prices, err := getPrices(id, apiKey)
		if err != nil {
			log.Printf("Error getting prices for gas station %s: %v", name, err)
			continue
		}

		messages = append(messages, fmt.Sprintf("%s[e10: %.3f, e5: %.3f]", name, prices["e10"], prices["e5"]))
	}

	if len(messages) > 0 {
		msg := tgbotapi.NewMessage(chatId, strings.Join(messages, "\n"))
		bot.Send(msg)
	}
}

func getPrices(id string, apikey string) (map[string]float64, error) {
	resp, err := http.Get("https://creativecommons.tankerkoenig.de/json/detail.php?apikey=" + apikey + "&id=" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}
	prices := make(map[string]float64)
	prices["e5"], err = extractPrice("e5", m["station"].(map[string]interface{}))
	prices["e10"], err = extractPrice("e10", m["station"].(map[string]interface{}))
	return prices, err
}

func extractPrice(fuelType string, data map[string]interface{}) (float64, error) {
	log.Printf("Extracting price for fuel type on %s", data)
	price, ok := data[fuelType]
	if !ok {
		return 0, errors.New("Fuel type not found")
	}
	return price.(float64), nil
}
