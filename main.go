package main

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type LPass struct {
	Username string
	Cachedir string
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
	// TODO: add args into the cached file name (even if we sha everything)
	// TODO: need support for turning this off & on
	var childProc *exec.Cmd = nil
	var response []byte
	var found bool
	var err error
	response, found = self.cacheGet("List.dat")

	if !found {
		childProc, err = self.Exec(append([]string{"ls", "--format=%/ai	%/an	%/aN	%/au	%/ap	%/am	%/aU	%/as	%/ag"}))
		if err != nil {
			log.Fatal(fmt.Sprintf("Lpass: Error: executing help returned an error: %s\n", err.Error()))
			return nil, err
		}
		response, err = childProc.CombinedOutput()
	}

	fmt.Fprintf(os.Stderr, "Output:\n")
	fmt.Fprintf(os.Stdout, "%s\n", response)

	self.cachePut("List.dat", string(response))

	return childProc, nil
}

func defaultUserName() string {
	uname := os.Getenv("LPASSUSER")
	if uname != "" {
		return uname
	}

	childProc := exec.Command("git", "config", "--get", "user.email")
	output, err := childProc.CombinedOutput()
	if err == nil {
		parts := strings.SplitN(string(output), "\n", 2)
		return parts[0]
	}

	// ok, can't get it from git
	return ""
}

func FileExists(path string) bool {
	// http://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
	_, err := os.Stat(path)

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	// return true
	return false
}

func DirExists(dname string) bool {
	return FileExists(dname)
}

func (self *LPass) cacheGet(key string) ([]byte, bool) {
	cfile := path.Join(self.Cachedir, key)

	if FileExists(cfile) {
		bytes, err := ioutil.ReadFile(cfile)
		if err != nil {
			log.Fatalf("Error reading cache file: %s for key %s : %s", cfile, key, err)
		}
		return bytes, true
	}

	return []byte{}, false
}

func (self *LPass) cachePut(key, value string) {
	cfile := path.Join(self.Cachedir, key)
	err := ioutil.WriteFile(cfile, []byte(value), 0600)
	if err != nil {
		log.Fatalf("Error writing cache file: %s for key %s : %s", cfile, key, err)
	}
}

func main() {
	// TODO: allow override with config file, override with ENV var, override with cli switch, default to env var for now
	lpass := &LPass{
		Username: defaultUserName(),
		Cachedir: "",
	}

	app := cli.NewApp()
	app.Name = "rlpass"
	app.Usage = "Wrapper around lpass cli tooling"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "username",
			Value: defaultUserName(),
			Usage: "Your LastPass Login name (probably your email address)",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "cachedir",
			Value: "./.rlpass/cache",
			Usage: "Local cache directory",
		},
	}

	// TOOD: command processor (list, find, sync-pull sync-push)
	app.Action = func(c *cli.Context) error {
		lpass.Username = c.String("username")
		lpass.Cachedir = c.String("cachedir")

		log.Printf("app.Action: DirExists(%s) => %q", lpass.Cachedir, DirExists(lpass.Cachedir))

		if !DirExists(lpass.Cachedir) {
			log.Printf("app.Action: creating: %s", lpass.Cachedir)
			err := os.MkdirAll(lpass.Cachedir, 0700)
			log.Printf("app.Action: created: dir=%s : err=%s", lpass.Cachedir, err)
			if err != nil {
				log.Fatalf("Error creating dir=%s : err=%s", lpass.Cachedir, err)
			}
		}

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
