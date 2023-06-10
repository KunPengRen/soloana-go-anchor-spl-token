package token_contract

import (
	"context"
	"fmt"
	"solana-go-demo/wallet_manager"
	"time"

	"github.com/gagliardetto/solana-go"
	atok "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

var ctx = context.TODO()
var client = rpc.New(rpc.LocalNet.RPC)
var commitment = rpc.CommitmentConfirmed
var confirmationCommitment = rpc.ConfirmationStatusConfirmed
var confirmationTimeout = time.Duration(5) * time.Minute
var confirmationDelay = time.Duration(5) * time.Second
var wm = wallet_manager.NewWalletManagerWithOpts(ctx, client, commitment, confirmationCommitment, confirmationTimeout, confirmationDelay, false)

func main() {

}

func test_mint() {
	var instructions []solana.Instruction
	key := solana.NewWallet()
	_, err := airdrop(key.PublicKey(), 1000)
	if err != nil {
		fmt.Printf("failed to request airdrop: %s", err.Error())
	}
	mint := solana.NewWallet()
	atokAddress, _, err := solana.FindAssociatedTokenAddress(key.PublicKey(), mint.PublicKey())

	_, err = wm.Client.GetAccountInfoWithOpts(context.TODO(), atokAddress, &rpc.GetAccountInfoOpts{
		Commitment: wm.Commitment,
	})
	var createInstruction *atok.Instruction
	if err != nil {
		createInstruction = atok.NewCreateInstructionBuilder().
			SetPayer(key.PublicKey()).
			SetMint(mint.PublicKey()).
			SetWallet(key.PublicKey()).
			Build()
	}
	instructions = append(instructions, createInstruction)

}
func test_transfer() {

}

func airdrop(receiver solana.PublicKey, lamports uint64) (solana.Signature, error) {
	airDropSig, err := client.RequestAirdrop(ctx, receiver, lamports, commitment)
	if err != nil {
		return solana.Signature{}, errors.Errorf("failed to request airdrop: %s", err.Error())
	}
	awaitedSig, err := wm.awaitSignaturesConfirmation([]solana.Signature{airDropSig})
	if err != nil {
		return solana.Signature{}, errors.Errorf("failed to confirm airdrop: %s", err.Error())
	}
	return awaitedSig, err
}
