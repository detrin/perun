#!/usr/bin/env bash

set -e

weather=$(curl -s wttr.in/$1 | head -n 17 | tail -n 10)

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