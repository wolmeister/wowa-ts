# Directories
DST_FOLDER = dist
SRC_FOLDER = src

# Source files
TS_FILES = $(shell find $(SRC_FOLDER) -name '*.ts')

# Output files
LINUX_OUTPUT = $(DST_FOLDER)/wowa-linux64
WIN_OUTPUT = $(DST_FOLDER)/wowa-win64.exe

# Temporary object file to track changes
TEMP_LINT_FILE = $(DST_FOLDER)/.lint_timestamp

# Targets
all: $(LINUX_OUTPUT) $(WIN_OUTPUT)

$(DST_FOLDER):
	mkdir -p $(DST_FOLDER)

$(LINUX_OUTPUT): $(TS_FILES)
	bun build --compile --minify --bytecode --target=bun-linux-x64-modern --outfile $@ src/index.ts

$(WIN_OUTPUT): $(TS_FILES)
	bun build --compile --minify --bytecode --target=bun-windows-x64-modern --outfile $@ src/index.ts

# Development target
dev:
	bun run --watch src/index.ts

# Lint target
lint: $(TEMP_LINT_FILE)

$(TEMP_LINT_FILE): $(TS_FILES)
	bunx @biomejs/biome check ./src --apply-unsafe --verbose && bunx tsc
	@touch $@

# Clean target
clean:
	rm -rf $(DST_FOLDER) $(TEMP_LINT_FILE)

# Phony targets
.PHONY: all clean dev lint
