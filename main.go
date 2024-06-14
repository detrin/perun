package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/jessevdk/go-flags"
	"github.com/tidwall/gjson"
)

const (
	version = "v0.1.0"
	author  = "Daniel Herman"
	repo    = "https://github.com/detrin/weather-to-speech"
)

type Options struct {
	City    string  `short:"c" long:"city" description:"City name (default: prague)" default:"prague"`
	Lat     float64 `short:"l" long:"lat" description:"Latitude if city is not defined"`
	Lon     float64 `short:"o" long:"lon" description:"Longitude if city is not defined"`
	APIKey  string  `short:"a" long:"api_key" description:"OpenAI API key"`
	Version bool    `short:"v" long:"version" description:"Show version information and exit"`
}

var CityCoordinates = map[string][2]float64{
	"prague": {50.0874654, 14.4212503},
}

func getOpenAIKey(apiKey string) string {
	if apiKey != "" {
		return apiKey
	}
	return os.Getenv("OPENAI_API_KEY")
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Println("Version: ", version)
		fmt.Println("Author: ", author)
		fmt.Println("Repository: ", repo)
		return
	}

	openAIKey := getOpenAIKey(opts.APIKey)
	if openAIKey == "" {
		log.Fatal("OpenAI API key is required. Please provide it using the --api_key flag or the OPENAI_API_KEY environment variable.")
	}

	var latitude, longitude float64
	lowercaseCity := strings.ToLower(opts.City)
	if coords, ok := CityCoordinates[lowercaseCity]; ok {
		latitude, longitude = coords[0], coords[1]
	} else if opts.Lat != 0.0 && opts.Lon != 0.0 {
		latitude, longitude = opts.Lat, opts.Lon
	} else {
		log.Fatal("Please provide a city name from the predefined list or latitude and longitude")
	}

	weatherUrl := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation&daily=temperature_2m_max,temperature_2m_min&forecast_days=1", latitude, longitude)
	resp, err := http.Get(weatherUrl)
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v", err)
	}
	defer resp.Body.Close()
	weatherData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read weather data: %v", err)
	}

	weatherJson := gjson.ParseBytes(weatherData)
	userContent := fmt.Sprintf("You are a meteorologist. Summarize the weather in two sentences. Tell me if I need an umbrella. %s", weatherJson.String())

	gptJsonPayload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": userContent,
			},
		},
		"temperature":       1,
		"max_tokens":        256,
		"top_p":             1,
		"frequency_penalty": 0,
		"presence_penalty":  0,
	}

	gptJson, err := json.Marshal(gptJsonPayload)
	if err != nil {
		log.Fatalf("Failed to marshal GPT payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(gptJson))
	if err != nil {
		log.Fatalf("Failed to create GPT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openAIKey))

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get GPT response: %v", err)
	}
	defer resp.Body.Close()
	gptRespData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read GPT response: %v", err)
	}

	responseText := gjson.GetBytes(gptRespData, "choices.0.message.content").String()
	ttsJsonPayload := map[string]interface{}{
		"model": "tts-1",
		"input": responseText,
		"voice": "nova",
	}
	ttsJson, err := json.Marshal(ttsJsonPayload)
	if err != nil {
		log.Fatalf("Failed to marshal TTS payload: %v", err)
	}

	req, err = http.NewRequest("POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(ttsJson))
	if err != nil {
		log.Fatalf("Failed to create TTS request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openAIKey))

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get TTS response: %v", err)
	}
	defer resp.Body.Close()
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read TTS response: %v", err)
	}

	audioFilePath := "weather.mp3"
	if err := os.WriteFile(audioFilePath, audioData, 0644); err != nil {
		log.Fatalf("Failed to save audio file: %v", err)
	}

	f, err := os.Open(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done

	if err := os.Remove(audioFilePath); err != nil {
		log.Fatalf("Failed to remove audio file: %v", err)
	}
}
