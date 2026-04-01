package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	authmail "github.com/go-sum/forge/internal/adapters/authmail"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/repository"
	"github.com/go-sum/forge/config"
	authmodel "github.com/go-sum/auth/model"
	authsvc "github.com/go-sum/auth/service"
	"github.com/go-sum/server/database"
	"github.com/go-sum/send"
	"github.com/spf13/cobra"
)

func newAdminCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "admin <email>",
		Short: "Elevate a verified user to admin (only when no admin exists)",
		Args:  cobra.ExactArgs(1),
		RunE:  runAdmin,
	}
}

func runAdmin(_ *cobra.Command, args []string) error {
	email := args[0]
	ctx := context.Background()

	// Load config (honours APP_ENV overlay, e.g. development).
	cfg, err := config.LoadFrom("config", os.Getenv("APP_ENV"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Connect to the database.
	pool, err := database.Connect(ctx, cfg.DSN())
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer pool.Close()

	// Build the repository layer.
	repos := repository.NewRepositories(pool)

	// Confirm the target user exists before sending any email.
	_, err = repos.User.GetByEmail(ctx, email)
	if errors.Is(err, model.ErrUserNotFound) {
		return fmt.Errorf("no user found with email %s", email)
	}
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}

	// Guard: elevation is only permitted when no admin account exists yet.
	hasAdmin, err := repos.User.HasAdmin(ctx)
	if err != nil {
		return fmt.Errorf("check admin: %w", err)
	}
	if hasAdmin {
		return fmt.Errorf("an admin account already exists; elevation via CLI is not permitted")
	}

	// Wire the auth service exactly as bootstrap.go does in initServices().
	sender, err := send.New(cfg.Service.Send.Delivery)
	if err != nil {
		return fmt.Errorf("init sender: %w", err)
	}
	sendFrom := send.DefaultRegistry.SendFrom(cfg.Service.Send.Delivery)
	notifier := authmail.New(sender, sendFrom)
	authSvc := authsvc.NewAuthService(repos.User, authsvc.Config{
		Method:   cfg.Service.Auth.Methods.EmailTOTP,
		Notifier: notifier,
		TokenCodec: authsvc.NewEncryptedTokenCodec(
			cfg.App.Session.AuthKey,
			cfg.App.Session.EncryptKey,
		),
	})

	// Send the TOTP verification code to the user's email address.
	fmt.Printf("Sending verification code to %s ...\n", email)
	flow, err := authSvc.BeginSignin(ctx, authmodel.BeginSigninInput{Email: email}, "")
	if err != nil {
		return fmt.Errorf("begin signin: %w", err)
	}

	// Prompt the operator for the code interactively.
	fmt.Print("Enter verification code: ")
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return fmt.Errorf("read code: %w", err)
	}

	// Verify the code against the pending flow.
	result, err := authSvc.VerifyPendingFlow(ctx, flow, authmodel.VerifyInput{Code: code})
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Elevate. Empty email/displayName → COALESCE keeps existing values.
	if _, err := repos.User.Update(ctx, result.User.ID, "", "", model.RoleAdmin); err != nil {
		return fmt.Errorf("elevate user: %w", err)
	}

	fmt.Printf("User %s is now an admin.\n", email)
	return nil
}
