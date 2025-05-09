SHELL=/bin/bash

# Define environment variables
export AUTH_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlbl91c2UiOiJhY2Nlc3MiLCJpc3MiOiJnby1vYXV0aC1sYW1iZGEiLCJ2ZXJzaW9uIjoiMyIsImp3dF9pZCI6Ijc1ZmE5YjExLTFlYzUtNGFhYi04YzBlLWYxZDUyZGRkYzg2NSIsInVzZXJuYW1lIjoiYWRtaW4iLCJzY29wZSI6WyJhZG1pbiJdLCJleHAiOjE3NDYzNDYzODZ9.BfOao-2LGwo_aFJcMmuuI-lD4F46B5Yr4_A3PbiH9WD3-V6qNUQEQQZ6O-wo0LVDe72clJimfyoyKAGagG4y14kZG6Zpt6JCbrvzZTRkreUmelXb7a3pl1__g3VGXhDwIH0V0O519M0Aa4HwteGiqposJNzUPtW11QCVmZDrnvTXXubV5FzS23uYg7LUJyOQovkCwb9So9nw3WD4heYne0PsfTkD8LQ8H1F42GPm60ncXyMT2Hj1txEJhMHdNhZ2dUeuc2lFwY7OAolOetlL5eb4iWAK9EDGgEboD-Zf8JbcF-Erzbho-Tj4f1b7b7-63Mx0_JGAQGDV8hTli3M1KLFAGVJXoyE3cVNlSi_jNnRQIkxLiZqrdoFDK8Md6v-QzoO7jB1QfMy4owpaixgl-tkymFogsnB5kjj9er5E5ulElNgJnFAZ3gvorrR0VZAeWpLT0ErKsGo2Q7Zsrt2RGjo5ticl-Fq2vxyRp8Rb9kWOrqsEj1ZFY_-Dpv-PjdOf1JMEgYQjKxz6nnS5Dy6t4JZ7h1jE1yiwYvzRmnsrSPD8z55fvaTvQmVfMKWkQtiwP6qcbjqJwM36JHFFp8bd3cmuusBBTDDuL1NlkzZdwBVkhgcCa2eQkIxM5XXUZtUsQ312iS10FIllDlkkM9hZ9188k8OVks23DxiSLIkHCyM
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