package csv

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oarkflow/errors"
	"github.com/oarkflow/gologger"
	"github.com/oarkflow/log"
	"github.com/oarkflow/pkg/expr"
	"github.com/oarkflow/xid"

	"github.com/oarkflow/nepse/nepse/csv/csvtool"
)

type Query struct {
	File        string `json:"file" form:"file"`
	QueryString string `json:"query" form:"query"`
	Export      bool   `json:"export" form:"export"`
}

func (q *Query) QueryCsv() []map[string]csvtool.Value {
	return PrepareCsvResponseWithHeader(q.Query())
}

func (q *Query) Report() map[string]DataCount {
	getCols, _ := QueryCsv(q.File, "SELECT * FROM @file LIMIT 1")
	return ColDataReport(q.Query(), getCols)
}

func (q *Query) Columns() []string {
	getCols, _ := QueryCsv(q.File, "SELECT * FROM @file LIMIT 1")
	return getCols.Colnames
}

func (q *Query) Build() {
	if q.QueryString == "" {
		q.QueryString = "SELECT * FROM @file"
	}
}

func (q *Query) Query() csvtool.SingleQueryResult {
	if q.QueryString == "" {
		q.QueryString = "SELECT * FROM @file LIMIT 20"
	}

	fmt.Println(fmt.Sprintf("Query started at %s", time.Now()))
	data, err := QueryCsv(q.File, q.QueryString)

	fmt.Println(fmt.Sprintf("Query ended at %s", time.Now()))
	if err != nil {
		log.Info().Msg("Query execution failed ")

		panic(err)
	}

	return data
}

func (q *Query) ExportCsv() {
	data := q.Query()
	writer, err := gologger.New(q.File, 3000)
	if err != nil {
		panic(err)
	}

	writer.WriteString(strings.Join(data.Colnames, ","))
	fmt.Println(fmt.Sprintf("Writing to file started at %s", time.Now()))
	for _, v := range data.Vals {
		var tmp []string
		for _, vx := range v {
			tmp = append(tmp, vx.String())
		}
		writer.WriteString(strings.Join(tmp, ","))
	}
	fmt.Println(fmt.Sprintf("Writing to file ended at %s", time.Now()))
}

func QueryCsv(fileName string, query string, delim ...rune) (csvtool.SingleQueryResult, error) {
	delimeter := ','
	if len(delim) > 0 {
		delimeter = delim[0]
	}
	file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {

		return csvtool.SingleQueryResult{}, errors.New(fmt.Sprintf("File: %s not found or unable to open file", fileName))
	}
	defer file.Close()
	queryString := strings.Replace(query, "@file", "'"+fileName+"'", 1)
	queryString = cleanQuery(queryString)
	q := csvtool.QuerySpecs{
		QueryString: queryString,
		Comma:       delimeter,
	}

	res, err := csvtool.CsvQuery(&q)
	if err != nil {

		return res, errors.NewE(err, "error on CsvQuery", "")
	}

	return res, nil

}

func QueryCsvWithStats(fileName string, query string) (csvtool.SingleQueryResult, *csvtool.CsvDetail, error) {
	queryForCsvDetail := strings.Split(query, " LIMIT ")[0]
	csvDetail, err := CsvStats(fileName, queryForCsvDetail+"LIMIT 5")
	if err != nil {
		return csvtool.SingleQueryResult{}, nil, err
	}
	res, err := QueryCsv(fileName, query)
	if err != nil {
		return csvtool.SingleQueryResult{}, nil, err
	}
	return res, csvDetail, nil
}

func CsvStats(fileName string, query string) (*csvtool.CsvDetail, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {

		return nil, errors.New("File not found or unable to open file")
	}
	defer file.Close()

	queryString := strings.Replace(query, "@file", "'"+fileName+"'", 1)
	queryString = cleanQuery(queryString)
	q := &csvtool.QuerySpecs{
		QueryString: queryString,
		Comma:       ',',
	}

	data, err := csvtool.CsvStats(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to CsvStats", "")
	}
	totalRecords, err := csvtool.LineCounter(fileName)
	if err != nil {
		return nil, errors.NewE(err, "unable to LineCounter", "")
	}
	data.TotalRecords = totalRecords
	return data, nil
}

func cleanQuery(qry string) string {
	re := regexp.MustCompile(`\r?\n`)
	qry = re.ReplaceAllString(qry, " ")
	qry = strings.Join(strings.Fields(strings.TrimSpace(qry)), " ")
	return qry
}

func PrepareCsvResponseWithHeader(res csvtool.SingleQueryResult) []map[string]csvtool.Value {
	var data []map[string]csvtool.Value
	for _, v := range res.Vals {
		tmp := map[string]csvtool.Value{}
		for j, k := range v {
			col := res.Colnames[j]
			tmp[col] = k
		}
		data = append(data, tmp)
	}
	return data
}

