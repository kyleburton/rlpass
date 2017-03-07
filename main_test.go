package main

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestLPassEntryParsing(t *testing.T) {
	s1 := "1065230732160897586/\tGenerated Password for some.where/\t(none)/Generated Password for some.where/\tvpq#0f7=wj:o)$:ch9egq*/\t2016-05-23 18:12/\t2016-05-23 15:12/"
	e1 := ParseLPassEntry(s1)

	if e1.AccountId != "1065230732160897586" {
		t.Error(fmt.Sprintf("Error parsing LPassEntry, expected AccountId='%s', got '%s' from '%s'",
			"1065230732160897586",
			e1.AccountId,
		))
	}

	s2 := e1.ToString()

	if s1 != s2 {
		t.Error(fmt.Sprintf("Error 'round tripping' LPassEntry, expected '%s', got '%s'",
			s1,
			s2,
		))
	}
}

func TestParseLPassList(t *testing.T) {
	s1 := `2153813569506231076259/	Generated Password for some.where/	(none)/Generated Password for some.where/		f8w016ehv3vlzedh/	2016-05-23 18:12/	2016-05-23 15:12/
1221777143672919814263/	Generated Password for some.where/	(none)/Generated Password for some.where/		yrct59c4tirp1rj0mbzmlouoh/	2016-04-08 16:51/	2016-04-08 13:51/
3172274455914884164282/	account.mojang.com/	(none)/account.mojang.com/	me@some.where/	l0ew0i1fkhxas5s9yf8n8z5v0v2l/	2016-03-11 00:57/
1555133843826238512084/	accounts.google.com/	(none)/accounts.google.com/	me@some.where/	u7fwtkos1wp75hueez5e/	2017-02-15 23:00/
6171571597827587571532/	github.com/	(none)/github.com/	me@some.where/	i0who8k5tvvxfu6c7qvvw7/	2016-03-11 00:57/`

	ents := ParseLPassList(s1)

	/*
		for _, ent := range ents {
			fmt.Printf("ent=%+v\n", ent)
		}
	*/

	if len(ents) != 5 {
		t.Error(fmt.Sprintf("Error ParseLPassList, expected 5 results, got %d",
			len(ents),
		))
	}

}

func TestParseShowFirstLine(t *testing.T) {
	s1 := `(none)/tivo.com [id: 5926414273882541009]`
	ent, err := ParseShowFirstLine(s1)
	if err != nil {
		t.Error(err)
		return
	}

	if ent == nil {
		t.Error("Error, ParseShowFirstLine failed, returned nil")
		return
	}

	if ent.AccountName != "tivo.com" {
		t.Error(fmt.Sprintf("Error: expected ent.AccountId to be %s, got '%s'",
			"tivo.com",
			ent.AccountName,
		))
	}
	if ent.AccountId != "5926414273882541009" {
		t.Error(fmt.Sprintf("Error: expected ent.AccountId to be %s, got '%s'",
			"5926414273882541009",
			ent.AccountId,
		))
	}
}

func TestParseShowFirstLineNoFolder(t *testing.T) {
	s1 := `some site with no folder [id: 2921956165454513309301]`
	e1 := "some site with no folder"

	ent, err := ParseShowFirstLine(s1)
	if err != nil {
		t.Error("Error, ParseShowFirstLine failed, returned nil")
		return
	}

	if ent.AccountName != e1 {
		t.Errorf("Error: expected AccountName='%s' from '%s', got:'%a'",
			e1,
			s1,
			ent,
		)
	}

}

func TestParseShow(t *testing.T) {
	s1 := `(none)/tivo.com [id: 5926414273882541009]
	Username: me@some.where
	Password: 1402931102206281341285
	URL: https://www.tivo.com
	cams_cb_username: me@some.where
	cams_cb_password: 1402931102206281341285
	remember_email: Checked`

	note, err := ParseShow(s1)

	if err != nil {
		t.Error(err)
		return
	}

	if note == nil {
		t.Error(fmt.Sprintf("Error: parse failed with nil?"))
		return
	}

	if val, ok := note.Properties["remember_email"]; ok && val != "Checked" {
		t.Error(fmt.Sprintf("Error: expected remember_email to be Checked, was '%s'",
			note.Properties["remember_email"],
		))
	}

	if note.Credential == nil || note.Credential.Username != "me@some.where" {
		t.Error(fmt.Sprintf("Error: expected note.Credential.Username to be me@some.where, was '%s'",
			note.Properties["remember_email"],
		))
	}
}

