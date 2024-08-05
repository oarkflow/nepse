package csvtool

import (
	"bufio"
	. "fmt"
	"os"
	"regexp"
	. "strconv"
	"strings"

	"github.com/oarkflow/errors"
)

var flags Flags
var FPaths FilePaths
var messager chan string
var gotpass chan string
var passprompt chan bool
var saver chan saveData
var savedLine chan bool
var fileclick chan string
var browsersOpen = 0
var slash string
var printer Printer

// wrapper for CsvQuery
func runCsvQuery(query string, req *webQueryRequest) (SingleQueryResult, error) {
	q := QuerySpecs{QueryString: query}
	if (req.FileIO & F_CSV) != 0 {
		q.save = true
	}
	res, err := CsvQuery(&q)
	res.Query = query
	return res, err
}

// run webQueryRequest with multiple queries deliniated by semicolon
func runQueries(req *webQueryRequest) ([]SingleQueryResult, error) {
	query := req.Query
	// remove uneeded characters from end of string
	ending := regexp.MustCompile(`;\s*$`)
	query = ending.ReplaceAllString(query, ``)
	queries := strings.Split(strings.Replace(query, "\\n", "", -1), ";")
	req.Qamount = len(queries)
	// send info to realtime saver
	if (req.FileIO & F_CSV) != 0 {
		saver <- saveData{
			Number:  req.Qamount,
			Type:    CH_SAVPREP,
			Message: req.SavePath,
		}
	}
	// run queries in a loop
	var results []SingleQueryResult
	var result SingleQueryResult
	var err error
	for i := range queries {
		// run query
		result, err = runCsvQuery(queries[i], req)
		message("Finishing a query...")
		results = append(results, result)
		if err != nil {
			message(Sprint(err))
			return results, errors.New("Query " + Itoa(i+1) + " Error: " + Sprint(err))
		}
	}
	return results, nil
}

func runCommand() {
	if *flags.command == "" {
		return
	}
	q := QuerySpecs{QueryString: *flags.command, save: true} // sends output to stdout
	saver <- saveData{Type: CH_SAVPREP}
	CsvQuery(&q)
	saver <- saveData{Type: CH_NEXT}
	os.Exit(0)
}

func readStdin() {
	fi, _ := os.Stdin.Stat()
	if fi.Mode()&os.ModeNamedPipe != 0 {
		reader := bufio.NewReader(os.Stdin)
		buf := make([]byte, 10000)
		reader.Read(buf)
		*flags.command = string(buf)
	}
}
