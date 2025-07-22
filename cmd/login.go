package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to system",
	RunE:  loginCmdF,
}

func init() {
	RootCmd.AddCommand(LoginCmd)
}

func loginCmdF(command *cobra.Command, args []string) error {
	client, err := InitAnonClientCommandContextCobra(command)
	if err != nil {
		return err
	}

	fmt.Printf("Log in on %s\n", client.BaseURL)

	emailPrompt := promptui.Prompt{
		Label:    "Email",
		Validate: validateEmail,
	}

	email, err := emailPrompt.Run()
	if err != nil {
		return fmt.Errorf("wrror entering email: %v", err)
	}

	passwordPrompt := promptui.Prompt{
		Label:    "Password",
		Mask:     '*',
		Validate: validatePassword,
	}

	password, err := passwordPrompt.Run()
	if err != nil {
		return fmt.Errorf("Error entering password: %v", err)
	}

	user, authResponse, err := client.Login(email, password)

	if err != nil {
		return err
	}

	authRC := NewAuthRC(authResponse)

	err = SaveAuthRC(client.BaseURL, authRC)
	if err != nil {
		return fmt.Errorf("failed save token to file: %v", err)
	}

	fmt.Printf("Welcome, %s\n", user.Username)

	return nil
}

func validateEmail(input string) error {
	if len(input) == 0 {
		return fmt.Errorf("Email cannot be empty")
	}
	return nil
}

func validatePassword(input string) error {
	if len(input) < 8 {
		return fmt.Errorf("Password must be at least 6 characters long")
	}
	return nil
}
