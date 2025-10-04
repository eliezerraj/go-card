SHELL=/bin/bash

# Define environment variables
export AUTH_TOKEN=
export URL_POST_CARD=https://go-global.architecture.caradhras.io/card/card
export URL_POST_ATC=https://go-global.architecture.caradhras.io/card/atc
export URL_POST_TOKEN=https://go-global.architecture.caradhras.io/card/cardToken/696673f52a85f4b994700313b935f908b3526f849a1c5d348dee63265817b758

# Default target
all: env load

# Show environment variables
env:
	@echo "Current Environment Variables:"
	@echo "AUTH_TOKEN=$(AUTH_TOKEN)"
	@echo "URL_POST_CARD=$(URL_POST_CARD)"
	@echo "URL_POST_ATC=$(URL_POST_ATC)"

load:
	@echo "Run Load Card..."

	@for ((i=100; i<=999; i++)); do \
		echo "Posting iteration $$i... {"card_number":"111.111.000.$$i","account_id":"ACC-$$i","holder":"holder-$$i","type":"CREDIT","model":"CHIP","status":"ISSUED"} "; \
		curl -X POST $(URL_POST_CARD) \
		    --header "Content-Type: application/json" \
			--header "Authorization: $(AUTH_TOKEN)" \
		    --data '{"card_number":"111.111.000.'$$i'","account_id":"ACC-'$$i'","holder":"holder-'$$i'","type":"CREDIT","model":"CHIP","status":"ISSUED"}'; \
		echo ""; \
	done

siege_get_token:
	@echo "Run card get token  ..."

	@siege -c80 -t2m -d0.5 -v --content-type "application/json" --header="Authorization: $(AUTH_TOKEN)" '$(URL_POST_TOKEN)'

siege_atc:
	@echo "Run card atc ..."

	@siege -c50 -t60s -d0.5 -v --content-type "application/json" --header="Authorization: $(AUTH_TOKEN)" '$(URL_POST_ATC) POST {"card_number": "111.111.111.100"}'
	
.PHONY: all env load