#!/bin/bash

# Install python3 packages
pip3 install --user -r requirements.txt

# Install go packages
cd inge && go mod tidy