type DataCount map[csvtool.Value]int64

type ColData map[string]DataCount

func ColDataReport(res csvtool.SingleQueryResult, cols csvtool.SingleQueryResult, includeNumericData ...bool) ColData {
	numericData := false
	if len(includeNumericData) > 0 {
		numericData = true
	}
	colData := make(ColData, len(cols.Colnames))
	for _, v := range cols.Vals {
		for j, _ := range v {
			col := res.Colnames[j]
			colType := csvtool.GetDataType(res.Types[j])
			if !numericData {
				if !(colType == "integer" || colType == "float") {
					colData[col] = make(map[csvtool.Value]int64)
				}
			} else {
				colData[col] = make(map[csvtool.Value]int64)
			}
		}
	}
	for _, v := range res.Vals {
		for j, k := range v {
			col := res.Colnames[j]
			colType := csvtool.GetDataType(res.Types[j])
			if !numericData {
				if !(colType == "integer" || colType == "float") {
					colData[col][k]++
				}
			} else {
				colData[col][k]++
			}
		}
	}
	return colData
}

func PrepareCsvResponseWithHeaderJson(res csvtool.SingleQueryResult) []map[string]any {
	var data []map[string]any
	for _, v := range res.Vals {
		tmp := make(map[string]any)
		for j, k := range v {
			col := res.Colnames[j]
			tmp[col] = k.String()
		}
		data = append(data, tmp)
	}
	return data
}

func PrepareCsvResponseWithHeaderJsonWithIndex(res csvtool.SingleQueryResult) map[string]map[string]string {
	data := make(map[string]map[string]string)
	for _, v := range res.Vals {
		tmp := make(map[string]string)
		for j, k := range v {
			col := res.Colnames[j]
			tmp[col] = k.String()
		}
		id := xid.New().String()
		tmp["index_id"] = id
		data[id] = tmp
	}
	return data
}

func PrepareTableTitle(res csvtool.SingleQueryResult) []string {
	var data []string
	for _, v := range res.Colnames {
		tmp := strings.ReplaceAll(v, "_", " ")
		tmp = strings.Title(tmp)
		data = append(data, tmp)
	}
	return data
}

func QueryCsvData(file string, field string, search string, queryString string) (csvtool.SingleQueryResult, error) {
	if field != "" && search != "" {
		queryString = fmt.Sprintf("SELECT * FROM @file WHERE %s LIKE '%%%s%%'", field, search)
	} else if queryString == "" {
		queryString = "SELECT * FROM @file"
	}
	queryString = queryString + "  LIMIT 20"

	return QueryCsv(file, queryString)
}

type FieldStats struct {
	NumericFields []csvtool.ColumnDetail
	AlphaFields   []csvtool.ColumnDetail
	DateFields    []csvtool.ColumnDetail
}

func CsvDataStats(file string) (*FieldStats, error) {
	csvDetail, err := CsvStats(file, "SELECT * FROM @file LIMIT 5")
	if err != nil {
		return nil, err
	}
	var numericFields, alphaFields, dateFields []csvtool.ColumnDetail
	for _, field := range csvDetail.Columns {
		switch field.Type {
		case "string":
			alphaFields = append(alphaFields, field)
		case "date":
			dateFields = append(dateFields, field)
		case "integer", "float":
			numericFields = append(numericFields, field)
		}
	}
	fields := &FieldStats{
		NumericFields: numericFields,
		AlphaFields:   alphaFields,
		DateFields:    dateFields,
	}
	return fields, nil
}

type QueryFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

func (f *QueryFilter) ToString() string {
	switch f.Operator {
	case "eq":
		return fmt.Sprintf("'%s' = '%s'", f.Field, f.Value)
	case "neq":
		return fmt.Sprintf("'%s' <> '%s'", f.Field, f.Value)
	case "gt":
		return fmt.Sprintf("'%s' > '%s'", f.Field, f.Value)
	case "gte":
		return fmt.Sprintf("'%s' >= '%s'", f.Field, f.Value)
	case "lt":
		return fmt.Sprintf("'%s' < '%s'", f.Field, f.Value)
	case "lte":
		return fmt.Sprintf("'%s' <= '%s'", f.Field, f.Value)
	case "contains":
		return fmt.Sprintf("'%s' LIKE '%%%s%%'", f.Field, f.Value)
	case "starts_with":
		return fmt.Sprintf("'%s' LIKE '%s%%'", f.Field, f.Value)
	case "ends_with":
		return fmt.Sprintf("'%s' LIKE '%%%s'", f.Field, f.Value)
	default:
		return ""
	}
}

type QueryAggregate struct {
	Field     string `json:"field"`
	Operation string `json:"workflow"`
	Distinct  bool   `json:"distinct"`
}