func TestParseShowWithNote(t *testing.T) {
	n1 := `{"note": "this will be json to and from",
"field": "value on a second line",
"thing": "on a third line"}`

	s1 := `Test/Test Note for Notes [id: 4281390154665116890]
URL: http://sn
Notes: ` + n1

	note, err := ParseShow(s1)

	if err != nil {
		t.Error(err)
		return
	}

	if note == nil {
		t.Error(fmt.Sprintf("Error: parse failed with nil?"))
		return
	}

	if note.RawNotes == "" {
		t.Error(fmt.Sprintf("Error: failed to parse note!"))
		return
	}

	if note.RawNotes != n1 {
		t.Error(fmt.Sprintf("Error: expected note:'%s', got:'%s'", n1, note.RawNotes))
		return
	}

	if note.GetString("note") != "this will be json to and from" {
		t.Error(fmt.Sprintf("Error: expected note.GetString(\"note\") to be '%s', got '%s'",
			"this will be json to and from",
			note.GetString("note"),
		))
		return
	}

}

// some entropy for your needs
/*
!!bash<cr>
:exec '!'.getline('.')
for x in $(seq 100); do echo -n $RANDOM; done | fold -w 22  | head -n -1
1004837402136515081110
3829049298952452913804
3419973201164235313820
3214692296825854481098
3851999531282311833231
4320230163371653332629
2785512815772223533133
1429201368300061342548
6114366259101606370311
9750188012841132464856
2872443223058229194998
1160574473190015831536
8073266721675423329240
0725661781648301480118
8822875345567225535814
2981998125642284791856
2970212852089724074555

*/

func TestScrubPathOfSpecialCharacters(t *testing.T) {
	s1 := "kyle.burton@gmail.com/somesite.com"
	e1 := "kyle.burton-gmail.com/somesite.com"
	a1 := ScrubPathOfSpecialCharacters(s1)
	if e1 != a1 {
		t.Error(fmt.Sprintf("Error: expected ScrubPathOfSpecialCharacters(%s) to be '%s', got '%s'",
			s1, e1, a1,
		))
		return
	}
}

func TestSecureNoteToPath(t *testing.T) {
	s1 := `(streaming-services)/tivo.com [id: 5926414273882541009]
	Username: me@some.where
	Password: 1402931102206281341285
	URL: https://www.tivo.com
	cams_cb_username: me@some.where
	cams_cb_password: 1402931102206281341285
	remember_email: Checked`

	note, err := ParseShow(s1)

	if err != nil {
		t.Error(err)
		return
	}

	if note == nil {
		t.Error(fmt.Sprintf("Error: parse failed with nil?"))
		return
	}

	p := note.EntryInfo.ToPath("credentials")
	expected_path := "credentials/-streaming-services-/tivo.com/credential.json"

	if p != expected_path {
		t.Error(fmt.Sprintf("Error: EntryInfo.ToPath: expected '%s', got '%s' from '%s'",
			expected_path,
			p,
			note.EntryInfo.AccountNameIncludingPath,
		))
		return
	}
}

func TestParseShowWithCertificate(t *testing.T) {
	fixture_file := "fixtures/show/note-with-certificate.out"
	data, err := ioutil.ReadFile(fixture_file)

	if err != nil {
		t.Errorf("Error reading fixture file '%s' : %s", fixture_file, err)
		return
	}

	note, err := ParseShow(string(data))

	if err != nil {
		t.Error(err)
		return
	}

	if note == nil {
		t.Error(fmt.Sprintf("Error: parse failed with nil?"))
		return
	}

	val, ok := note.Properties["Certificate"]
	if !ok {
		t.Error(fmt.Sprintf("Error: did not parse Certificate out of '%s'", fixture_file))
	}

	if len(val) < 1 {
		t.Errorf("Error: Certificate (expected len>0, was %d) not parsed out of '%s'",
			fixture_file,
			len(val),
		)
	}
}
