SHELL=/bin/bash

# Define environment variables
export AUTH_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlbl91c2UiOiJhY2Nlc3MiLCJpc3MiOiJnby1vYXV0aC1sYW1iZGEiLCJ2ZXJzaW9uIjoiMyIsImp3dF9pZCI6IjU3NDVkNjlhLTAyZTItNDQyOS1hMTU1LWM5ZTUzMzMwNGFkYSIsInVzZXJuYW1lIjoiYWRtaW4iLCJzY29wZSI6WyJhZG1pbiJdLCJleHAiOjE3NDU5MTUwMTl9.aQSUdOZYM-EpdXkn_FXGp6MBnWRXn0u40r3jijEP-RdoE-4VLkRH0iJpzclPlZTTNb9cHeFIxaHAbH1sFwLN8Pzi8LjlP39g2X53wfn77xI_v_a0DtIwtGCIxd16cVyPS7578Jg4MXlAnxvsdsYOWG2ouTnx7FDgXcgxNJn92c1YuupXmWTRRCaoiqKHQqIFhvDUbAJcMwSHYMK9rzBHsssuFnV709jGrOyiBg18GdEKldeGZdWyw94_-K8HCjpW7DC9gz1Wex1VYkApNGHiSbtCu6zGrCwnWXD-AEMAjjlUD24f_pPwdxYXj4v8tPq28utwN8U83SOqXrjFC5x_Qn8tCuWMB2vBBmaAoaCnxnyLea1RDxL1GA4Xb3yrhuvGMMxFiYoMnSXchODGHASYMtLK601fzQpZX2LS1JJepje_b_FBRkEruHktwnIBjM8LgcIfvRhsljwCRF2glHBWT9BGmqYi9gO4TcfeKUUg_Few7FBNYDfvEPsYRKbIm4G3cdCmZdr12ElDpsrG2DKyvNdrgjT5Q3o9U7eHM3lE8wbEEw2EXtdNSm3fDK4KVRUm9mRyhA2a21huN8t5i-fYZHcXBYrtneVBLT2-gMIqb3GXdC-QPNj-2Te-4CT2IUgnPlKPtC3xDTDM6euaCs_OcOtd-Jym8aKnnU-Blssf5lc
export URL_POST=https://go-global.architecture.caradhras.io/card/card

# Default target
all: env load

# Show environment variables
env:
	@echo "Current Environment Variables:"
	@echo "AUTH_TOKEN=$(AUTH_TOKEN)"
	@echo "URL_POST=$(URL_POST)"

load:
	@echo "Run Load Card..."

	@for ((i=100; i<=999; i++)); do \
		echo "Posting iteration $$i... {"card_number":"111.111.000.$$i","account_id":"ACC-$$i","holder":"holder-$$i","type":"CREDIT","model":"CHIP","status":"ISSUED"} "; \
		curl -X POST $(URL_POST) \
		    --header "Content-Type: application/json" \
			--header "Authorization: $(AUTH_TOKEN)" \
		    --data '{"card_number":"111.111.000.'$$i'","account_id":"ACC-'$$i'","holder":"holder-'$$i'","type":"CREDIT","model":"CHIP","status":"ISSUED"}'; \
		echo ""; \
	done

.PHONY: all env load