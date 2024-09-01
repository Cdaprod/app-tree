# Makefile for app-tree

BINARY_NAME=app-tree
INSTALL_PATH=/usr/local/bin

.PHONY: all build clean install uninstall

all: build

build:
	@echo "Building app-tree..."
	@go build -o $(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)

install: build
	@echo "Installing app-tree to $(INSTALL_PATH)..."
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete! You can now use 'app-tree' from anywhere."

uninstall:
	@echo "Uninstalling app-tree..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation complete."