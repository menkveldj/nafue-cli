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
	"path/filepath"
	"syscall"
	"github.com/menkveldj/nafue/config"
	"github.com/menkveldj/nafue-cli/utility"
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
		println("Please run with a sub-command. For more information try \"nafue help\"")
		return errors.New("This is an error")
	}

	app.Run(os.Args)
}

func getFile(c *cli.Context) error {

	// verify url exists
	url := c.Args().First()
	if url == "" {
		fmt.Printf("You must enter a url")
		os.Exit(0)
	}

	// get temp file
	secureData := utility.CreateTempFile()
	defer secureData.Close()

	// get status about file
	fileInfo, err := secureData.Stat()
	if err != nil {
		logError(err)
		return err
	}
	//defer utility.DeleteTempFile(secureData.Name())

	// get file from url
	fileHeader, err := nafue.GetFile(url, secureData)
	if err != nil {
		fmt.Printf("File never existed or was deleted.\n")
		os.Exit(0)
	}

	// tryUnseal func
	attemptUnseal := func() error{
		pass, err := promptPassword()
		if err != nil {
			fmt.Printf("Unable to decrypt file.\n")
			os.Exit(0)
		}
		err = nafue.UnsealFile(secureData, pass, fileHeader, fileInfo)
		if err != nil {
			return err
		}

		return nil
	}
	attemptUnseal()
	//
	//// do decrypt
	//var r io.Reader
	//var name string
	//i := 0
	//for r, name, err = decrypt(); err != nil; i++ {
	//	fmt.Printf("%s\n", err)
	//	if i == 3 {
	//		fmt.Printf("To many failed attempts. File was deleted.\n")
	//		os.Exit(0)
	//	}
	//}
	//out, err := os.Create(name)
	//if err != nil {
	//	panic(err)
	//}
	//io.Copy(out, r)
	//fmt.Printf("File saved to: %s\n", name)

	return nil
}

func shareFile(c *cli.Context) error {
	// get file handle to seal
	file := c.Args().First()
	if file == "" {
		log.Println("You must enter a file")
		os.Exit(0)
	}

	// open file for reading
	f, err := os.Open(file)
	defer f.Close()

	if err != nil {
		logError(err)
		return err
	}

	// get status about file
	fileInfo, err := f.Stat()
	if err != nil {
		logError(err)
		return err
	}

	// open file handle for writing
	sf := utility.CreateTempFile()
	defer sf.Close()
	defer utility.DeleteTempFile(sf.Name())

	var pass string
	for pass, err = promptPassword(); err != nil; {
		fmt.Printf("Can't Read Password: %s\n", err.Error())
	}
	shareUrl, err := nafue.SealShareFile(f, sf, fileInfo, filepath.Base(file), pass)
	if err != nil {
		logError(err)
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

func logError(err error) {
	fmt.Printf("%s\n", err)
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