type QueryBuilder struct {
	File       string           `json:"file,omitempty"`
	Fields     []string         `json:"fields,omitempty"`
	Aggregates []QueryAggregate `json:"aggregates,omitempty"`
	Filters    []QueryFilter    `json:"filters,omitempty"`
	Export     bool             `json:"export"`
	Preview    bool             `json:"preview"`
}

func (q *QueryBuilder) Build() string {
	sql := "SELECT "
	for key, field := range q.Fields {
		q.Fields[key] = "'" + field + "'"
	}
	selectedFields := " * "
	if len(q.Fields) > 0 {
		selectedFields = strings.Join(q.Fields, ", ")
	}
	sql += selectedFields
	var aggregates []string
	for _, aggregate := range q.Aggregates {
		operator := strings.ToUpper(aggregate.Operation)
		alias := "count_total_rows"
		if aggregate.Field != "*" {
			alias = strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(operator+"_"+aggregate.Field), " ", "_"), "-", "_")
		}
		field := fmt.Sprintf("%s('%s') as %s", operator, aggregate.Field, alias)
		aggregates = append(aggregates, field)
	}
	if len(aggregates) > 0 {
		sql += ", " + strings.Join(aggregates, ", ") + " FROM @file "
	} else {
		sql += " FROM @file "
	}
	var filters []string
	for _, filter := range q.Filters {
		filters = append(filters, filter.ToString())
	}
	if len(filters) > 0 {
		sql += " WHERE " + strings.Join(filters, " AND ")
	}
	if len(aggregates) > 0 {
		sql += " GROUP BY " + selectedFields
	}
	if q.Preview {
		sql += " LIMIT 10"
	}
	return cleanQuery(sql)
}

func (q *QueryBuilder) Run(preview ...bool) (csvtool.SingleQueryResult, *csvtool.CsvDetail, error) {
	return QueryCsvWithStats(q.File, q.Build())
}

var fieldTypes = []string{
	"phone",
	"link",
	"email",
	"ip",
	"cc",
	"visa_cc",
	"mc_cc",
	"btc_address",
	"street_address",
	"zip_code",
	"po_box",
	"ssn",
	"md5",
	"sha1",
	"sha256",
	"guid",
	"mac_address",
	"iban",
	"git_repo",
}

func DetectFieldWithTypes(file string) (map[string][]string, error) {
	var fieldWithTypes = make(map[string][]string)
	data, err := QueryCsv(file, "SELECT * FROM @file LIMIT 1")
	if err != nil {
		return nil, errors.NewE(err, "DetectFieldWithTypes", "")
	}
	for _, row := range PrepareCsvResponseWithHeader(data) {
		for field, value := range row {
			parsedFields := expr.ParseMultiple(value.String(), fieldTypes...)
			for pattern, val := range parsedFields {
				if len(val) > 0 {
					if _, ok := fieldWithTypes[pattern]; !ok {
						fieldWithTypes[pattern] = []string{}
					}
					fieldWithTypes[pattern] = append(fieldWithTypes[pattern], field)
				}
			}
		}
	}
	return fieldWithTypes, nil
}

func GetHeader(scanner *bufio.Scanner, comma rune) map[int]string {
	scanner.Scan()
	r := csv.NewReader(strings.NewReader(scanner.Text()))
	r.Comma = comma
	r.TrimLeadingSpace = true
	colHeader, _ := r.Read()
	colPosition := make(map[int]string)
	for key, col := range colHeader {
		colPosition[key] = col
	}
	return colPosition
}

func ToMap(reader io.Reader, phoneKey string, comma rune, defaultPrefix string) []map[string]any {
	scanner := bufio.NewScanner(reader)
	colPosition := GetHeader(scanner, comma)
	for k, v := range colPosition {
		colPosition[k] = clean([]byte(v))
	}
	jobs := make(chan []byte)
	results := make(chan map[string]any)

	wg := new(sync.WaitGroup)
	for w := 1; w <= 2; w++ {
		wg.Add(1)
		go ProcessRecord(jobs, results, wg, colPosition, phoneKey, comma, defaultPrefix)
	}
	go func() {
		for scanner.Scan() {
			jobs <- scanner.Bytes()
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var data []map[string]any
	for v := range results {
		data = append(data, v)
	}

	return data
}

func clean(s []byte) string {
	j := 0
	for _, b := range s {
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' || b == '_' {
			s[j] = b
			j++
		}
	}
	return strings.TrimSpace(string(s[:j]))
}

func ProcessRecord(jobs <-chan []byte, results chan<- map[string]any, wg *sync.WaitGroup, col map[int]string, phoneKey string, comma rune, defaultPrefix string) {
	defer wg.Done()
	for j := range jobs {
		data := make(map[string]any)
		r := csv.NewReader(bytes.NewReader(j))
		r.Comma = comma
		r.TrimLeadingSpace = true
		fields, _ := r.Read()
		for key, dt := range fields {
			data[col[key]] = strings.TrimSpace(dt)
		}
		results <- data
	}
}
