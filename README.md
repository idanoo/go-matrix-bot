# go-matrix-bot

Golang bot for matrix.    
Uses https://github.com/mautrix/go

```
docker run \
 -v ./tmp:/config \
 -e PUID=501 \
 -e PGID=20 \
 -e "MATRIX_HOST=https://matrix.org" \
 -e "MATRIX_USERNAME=user" \
 -e "MATRIX_PASSWORD=pass" \
 gomatrixbot:latest
```
