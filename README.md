# nafue-cli
## the Nafue Cli utilizes the Nafue Library.

# Nafue Security Services
# Menklab LLC

## Requirements
- Go Version >= 1.6.0

## Setup Env
1. Clone this repository.
2. Modify config/config.go to match needed env
3. Install Reflex: go get github.com/cespare/reflex
4. Install GoVendor: go get -u github.com/kardianos/govendor
5. Install Deps: ./utility.sh deps

## Build & Run
go build -o nafue main.go
./nafue **args**





