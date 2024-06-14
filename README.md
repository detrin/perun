# weather-to-speech
A simple bash script to use wttr.in + openai to produce spoken weather prognosis from terminal. 

## Usage Shell script

You have to install `curl`, `jq` and `afplay` on you system first. Then set `OPENAI_API_KEY` enivorment variable.
```
export OPENAI_API_KEY="*********"
```
and run 
```
bash weather.sh <city-name>
```

## Usage Go script

```
go get -u github.com/gopxl/beep  
go get -u github.com/gopxl/beep/mp3
go get -u github.com/gopxl/beep/speaker
go get -u github.com/tidwall/gjson
go get -u github.com/jessevdk/go-flags
```