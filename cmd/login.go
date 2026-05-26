package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/statusmate/statusmatectl/pkg/api"
)

var (
	usernameFlag      string
	passwordFlag      string
	passwordFromStdin bool
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to system",
	RunE:  loginCmdF,
}

func init() {
	LoginCmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username or email")
	LoginCmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password")
	LoginCmd.Flags().BoolVar(&passwordFromStdin, "password-stdin", false, "Read password from stdin")
	RootCmd.AddCommand(LoginCmd)
}

func loginCmdF(command *cobra.Command, args []string) error {
	client, err := InitAnonClientCommandContextCobra(command)
	if err != nil {
		return err
	}
	fmt.Printf("Log in on %s\n", client.BaseURL)

	email := usernameFlag
	if email == "" {
		prompt := promptui.Prompt{
			Label:    "Email",
			Validate: validateEmail,
		}
		email, err = prompt.Run()
		if err != nil {
			return fmt.Errorf("error entering email: %v", err)
		}
	}

	var password string
	switch {
	case passwordFromStdin:
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read password from stdin: %v", err)
		}
		password = strings.TrimSpace(input)
	case passwordFlag != "":
		password = passwordFlag
	default:
		prompt := promptui.Prompt{
			Label:    "Password",
			Validate: validatePassword,
		}
		password, err = prompt.Run()
		if err != nil {
			return fmt.Errorf("error entering password: %v", err)
		}
	}

	user, authResponse, err := client.Login(email, password)
	if err != nil {
		return err
	}

	if authResponse.TwoFactorRequired {
		fmt.Println("A verification code has been sent to your email.")
		prompt := promptui.Prompt{
			Label:    "Enter 6-digit code",
			Validate: validatePinCode,
		}
		pinCode, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("error entering pin code: %v", err)
		}
		user, authResponse, err = client.TwoFactorVerify(pinCode, authResponse.TwoFactorToken)
		if err != nil {
			return err
		}
	}

	authRC := api.NewAuthRC(authResponse)
	err = api.SaveAuthRC(client.BaseURL, authRC)
	if err != nil {
		return fmt.Errorf("failed to save token to file: %v", err)
	}

	fmt.Printf("Welcome, %s\n", user.Username)
	return nil
}

func validateEmail(input string) error {
	if len(input) == 0 {
		return errors.New("email cannot be empty")
	}
	return nil
}

func validatePassword(input string) error {
	if len(input) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	return nil
}

func validatePinCode(input string) error {
	if len(input) != 6 {
		return errors.New("code must be exactly 6 digits")
	}
	for _, ch := range input {
		if ch < '0' || ch > '9' {
			return errors.New("code must contain only digits")
		}
	}
	return nil
}
