package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	_ "syscall"
)

type LPass struct {
	Username string
}

func (self *LPass) Exec(args []string) (*exec.Cmd, error) {
	// TODO: cache or otherwise remember this lookup?
	binaryPath, err := exec.LookPath("lpass")

	if err != nil {
		// TODO: log / output the error
		log.Fatal(fmt.Sprintf("Lpass: Error: unable to find the lpass binary: %s\n", err.Error()))
		return nil, err
	}

	log.Println(fmt.Sprintf("LPass.Exec: found lpass binary at %s\n", binaryPath))
	log.Println(fmt.Sprintf("LPass.Exec: executing lpass with args=%q\n", args))

	childProcess := exec.Command(binaryPath, args...)
	return childProcess, nil
}

func (self *LPass) Help(args []string) (*exec.Cmd, error) {
	childProc, err := self.Exec([]string{"help"})
	if err != nil {
		log.Fatal(fmt.Sprintf("Lpass: Error: executing help returned an error: %s\n", err.Error()))
		return nil, err
	}

	output, err := childProc.CombinedOutput()

	fmt.Fprintf(os.Stderr, "Output:\n")
	fmt.Fprintf(os.Stdout, "%s\n", output)
	// NB: unfortunately returns a 1 from help, which we'll need to disregard
	// if err != nil {
	// 	log.Fatal(fmt.Sprintf("Lpass: Error: getting output from help returned an error: %s\n", err.Error()))
	// 	return nil, err
	// }

	return childProc, nil
}

func (self *LPass) Login(args []string) (*exec.Cmd, error) {
	childProc, err := self.Exec(append([]string{"login", "--trust", self.Username}))
	if err != nil {
		log.Fatal(fmt.Sprintf("Lpass: Error: executing help returned an error: %s\n", err.Error()))
		return nil, err
	}
	output, err := childProc.CombinedOutput()

	fmt.Fprintf(os.Stderr, "Output:\n")
	fmt.Fprintf(os.Stdout, "%s\n", output)

	return childProc, nil
}

func (self *LPass) List(args []string) (*exec.Cmd, error) {
	// ls --format=""
	childProc, err := self.Exec(append([]string{"ls", "--format=%/ai	%/an	%/aN	%/au	%/ap	%/am	%/aU	%/as	%/ag"}))
	if err != nil {
		log.Fatal(fmt.Sprintf("Lpass: Error: executing help returned an error: %s\n", err.Error()))
		return nil, err
	}
	output, err := childProc.CombinedOutput()

	fmt.Fprintf(os.Stderr, "Output:\n")
	fmt.Fprintf(os.Stdout, "%s\n", output)

	return childProc, nil
}

func main() {
	// TODO: allow override with config file, override with ENV var, override with cli switch, default to env var for now
	lpass := &LPass{
		Username: os.Getenv("LPASSUSER"),
	}

	app := cli.NewApp()
	// TOOD: command processor (list, find, sync-pull sync-push)
	app.Action = func(c *cli.Context) error {
		if len(c.Args()) < 1 {
			lpass.Help([]string{})
			return nil
		}

		cmd := c.Args().Get(0)
		fmt.Printf("args: %q cmd=%s\n", c.Args(), cmd)

		switch cmd {
		case "help":
			lpass.Help(c.Args()[1:])
		case "login":
			fmt.Printf("is login\n")
			lpass.Login(c.Args()[1:])
		case "list":
			lpass.List(c.Args()[1:])
		case "ls":
			lpass.List(c.Args()[1:])
		default:
			fmt.Printf("unrecognized command: '%s'\n", cmd)
			lpass.Help([]string{})
		}
		return nil
	}
	app.Run(os.Args)

	// lpass.Login()
	// lpass.Help()
}
