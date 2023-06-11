package main

import (
	"context"
	"fmt"
	"solana-go-demo/wallet_manager"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	atok "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

var ctx = context.TODO()
var client = rpc.New(rpc.LocalNet.RPC)
var wsClient, err = ws.Connect(context.Background(), rpc.LocalNet.WS)
var commitment = rpc.CommitmentConfirmed
var confirmationCommitment = rpc.ConfirmationStatusConfirmed
var confirmationTimeout = time.Duration(5) * time.Minute
var confirmationDelay = time.Duration(5) * time.Second
var wm = wallet_manager.NewWalletManagerWithOpts(ctx, client, commitment, confirmationCommitment, confirmationTimeout, confirmationDelay, false)

func main() {
	test_mint()
}

func test_mint() {
	to := solana.MustPublicKeyFromBase58("9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6b")
	fmt.Println(to)
	from := solana.NewWallet()
	fmt.Println(from.PublicKey())
	// request airdrp
	out, err := client.RequestAirdrop(ctx, from.PublicKey(), solana.LAMPORTS_PER_SOL*2, rpc.CommitmentConfirmed)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(out)
	time.Sleep(30 * time.Second)
	recent, err := client.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println(err)
		return
	}
	MintTokenAccount := solana.NewWallet()
	lamports, err := client.GetMinimumBalanceForRentExemption(
		context.TODO(),
		82,
		rpc.CommitmentFinalized,
	)
	createInstruction, err := system.NewCreateAccountInstruction(lamports, 82, solana.TokenProgramID, from.PublicKey(), MintTokenAccount.PublicKey()).ValidateAndBuild()
	if err != nil {
		fmt.Println(err)
		return
	}
	mintInstruction, err := token.NewInitializeMintInstructionBuilder().
		SetDecimals(9).
		SetMintAuthority(from.PublicKey()).
		SetMintAccount(MintTokenAccount.PublicKey()).
		SetSysVarRentPubkeyAccount(solana.SysVarRentPubkey).ValidateAndBuild()
	if err != nil {
		fmt.Println(err)
		return
	}
	// create ata
	atokAddress, _, err := solana.FindAssociatedTokenAddress(from.PublicKey(), MintTokenAccount.PublicKey())
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = wm.Client.GetAccountInfoWithOpts(context.TODO(), atokAddress, &rpc.GetAccountInfoOpts{
		Commitment: wm.Commitment,
	})
	var createATAInstruction *atok.Instruction
	if err != nil {
		createATAInstruction = atok.NewCreateInstructionBuilder().
			SetPayer(from.PublicKey()).
			SetMint(MintTokenAccount.PublicKey()).
			SetWallet(from.PublicKey()).
			Build()
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			createInstruction,
			mintInstruction,
			createATAInstruction,
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(from.PublicKey()),
	)
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if from.PublicKey().Equals(key) {
				return &from.PrivateKey
			}
			if MintTokenAccount.PublicKey().Equals(key) {
				return &MintTokenAccount.PrivateKey
			}
			return nil
		},
	)

	sig, err := confirm.SendAndConfirmTransaction(
		ctx,
		client,
		wsClient,
		tx,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sig)
	//test go supply
	res, err := client.GetTokenSupply(ctx, MintTokenAccount.PublicKey(), commitment)
	if err != nil {
		fmt.Println(err)
		return
	}
	spew.Dump(res)
	//mint token to toaddress
	fmt.Println("start mint token to toaddress")
	mintToInstruction := token.NewMintToInstructionBuilder().SetAmount(1000000000).SetMintAccount(MintTokenAccount.PublicKey()).SetDestinationAccount(atokAddress).SetAuthorityAccount(from.PublicKey()).Build()
	// mintToInstruction, err := token_contract.NewMintTokenInstruction(MintTokenAccount.PublicKey(), solana.TokenProgramID, atokAddress, from.PublicKey()).ValidateAndBuild()
	if err != nil {
		fmt.Println(err)
		return
	}
	tx, err = solana.NewTransaction(
		[]solana.Instruction{
			mintToInstruction,
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(from.PublicKey()),
	)
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if from.PublicKey().Equals(key) {
				return &from.PrivateKey
			}
			if MintTokenAccount.PublicKey().Equals(key) {
				return &MintTokenAccount.PrivateKey
			}
			return nil
		})
	sig, err = confirm.SendAndConfirmTransaction(
		ctx,
		client,
		wsClient,
		tx,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sig)
	//test go supply
	res, err = client.GetTokenSupply(ctx, MintTokenAccount.PublicKey(), commitment)
	if err != nil {
		fmt.Println(err)
		return
	}
	spew.Dump(res)
}

func test_transfer() {

}
