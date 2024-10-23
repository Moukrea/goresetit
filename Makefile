# Gommit setup
.PHONY: gommit-setup

gommit-setup:
	@echo "Setting up Gommit..."
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File .gommit/gommit-hook-setup.ps1
else
	@sh .gommit/gommit-hook-setup.sh
endif

# Include this in your existing Makefile or use it as a standalone Makefile