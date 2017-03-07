package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"syscall"
)

type LPass struct {
	Username string
	Cachedir string
}

type LPassEntry struct {
	AccountId                string `LPassListFormat:"%ai",json:"id"`
	AccountName              string `LPassListFormat:"%an",json:"name"`
	AccountNameIncludingPath string `LPassListFormat:"%aN",json:"path"`
	AccountUser              string `LPassListFormat:"%au",json:"user"`
	AccountPassword          string `LPassListFormat:"%ap",json:"password"`
	AccountModificationTime  string `LPassListFormat:"%am",json:"mtime"`
	AccountLastTouchTime     string `LPassListFormat:"%aU",json:"atime"`
	AccountShareName         string `LPassListFormat:"%as",json:"share-name"`
	AccountGroupName         string `LPassListFormat:"%ag",json:"group-name"`
	// NB: not sure we're going to use these
	// FieldName                `LPassListFormat:"%fn",json:"field-name"`
	// FieldValue               `LPassListFormat:"%fv",json:"field-value"`
}

type StandardCredential struct {
	Name          string
	Owner         string
	Description   string
	IssuedAt      string
	IssuedBy      string
	IssuedTo      string
	ExpiresAt     string
	LastRotatedAt string
	Usage         string
	Help          string
	ProjectUrl    string
	Username      string
	Password      string
	Credential    string
	Url           string
}

type NoteObject interface{}

type LPassSecureNote struct {
	EntryInfo  *LPassEntry
	Properties map[string]string
	Credential *StandardCredential
	Notes      NoteObject
	RawNotes   string
}

// Get             => pull a *Note
// GetString       => pull a string
// GetInt          => pull a int
// GetArray        => pull an array []*Note
// GetIntArray     => pull an array []string
// GetStringArray  => pull an array []string
// func (self *NoteObject) Get(field string) *NoteObject {
//   m := map[string]interface{}(self)
//   return m[field]
// }

func (self *LPassSecureNote) GetString(k string) string {
	// return string(map[string]string(self.Notes)[k])
	v := self.Notes.(map[string]interface{})[k]
	return v.(string)
}

func (self *LPass) Exec(args []string) (*exec.Cmd, error) {
	// TODO: cache or otherwise remember this lookup?
	binaryPath, err := exec.LookPath("lpass")

	if err != nil {
		// TODO: log / output the error
		log.Fatal(fmt.Sprintf("LPass: Error: unable to find the lpass binary: %s\n", err.Error()))
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
		log.Fatal(fmt.Sprintf("LPass: Error: executing help returned an error: %s\n", err.Error()))
		return nil, err
	}

	output, err := childProc.CombinedOutput()

	fmt.Fprintf(os.Stderr, "Output:\n")
	fmt.Fprintf(os.Stdout, "%s\n", output)
	// NB: unfortunately lpass returns a 1 from help, which we'll need to disregard
	// if err != nil {
	// 	log.Fatal(fmt.Sprintf("LPass: Error: getting output from help returned an error: %s\n", err.Error()))
	// 	return nil, err
	// }

	return childProc, nil
}

// NB: since it uses a password reader we probably have to do an exec
func (self *LPass) Login(args []string) (*exec.Cmd, error) {
	if self.Username == "" {
		panic("Error: you have to set your lastpass username!")
	}

	binaryPath, err := exec.LookPath("lpass")
	if err != nil {
		panic(err)
	}

	argv := []string{binaryPath, "login", "--trust", self.Username}
	env := os.Environ()
	fmt.Printf("Executing: %s\n", argv)
	err = syscall.Exec(binaryPath, argv, env)
	if err != nil {
		panic(err)
	}

	panic("WHOAH, it shouldn't be possible to get here!")

	return nil, nil
}

func ParseLPassEntry(s string) *LPassEntry {
	ent := &LPassEntry{}
	ent.Parse(s)
	return ent
}

