SHELL=/bin/bash

# Define environment variables
export AUTH_TOKEN=

export URL_POST_CARD=https://go-global-apex.architecture.caradhras.io/card/card

# Default target
all: env load

# Show environment variables
env:
	@echo "Current Environment Variables:"
	@echo "AUTH_TOKEN=$(AUTH_TOKEN)"
	@echo "URL_POST_CARD=$(URL_POST_CARD)"

load:
	@echo "Run Load Card..."
	@PAN=555000000000; \
	for ((i=1; i<=1000; i++)); do \
		ACC_ID=$$((4999 + i)); \
		PAN_ID=$$((PAN + ACC_ID)); \
		CARD_PAN="$$(printf "%012d" $$PAN_ID | sed -E 's/(.{3})(.{3})(.{3})(.{3})/\1.\2.\3.\4/')"; \
		echo "Posting iteration $$i... {"card_number":"$$CARD_PAN","account_id":"ACC-$$ACC_ID","holder":"holder-$$ACC_ID","type":"CREDIT","model":"CHIP","status":"ISSUED"} "; \
		curl -X POST $(URL_POST_CARD) \
		    --header "Content-Type: application/json" \
			--header "Authorization: $(AUTH_TOKEN)" \
		    --data '{"card_number":"'$$CARD_PAN'","account_id":"ACC-'$$ACC_ID'","holder":"holder-'$$ACC_ID'","type":"CREDIT","model":"CHIP","status":"ISSUED"}'; \
		echo ""; \
	done

siege_get_token:
	@echo "Run card get token  ..."

	@siege -c80 -t2m -d0.5 -v --content-type "application/json" --header="Authorization: $(AUTH_TOKEN)" '$(URL_POST_TOKEN)'

siege_atc:
	@echo "Run card atc ..."

	@siege -c50 -t60s -d0.5 -v --content-type "application/json" --header="Authorization: $(AUTH_TOKEN)" '$(URL_POST_ATC) POST {"card_number": "111.111.111.100"}'
	
.PHONY: all env load