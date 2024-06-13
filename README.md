# weather-to-speech
A simple bash script to use wttr.in + openai to produce spoken weather prognosis from terminal. 

## Usage

You have to install `afplay` on you system first. Then set `OPENAI_API_KEY` enivorment variable.
```
export OPENAI_API_KEY="*********"
```
and run 
```
bash weather.sh <city-name>
```