func (self *LPassEntry) Parse(line string) *LPassEntry {
	parts := strings.SplitN(line, "\t", 9)

	if len(parts) < 6 {
		panic(fmt.Sprintf("Error: expected at least 6 parts, got %d from '%s'",
			len(parts),
			line))
	}

	cleanedParts := make([]string, len(parts))

	for idx, s := range parts {
		cleanedParts[idx] = strings.TrimSuffix(s, "/")
	}

	self.AccountId = cleanedParts[0]
	self.AccountName = cleanedParts[1]
	self.AccountNameIncludingPath = cleanedParts[2]
	self.AccountUser = cleanedParts[3]
	self.AccountPassword = cleanedParts[4]
	if len(cleanedParts) > 5 {
		self.AccountModificationTime = cleanedParts[5]
	}
	if len(cleanedParts) > 6 {
		self.AccountLastTouchTime = cleanedParts[6]
	}
	if len(cleanedParts) > 7 {
		self.AccountShareName = cleanedParts[7]
	}
	if len(cleanedParts) > 8 {
		self.AccountGroupName = cleanedParts[8]
	}

	return self
}

func (self *LPassEntry) ToArray() []string {
	return []string{
		self.AccountId,
		self.AccountName,
		self.AccountNameIncludingPath,
		self.AccountUser,
		self.AccountPassword,
		self.AccountModificationTime,
		self.AccountLastTouchTime,
		self.AccountShareName,
		self.AccountGroupName,
	}
}

func (self *LPassEntry) ToString() string {
	vals := make([]string, 0)

	for _, s := range self.ToArray() {
		if s != "" {
			vals = append(vals, s+"/")
		}
	}

	return strings.Join(vals, "\t")
}

func (self *LPassEntry) ToJson() []byte {
	b, err := json.MarshalIndent(self, "", "  ")
	if err != nil {
		panic(err)
	}

	return b
}

func (self *LPassEntry) ToPath(prefix string) string {
	reg, err := regexp.Compile("[^\\.-_/A-Za-z0-9]")
	if err != nil {
		panic(err)
	}

	pathed_name := reg.ReplaceAllString(self.AccountNameIncludingPath, "-")
	return path.Join(prefix, pathed_name, "credential.json")
}

func (self *LPassSecureNote) ToJson() []byte {
	b, err := json.MarshalIndent(self, "", "  ")
	if err != nil {
		panic(err)
	}

	return b
}

func ParseLPassList(s string) []*LPassEntry {
	lines := strings.Split(s, "\n")

	// fmt.Fprintf(os.Stderr, "ParseLPassList: got %d lines\n", len(lines))

	entries := make([]*LPassEntry, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}
		entries = append(entries, ParseLPassEntry(line))
	}

	return entries
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
		childProc, err = self.Exec(append([]string{"ls", "--format=%/ai\t%/an\t%/aN\t%/au\t%/ap\t%/am\t%/aU\t%/as\t%/ag"}))
		if err != nil {
			log.Fatal(fmt.Sprintf("LPass: Error: executing help returned an error: %s\n", err.Error()))
			return nil, err
		}
		response, err = childProc.CombinedOutput()
	}

	entries := ParseLPassList(string(response))
	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(b))

	self.cachePut("List.dat", string(response))

	return childProc, nil
}

func ParseShowFirstLine(s string) (*LPassEntry, error) {
	// `(none)/tivo.com [id: 5926414273882541009]`
	spos := strings.Index(s, "[")
	epos := strings.LastIndex(s, "]")

	if spos == -1 || epos == -1 {
		panic(fmt.Sprintf("Error: expected show's first line to have properties, it was: '%s'", s))
	}

	accountNameIncludingPath := strings.Trim(s[0:spos], " \t\r\n")
	lastSlashPos := strings.LastIndex(accountNameIncludingPath, "/")
	if lastSlashPos == -1 {
		panic(fmt.Sprintf("Error: expected firstline to have an accountNameIncludingPath, it was: '%s'", accountNameIncludingPath))
	}

	accountName := strings.Trim(s[lastSlashPos+1:spos-1], " \t\r\n")

	pairs := strings.Split(s[spos+1:epos], ", ")
	parts := make(map[string]string)

	for _, pair := range pairs {
		kv := strings.SplitN(pair, ": ", 2)
		parts[kv[0]] = strings.Trim(kv[1], " \t\r\n")
	}

	ent := &LPassEntry{
		AccountId:                parts["id"],
		AccountName:              accountName,
		AccountNameIncludingPath: accountNameIncludingPath,
	}

	return ent, nil
}

