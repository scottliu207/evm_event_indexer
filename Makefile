ABIGEN = abigen
OUT_DIR = out
GEN_DIR = generated
CONTRACT_NAME = BasicERC20
PKG = basic_erc_20

.PHONY: gen
gen: $(GEN_DIR)/$(PKG).go

# 用 forge 編譯出 JSON
$(OUT_DIR)/$(CONTRACT_NAME).sol/$(CONTRACT_NAME).json:
	@echo "Compiling contracts with Forge..."
	forge build

# 從 JSON 提取 ABI 和 BIN 並生成 Go binding
$(GEN_DIR)/$(PKG).go: $(OUT_DIR)/$(CONTRACT_NAME).sol/$(CONTRACT_NAME).json
	@mkdir -p $(GEN_DIR)
	@echo "Extracting ABI and BIN from Forge output..."
	jq -r '.abi' $(OUT_DIR)/$(PKG).sol/$(CONTRACT_NAME).json > $(OUT_DIR)/$(PKG).sol/$(CONTRACT_NAME).abi
	jq -r '.bytecode.object' $(OUT_DIR)/$(PKG).sol/$(CONTRACT_NAME).json > $(OUT_DIR)/$(PKG).sol/$(CONTRACT_NAME).bin
	@echo "Generating Go binding..."
	$(ABIGEN) --abi $(OUT_DIR)/$(PKG).sol/$(CONTRACT_NAME).abi \
		--bin $(OUT_DIR)/$(PKG).sol/$(CONTRACT_NAME).bin \
		--pkg $(PKG) --out $(GEN_DIR)/$(PKG).go

.PHONY: clean
clean:
	rm -rf $(OUT_DIR) $(GEN_DIR) cache
