package aiw3defi

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
)

// AIW3DefiClientFactory creates instances of AIW3DefiClient for load testing
type AIW3DefiClientFactory struct {
	txConfig client.TxConfig
}

var _ loadtest.ClientFactory = (*AIW3DefiClientFactory)(nil)

// NewAIW3DefiClientFactory creates a new factory for AIW3 DeFi clients
func NewAIW3DefiClientFactory(txConfig client.TxConfig) *AIW3DefiClientFactory {
	return &AIW3DefiClientFactory{
		txConfig: txConfig,
	}
}

// AIW3DefiClient generates bank send transactions for load testing
type AIW3DefiClient struct {
	txConfig      client.TxConfig
	chainID       string
	denom         string
	transferAmount sdk.Int
	senderKey     cryptotypes.PrivKey
	senderAddr    sdk.AccAddress
	recipientAddr sdk.AccAddress
	accountNumber uint64
	sequence      uint64
}

var _ loadtest.Client = (*AIW3DefiClient)(nil)

func (f *AIW3DefiClientFactory) ValidateConfig(cfg loadtest.Config) error {
	// Validate that the configuration is compatible with AIW3 DeFi
	if cfg.Connections <= 0 {
		return fmt.Errorf("connections must be > 0")
	}
	if cfg.Rate <= 0 {
		return fmt.Errorf("rate must be > 0")
	}
	return nil
}

func (f *AIW3DefiClientFactory) NewClient(cfg loadtest.Config) (loadtest.Client, error) {
	// Generate a random mnemonic for this client
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entropy: %w", err)
	}
	
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// Derive sender private key directly from mnemonic
	derivedPriv, err := hd.Secp256k1.Derive()(mnemonic, "", "m/44'/118'/0'/0/0")
	if err != nil {
		return nil, fmt.Errorf("failed to derive sender private key: %w", err)
	}

	senderKey := hd.Secp256k1.Generate()(derivedPriv)
	senderAddr := sdk.AccAddress(senderKey.PubKey().Address())

	// Generate recipient address from different mnemonic
	recipientEntropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recipient entropy: %w", err)
	}
	
	recipientMnemonic, err := bip39.NewMnemonic(recipientEntropy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recipient mnemonic: %w", err)
	}

	// Derive recipient private key
	recipientDerivedPriv, err := hd.Secp256k1.Derive()(recipientMnemonic, "", "m/44'/118'/0'/0/1")
	if err != nil {
		return nil, fmt.Errorf("failed to derive recipient private key: %w", err)
	}

	recipientKey := hd.Secp256k1.Generate()(recipientDerivedPriv)
	recipientAddr := sdk.AccAddress(recipientKey.PubKey().Address())

	// Random transfer amount between 1000 and 10000 uaiw (0.001 to 0.01 AIW)
	randomAmount, err := rand.Int(rand.Reader, big.NewInt(9001))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random amount: %w", err)
	}
	transferAmount := sdk.NewInt(randomAmount.Int64() + 1000)

	return &AIW3DefiClient{
		txConfig:       f.txConfig,
		chainID:        "aiw3defi-devnet", // Default chain ID
		denom:          "uaiw",            // AIW3 DeFi token
		transferAmount: transferAmount,
		senderKey:      senderKey,
		senderAddr:     senderAddr,
		recipientAddr:  recipientAddr,
		accountNumber:  0, // Will be set from actual account info
		sequence:       0, // Will be incremented for each transaction
	}, nil
}

// GenerateTx creates a bank send transaction for load testing
func (c *AIW3DefiClient) GenerateTx() ([]byte, error) {
	// Create bank send message
	msg := banktypes.NewMsgSend(
		c.senderAddr,
		c.recipientAddr,
		sdk.NewCoins(sdk.NewCoin(c.denom, c.transferAmount)),
	)

	// Create transaction builder
	txBuilder := c.txConfig.NewTxBuilder()
	
	// Set messages
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set gas limit and fee
	gasLimit := uint64(200000) // Standard gas limit for bank send
	gasPrice := sdk.NewDecWithPrec(1, 3) // 0.001 uaiw per gas
	feeAmount := gasPrice.MulInt64(int64(gasLimit)).TruncateInt()
	fee := sdk.NewCoins(sdk.NewCoin(c.denom, feeAmount))
	
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(fee)

	// Set memo for identification
	txBuilder.SetMemo(fmt.Sprintf("LoadTest:%s", c.senderAddr.String()[:8]))

	// Create signature data
	sigV2 := signing.SignatureV2{
		PubKey: c.senderKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: c.sequence,
	}

	// Set the signature (empty for now)
	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}

	// Create signing data
	signMode := c.txConfig.SignModeHandler().DefaultMode()
	signerData := authsigning.SignerData{
		ChainID:       c.chainID,
		AccountNumber: c.accountNumber,
		Sequence:      c.sequence,
	}

	// Sign the transaction
	signBytes, err := c.txConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to get sign bytes: %w", err)
	}

	signature, err := c.senderKey.Sign(signBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Update signature with actual signature
	sigV2.Data.(*signing.SingleSignatureData).Signature = signature
	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, fmt.Errorf("failed to set final signatures: %w", err)
	}

	// Increment sequence for next transaction
	c.sequence++

	// Encode transaction to bytes
	txBytes, err := c.txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	return txBytes, nil
}