func ParseShow(s string) (*LPassSecureNote, error) {
	lines := strings.Split(s, "\n")

	if len(lines) < 1 {
		panic(fmt.Sprintf("Error: expected `lpass show` output to be multiple lines, got `%s`",
			s))
	}

	for idx, line := range lines {
		lines[idx] = strings.TrimLeft(line, " \t\n\r")
	}

	note := &LPassSecureNote{}
	ent, err := ParseShowFirstLine(lines[0])
	if err != nil {
		panic(err)
	}
	note.EntryInfo = ent

	note.Properties = make(map[string]string)

	for ii, line := range lines[1:] {
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, ": ", 2)

		// fmt.Fprintf(os.Stderr, "ParseShow: kv[%d] %s=%s\n", ii, kv[0], kv[1])

		if kv[0] == "Notes" {
			// the value and the rest of the lines are all the note, so we just stop here
			note.RawNotes = kv[1] + "\n" + strings.Join(lines[(ii+2):], "\n")
			note.RawNotes = strings.TrimSuffix(note.RawNotes, "\n ")
			break
		}

		if len(kv) != 2 {
			panic(fmt.Sprintf("Error parsing property, expected 2 fields, got %d from: '%s'", len(kv), line))
		}

		note.Properties[kv[0]] = strings.Trim(kv[1], " \t\r\n")
	}

	if note.RawNotes != "" {
		err = json.Unmarshal([]byte(note.RawNotes), &note.Notes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deserializing Notes json, see the contents of RawNotes: %s\n", err)
			json.Unmarshal([]byte("{}"), &note.Notes)
		}
	}

	note.EntryInfo = ent
	note.Credential = &StandardCredential{
		Username: note.Properties["Username"],
		Password: note.Properties["Password"],
		Url:      note.Properties["URL"],
	}

	return note, nil
}

func (self *LPass) Show(args []string) (*exec.Cmd, error) {
	// lpass show --color=never --all <<id>>
	var childProc *exec.Cmd = nil
	var response []byte
	var err error

	if len(args) != 1 {
		panic("Error: you must supply a ID")
	}

	childProc, err = self.Exec(append([]string{"show", "--color=never", "--all", args[0]}))
	if err != nil {
		log.Fatal(fmt.Sprintf("LPass: Error: executing help returned an error: %s\n", err.Error()))
		return nil, err
	}
	response, err = childProc.CombinedOutput()

	secureNote, err := ParseShow(string(response))

	if err != nil {
		return nil, err
	}

	fmt.Printf(string(secureNote.ToJson()))

	return nil, nil
}

func (self *LPass) Spec(args []string) (*exec.Cmd, error) {
	note := &LPassSecureNote{}
	note.EntryInfo = &LPassEntry{}
	note.Credential = &StandardCredential{}
	note.Properties = map[string]string{
		"_properties": "don't fill this in, it'll be ignored",
	}
	note.Notes = map[string]string{
		"_full-monty-here": "this should match your standard cred",
	}
	note.RawNotes = "this is ignored"
	data, err := json.MarshalIndent(note, "", "  ")

	if err != nil {
		panic(err)
	}

	fmt.Printf(string(data))
	fmt.Printf("\n")

	return nil, nil
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
		cli.StringFlag{
			Name:  "cachedir",
			Value: "./.rlpass/cache",
			Usage: "Local cache directory",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "help",
			Aliases: []string{"h"},
			Usage:   "show this help",
			Action: func(c *cli.Context) error {
				lpass.Help(c.Args())
				return nil
			},
		},
		{
			Name:    "login",
			Aliases: []string{"l"},
			Usage:   "shell out to lpass to login",
			Action: func(c *cli.Context) error {
				lpass.Login(c.Args())
				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "list your lastpass credentials, emits json",
			Action: func(c *cli.Context) error {
				lpass.List(c.Args())
				return nil
			},
		},
		{
			Name:    "show",
			Aliases: []string{"cat"},
			Usage:   "show a json formatted credential",
			Action: func(c *cli.Context) error {
				lpass.Show(c.Args())
				return nil
			},
		},
		{
			Name:    "spec",
			Aliases: []string{"cat"},
			Usage:   "json template for a secret note",
			Action: func(c *cli.Context) error {
				lpass.Spec(c.Args())
				return nil
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		lpass.Username = c.String("username")
		lpass.Cachedir = c.String("cachedir")

		if !DirExists(lpass.Cachedir) {
			log.Printf("app.Action: creating: %s", lpass.Cachedir)
			err := os.MkdirAll(lpass.Cachedir, 0700)
			log.Printf("app.Action: created: dir=%s : err=%s", lpass.Cachedir, err)
			if err != nil {
				log.Fatalf("Error creating dir=%s : err=%s", lpass.Cachedir, err)
			}
		}
		return nil
	}

	app.Run(os.Args)
}
