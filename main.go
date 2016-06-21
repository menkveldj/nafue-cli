package main

import (
	_ "github.com/joho/godotenv/autoload"
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"github.com/menkveldj/nafue"
	"log"
	"os"
	"syscall"
	"github.com/menkveldj/nafue/config"
)

func main() {
	// setup env as needed
	nafue.Init(getConfig())

	app := cli.NewApp()
	app.Name = "Nafue"
	app.Usage = "Anonymous, secure file transfers that self destruct after first use or 24 hours using client side encryption."
	app.Commands = []cli.Command{
		{
			Name:  "get",
			Usage: "get [file]",
			Action: getFile,
		},
		{
			Name:  "share",
			Usage: "share [file]",
			Action: shareFile,
		},
	}
	app.Action = func(c *cli.Context) error {
		fmt.Println("Please run with a sub-command. For more information try \"nafue help\"")
		return errors.New("This is an error")
	}

	app.Run(os.Args)
}

func getFile(c *cli.Context) error {
	// verify url exists
	url := c.Args().First()
	if url == "" {
		fmt.Printf("You must enter a url\n")
		os.Exit(0)
	}

	// get file from url
	fileHeader, secureFile, err := nafue.GetFile(url)
	if err != nil {
		fmt.Printf("File never existed or was deleted.\n")
		os.Exit(0)
	}

	// tryUnseal func
	var fileUri string
	attemptUnseal := func() error {
		pass, err := promptPassword()
		if err != nil {
			fmt.Printf("Unable to decrypt file.\n")
			os.Exit(0)
		}
		fileUri, err = nafue.UnsealFile(secureFile, pass, fileHeader)
		if err != nil {
			return err
		}

		return nil
	}
	attempts := 0
	for {
		// after 3 tries delete everything and quite
		if attempts >= 3 {
			fmt.Println("To many attempts. Deleting all temporary data.")
			os.Remove(secureFile)
			os.Exit(0)
		}

		// try to do unseal
		err = attemptUnseal()
		if err != nil && err != nafue.C_DECRYPT_UNAUTHENTICATED {
			fmt.Printf("Error decrypting file.\n", err)
			os.Exit(0)
		} else if err == nafue.C_DECRYPT_UNAUTHENTICATED {
			fmt.Println("Couldn't Authorize Data. Try entering your password again.")
			attempts++
		} else {
			break
		}

	}
	fmt.Println("FileUri: ", fileUri)
	return nil

}

func shareFile(c *cli.Context) error {
	// get file handle to seal
	fileUri := c.Args().First()
	if fileUri == "" {
		log.Println("You must enter a file")
		os.Exit(0)
	}

	var pass string
	var err error
	for pass, err = promptPassword(); err != nil; {
		fmt.Printf("Can't Read Password: %s\n", err.Error())
	}
	shareUrl, err := nafue.SealShareFile(fileUri, pass)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Share Link: ", shareUrl)
	return nil
}

func promptPassword() (string, error) {
	// ask for password
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return "", err
	}

	password := string(bytePassword)
	fmt.Println()
	return password, nil
}

func getConfig() config.Config {
	env := os.Getenv("NAFUE_ENV")
	switch env {

	case "development":
		return config.Development()
	case "local":
		return config.Local()
	default:
		return config.Production()
	}
}