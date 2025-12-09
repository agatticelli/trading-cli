.PHONY: build clean test run

build:
	go build -o trading-cli .

clean:
	rm -f trading-cli

test:
	go test ./...

run:
	./trading-cli --demo balance

# Development helpers
fmt:
	go fmt ./...

vet:
	go vet ./...

deps:
	go mod download
	go mod tidy

# Quick test commands
demo-balance:
	./trading-cli --demo balance

demo-positions:
	./trading-cli --demo positions

demo-open:
	./trading-cli --demo open \
		--symbol ETH-USDT \
		--side long \
		--entry 3950 \
		--sl 3900 \
		--risk 2 \
		--rr 2
