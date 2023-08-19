# ghosthugo

Import content from ghost cms

## Build

```
docker build -t ghosthugo .
```

## Config

Place a `.env` file in your hugo repo with the `GHOST_URL` and `GHOST_KEY` values:

```
GHOST_URL=http://127.0.0.1:3001
GHOST_KEY=deadbeefa20ebee47bafbb714
```

## Use

The entrypoint executes `ghosthugo`, imports the content, and then passes any arguments
to the hugo binary in the container. You might want to pass the current folder to let
it write into the `content/posts` folder. Exposing the hugo port can also come handy:

```
docker run --rm -it -v .:/site -p 1313:1313 ghosthugo server --bind 0.0.0.0 -p 1313
```
