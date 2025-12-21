ABIGEN = abigen
OUT_DIR = out
GEN_DIR = generated
CONTRACT_NAME = BasicERC20
PKG = basic_erc_20

# generate go bindings
.PHONY: gen
gen: $(GEN_DIR)/$(PKG).go

# compile contracts with forge
$(OUT_DIR)/$(CONTRACT_NAME).sol/$(CONTRACT_NAME).json:
	@echo "Compiling contracts with Forge..."
	forge build

# extract abi and bin from forge output and generate go binding
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

# run indexer service
.PHONY: run
run:
	go run cmd/indexer/main.go

# deploy contract to anvil
.PHONY: deploy
deploy:
	docker run --rm --network indexer-network \
		-v $(PWD)/contracts:/contracts \
		-v foundry-svm-cache:/root/.svm \
		-w /contracts \
		--entrypoint sh \
		ghcr.io/foundry-rs/foundry:stable \
		-c "forge create basic_erc_20.sol:BasicERC20 \
		--rpc-url http://anvil:8545 \
		--private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
		--broadcast \
		--constructor-args 'MyToken' 'MTK' 100000000000000000000" \


# transfer tokens to another address
.PHONY: transfer
transfer:
	docker run --rm --network indexer-network \
		--entrypoint cast \
		ghcr.io/foundry-rs/foundry:stable \
		send --rpc-url http://anvil:8545 \
		--private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
		0x5FbDB2315678afecb367f032d93F642f64180aa3 \
		"transfer(address,uint256)" \
		0x70997970C51812dc3A010C7d01b50e0d17dc79C8 \
		1000000000000000000
