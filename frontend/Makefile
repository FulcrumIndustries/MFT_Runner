# Frontend build system
NPM := npm
TAR := tar -zcvf
DIST_DIR := dist
PACKAGE_NAME := frontend-dist.tar.gz

.PHONY: install build clean serve package

install:
	@echo "Installing dependencies..."
	@$(NPM) install

build: install
	@echo "Building production assets..."
	@$(NPM) run build

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf node_modules $(DIST_DIR) $(PACKAGE_NAME)

serve: build
	@echo "Starting production server..."
	@$(NPM) run preview

package: build
	@echo "Packaging distribution files..."
	@$(TAR) $(PACKAGE_NAME) $(DIST_DIR)
	@echo "Package created: $(PACKAGE_NAME)"

deploy: package
	@echo "Ready for deployment - copy $(PACKAGE_NAME) to target server"

help:
	@echo "MFT Runner Frontend Build System"
	@echo "Usage:"
	@echo "  make install     - Install dependencies"
	@echo "  make build      - Build production version"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make serve      - Start production server"
	@echo "  make package    - Create deployable tarball"
	@echo "  make deploy     - Prepare for deployment" 