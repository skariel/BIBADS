// LICENSE: do whatever you want with this, use on your own risk
// AUTHOR: Ariel Keselman
// VERSION: 1.0.0 2016
//
// This program generates a bib file to be used with bibtex
// The name of the generated bib file is as needed by the tex file,
// for e.g. if the tex file contains `\bibliography{KN16.bib}` the
// generated file name will be `KN16.bib`
//
// The generated file will contain all the citations needed by the
// tex file that can be found in the Nasa ads database
//
// The program looks by bibcodes. Hence all the citations in the
// tex file need to reference bibcodes, e.g. `\cite{2009MNRAS.399..683J}`
//
// Aliases are optional, and are defined inside a comment in the tex file,
// for e.g.:
// % bibalias colles2DF 2001MNRAS.328.1039C
// % bibalias peeb80    1980lssu.book.....P
// Then the citation can look like `\cite{peeb80}`
//
// TO RUN THE PROGRAM:
//
// as a script, need a go installation (google golang):
// > go run bibads.go [your tex file here]
//
// compiling, this generates a self contained executable:
// > go build bibads.go
//
// then run:
// ./bibads [your tex file here]
//

package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func getBibCodeAliasesFromSource(fileName string) (bibCodeAliases map[string]string, err error) {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	fileStr := string(fileBytes)
	sep := strings.Split(fileStr, "\n")
	bibCodeAliases = make(map[string]string)
	for _, line := range sep {
		lineFields := strings.Fields(line)
		if len(lineFields) < 3 {
			continue
		}
		if strings.Index(lineFields[0], "%") != 0 {
			continue
		}
		if lineFields[0] == "%" && len(lineFields) < 4 {
			continue
		}
		if lineFields[0] == "%" && lineFields[1] != "bibalias" {
			continue
		} else if lineFields[1] != "bibalias" && lineFields[0] != "%bibalias" {
			continue
		}
		keyIndex := 2
		if lineFields[0] == "%bibalias" {
			keyIndex--
		}
		key := lineFields[keyIndex]
		val := lineFields[keyIndex+1]
		bibCodeAliases[key] = val
	}
	return bibCodeAliases, nil
}

func getBibFileNameAndBibCodesFromSource(fileName string) (bibFileName string, bibCodes map[string]bool, err error) {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", nil, err
	}
	fileStr := string(fileBytes)
	sep := strings.Split(fileStr, "\\bibliography{")
	if len(sep) < 2 {
		return "", nil, errors.New("No bib file name found in " + fileName)
	}
	bibFileName = strings.Split(sep[1], "}")[0]
	bibCodes = make(map[string]bool)
	sep = strings.Split(fileStr, "\\cite{")
	for _, cl := range sep[1:] {
		cl = strings.Split(cl, "}")[0]
		for _, c := range strings.Split(cl, ",") {
			bibCodes[strings.TrimSpace(c)] = true
		}
	}
	sep = strings.Split(fileStr, "\\citep{")
	for _, cl := range sep[1:] {
		cl = strings.Split(cl, "}")[0]
		for _, c := range strings.Split(cl, ",") {
			bibCodes[strings.TrimSpace(c)] = true
		}
	}
	return bibFileName, bibCodes, nil
}

func padRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

func getBibRef(bibCode string) (bibRef string, err error) {
	const querystr = "http://adsabs.harvard.edu/cgi-bin/nph-bib_query?data_type=BIBTEX&bibcode="
	response, err := http.Get(querystr + bibCode)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", errors.New(response.Status)
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	bodyStr := string(bodyBytes)
	bibRef = "@" + strings.Split(bodyStr, "@")[1]
	return bibRef, nil
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		println("please supply name of file to operate on")
		os.Exit(1)
	}
	if flag.NArg() > 1 {
		println("please supply only a single name of file to oprate on (a single non flag argument)")
		os.Exit(1)
	}

	texFileName := flag.Args()[0]
	bibCodeAliases, err := getBibCodeAliasesFromSource(texFileName)
	if err != nil {
		panic(err)
	}
	bibFileName, codes, err := getBibFileNameAndBibCodesFromSource(texFileName)
	if err != nil {
		panic(err)
	}
	println("bib file name: ", bibFileName)

	bibFileText := ""
	for code := range codes {
		realCode, isAlias := bibCodeAliases[code]
		var alias string
		if isAlias {
			alias = code
			code = realCode
		}
		print(padRight(code, " ", 19), "   ", padRight(alias, " ", 15), "   ...   ")
		bibRefText, err := getBibRef(code)
		if err != nil {
			println(err.Error())
			continue
		}
		println("OK")
		if isAlias {
			bibRefText = strings.Replace(bibRefText, realCode, alias, 1)
		}
		bibFileText = bibFileText + "\n\n" + bibRefText
	}

	ioutil.WriteFile(bibFileName, []byte(bibFileText), 0644)
}
