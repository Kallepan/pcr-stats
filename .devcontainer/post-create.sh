#!/bin/bash
# Install go packages
cd inge && go mod tidy

# Install python3 packages
cd ..
pip3 install --user -r requirements.txt


