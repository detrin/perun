# Perun

A simple bash script to use api.open-meteo.com + openai to produce weather summary in text/speech.

Perun is a Slavic god of thunder and lightning and king of the gods.

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