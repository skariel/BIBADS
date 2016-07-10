package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func getCacheFromBibFile(fileName string) (cache map[string]string, err error) {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	fileStr := string(fileBytes)
	sep := strings.Split(fileStr, "@")
	cache = make(map[string]string)
	for _, entry := range sep[1:] {
		entry = strings.TrimSpace(entry)
		code := strings.Split(strings.Split(entry, "{")[1], ",")[0]
		cache[code] = "@" + entry
	}
	return cache, nil
}

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
	sep = strings.Split(fileStr, "\\cite")
	for _, cl := range sep[1:] {
		cl = strings.Split(cl, "{")[1]
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

func getAliasedCachedBibText(code string, bibCodeAliases map[string]string, cache map[string]string, out chan string) {
	realCode, isAlias := bibCodeAliases[code]
	var alias string
	if isAlias {
		alias = code
		code = realCode
	}

	msg := padRight(code, " ", 19) + "   " + padRight(alias, " ", 15) + "   ...   "
	isInCache := false
	bibRefText := ""
	var err error
	if cache != nil {
		if isAlias {
			bibRefText, isInCache = cache[alias]
		} else {
			bibRefText, isInCache = cache[code]
		}
	}
	if !isInCache {
		bibRefText, err = getBibRef(code)
		if err != nil {
			println(msg, err.Error())
			out <- ""
			return
		}
		println(msg, "OK")
	} else {
		println(msg, "OK (cached)")
	}
	if isAlias {
		bibRefText = strings.Replace(bibRefText, "{"+realCode, "{"+alias, 1)
	}
	out <- bibRefText
}

func main() {
	noCachePtr := flag.Bool("nocache", false, "force fetch data from ads even if present in current bib file")
	flag.Parse()

	if flag.NArg() == 0 {
		println("please supply name of file to operate on. Add a -nocache flag to not use existing entries")
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

	var cache map[string]string
	println("nocache=", *noCachePtr)
	if !(*noCachePtr) {
		cache, _ = getCacheFromBibFile(bibFileName)
	}

	bibFileText := ""
	c := make(chan string)
	for code := range codes {
		go getAliasedCachedBibText(code, bibCodeAliases, cache, c)
	}
	for _ = range codes {
		bibText := <-c
		bibFileText += "\n\n" + bibText
	}

	ioutil.WriteFile(bibFileName, []byte(bibFileText), 0644)
}
