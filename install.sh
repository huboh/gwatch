#!/bin/sh

# Download the latest version of Gwatch and install it
echo "Installing Gwatch..."

GO_BIN_PATH=$(go env GOPATH)/bin

# Download the latest version
go install github.com/huboh/gwatch/cmd/gwatch@latest

# Ensure the Go binary path is in the user's PATH
if ! echo "$PATH" | grep -q "$GO_BIN_PATH"; then
    echo "export PATH=\$PATH:$GO_BIN_PATH" >> ~/.zshrc
    echo "export PATH=\$PATH:$GO_BIN_PATH" >> ~/.bashrc
    source ~/.zshrc
    source ~/.bashrc
fi

echo "Gwatch installation complete."