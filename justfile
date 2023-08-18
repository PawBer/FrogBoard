build:
	tailwindcss -i ./style.css -o ./cmd/frogboard/public/style.css
	go build ./cmd/frogboard

run:
	tailwindcss -i ./style.css -o ./cmd/frogboard/public/style.css
	go run ./cmd/frogboard
