NUM_OF_DATAPLANES ?= 100
NUM_OF_SERVICES ?= 80
KUMA_CP_ADDRESS ?= grpcs://localhost:5678

run:
	go run main.go run --dps "${NUM_OF_DATAPLANES}" --services "${NUM_OF_SERVICES}" --xds-server-address "${KUMA_CP_ADDRESS}"
