NUM_OF_DATAPLANES ?= 10
KUMA_CP_ADDRESS ?= grpcs://localhost:5678

run:
	go run main.go run --total "${NUM_OF_DATAPLANES}" --xds-server-address "${KUMA_CP_ADDRESS}"