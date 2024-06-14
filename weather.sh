#!/usr/bin/env bash

set -e

weather=$(curl -s "https://api.open-meteo.com/v1/forecast?latitude=50.0874654&longitude=14.4212503&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation&daily=temperature_2m_max,temperature_2m_min&forecast_days=1" | jq '{
  latitude, 
  longitude, 
  generationtime_ms, 
  utc_offset_seconds, 
  timezone, 
  timezone_abbreviation, 
  elevation, 
  hourly_units,
  hourly_records: [ 
    .hourly.time as $time
    | .hourly.temperature_2m as $temperature_2m 
    | .hourly.apparent_temperature as $apparent_temperature 
    | .hourly.precipitation_probability as $precipitation_probability 
    | .hourly.precipitation as $precipitation 
    | range(0; .hourly.time | length) 
    | { 
        time: $time[.], 
        temperature_2m: $temperature_2m[.], 
        apparent_temperature: $apparent_temperature[.], 
        precipitation_probability: $precipitation_probability[.], 
        precipitation: $precipitation[.] 
      }
  ],
  daily_units,
  daily
}')

json_payload=$(jq -n \
    --arg model "gpt-4o" \
    --arg role_user "user" \
    --arg user_content "You are a meteorologist. Summarize the weather in two sentences. Tell me if I need an umbrella. $weather" \
    '{
    model: $model,
    messages: [
        {
        role: $role_user,
        content: $user_content
        }
    ],
    temperature: 1,
    max_tokens: 256,
    top_p: 1,
    frequency_penalty: 0,
    presence_penalty: 0
    }')

response=$(curl -s -X POST https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d "$json_payload")
response_text=$(echo "$response" | jq -r '.choices[0].message.content')

tts_json_payload=$(jq -n \
  --arg input "$response_text" \
  --arg voice "nova" \
  '{
    model: "tts-1",
    input: $input,
    voice: $voice
  }')

curl -s -X POST https://api.openai.com/v1/audio/speech \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d "$tts_json_payload" \
  --output weather.mp3

afplay weather.mp3 2>/dev/null || mpg123 weather.mp3 2>/dev/null || echo "Please install a compatible audio player to play the speech.mp3 file"
rm weather.mp3