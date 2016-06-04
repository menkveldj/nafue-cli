package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/menkveldj/nafue"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

// todo add delete temp function
func main() {
	// setup env as needed
	nafue.Init()

	app := cli.NewApp()
	app.Name = "Nafue"
	app.Usage = "Anonymous, secure file transfers that self destruct after first use or 24 hours using client side encryption."
	app.Commands = []cli.Command{
		{
			Name:  "get",
			Usage: "get [file]",
			Action: func(c *cli.Context) error {
				url := c.Args().First()
				if url == "" {
					log.Println("You must enter a url")
					os.Exit(1)
				}
				body, header := nafue.TryGetURL(url)

				decrypt := func() (io.Reader, string, error) {
					pass, err := promptPassword()
					if err != nil {
						return bytes.NewBufferString(""), "", err
					}
					return nafue.TryDecrypt(body, header, pass)
				}
				var r io.Reader
				var name string
				var err error
				i := 0
				for r, name, err = decrypt(); err != nil; i++ {
					fmt.Errorf("%s\n", err)
					if i == 3 {
						fmt.Errorf("To many failed attempts. File was deleted.\n")
						os.Exit(1)
					}
				}
				out, err := os.Create(name)
				if err != nil {
					panic(err)
				}
				io.Copy(out, r)
				fmt.Printf("File saved to: %s\n", name)

				return nil
			},
		},
		{
			Name:  "share",
			Usage: "share [file]",
			Action: func(c *cli.Context) error {
				file := c.Args().First()
				if file == "" {
					log.Println("You must enter a file")
					os.Exit(1)
				}
				f, err := os.Open(file)

				if err != nil {
					logError(err)
					return err
				}
				fstat, err := f.Stat()
				if err != nil {
					logError(err)
					return err
				}
				// share file
				var pass string
				for pass, err = promptPassword(); err != nil; {
					fmt.Errorf("Encountered Exception: %s\n", err.Error())
				}
				shareURL := nafue.PutReader(f, fstat.Size(), filepath.Base(file), pass)
				fmt.Println("Share Link: ", shareURL)
				return nil
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		println("Please run with a sub-command. For more information try \"nafue help\"")
		return errors.New("This is an error")
	}

	app.Run(os.Args)
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
	fmt.Errorf("%s\n", err)
}

//salt:  [823823226, 793384892]
//salt b64:  MRqLei9KF7w=
//iv:  [-1726315756, -750854000, -575834770, -1005289789, -875997976, 1972462144, 348864951, -1912938221]
//iv b64:  mRqDFNM+4JDdrXVuxBR+w8vJVOh1kWJAFMtBt4364RM=
