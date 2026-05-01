#!/bin/bash

INPUT_FILE="input.csv"
OUTPUT_FILE="violations.csv"

# Create the header for the output file
echo "TXN_ID,REQUEST_TYPE,TOKEN_ID" > "$OUTPUT_FILE"

# Get tokens used in ABAS-BANK-INPUT
awk -F',' '$2 == "ABAS-BANK-INPUT" { print $3 }' "$INPUT_FILE" | sort | uniq > bank_input_tokens.txt

# Get tokens used in ABAS-OUTPUT
