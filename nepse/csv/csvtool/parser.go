/*
<query>             -> <options> <Select> <from> <where> <groupby> <having> <orderby> { limit # }
<options>           -> (c | nh | h) <options> | ε
<Select>            -> Select { top # } <Selections>
<Selections>        -> * <Selections> | {alias =} <exprAdd> {as alias} <Selections> | ε
<exprAdd>           -> <exprMult> ( + | - ) <exprAdd> | <exprMult>
<exprMult>          -> <exprNeg> ( * | / | % ) <exprMult> | <exprNeg>
<exprNeg>           -> { - } <exprCase>
<exprCase>          -> case <caseWhenPredList> { else <exprAdd> } end
                     | case <exprAdd> <caseWhenExprList> { else <exprAdd> } end
                     | <value>
<value>             -> column | literal | ( <exprAdd> ) | <function>
<caseWhenExprList>  -> <caseWhenExpr> <caseWhenExprList> | <caseWhenExpr>
<caseWhenExpr>      -> when <exprAdd> then <exprAdd>
<caseWhenPredList>  -> <casePredicate> <caseWhenPredList> | <casePredicate>
<casePredicate>     -> when <predicates> then <exprAdd>
<predicates>        -> <predicateCompare> { <logop> <predicates> }
<predicateCompare>  -> {not} <exprAdd> {not} <relop> <exprAdd>
                     | {not} <exprAdd> {not} between <exprAdd> and <exprAdd>
                     | {not} <exprAdd> in ( <expressions> )
                     | {not} ( predicates )
<from>              -> from filename { nh | noheader } { {as} alias } { nh | noheader } <join>
<join>              -> { left | inner | ε } { inner | outer | ε } join file as alias on <predicates> { <join> }
<where>             -> where <predicates> | ε
<having>            -> having <predicates> | ε
<groupby>           -> group by <expressions> | ε
<expressions>       -> <exprAdd> { <expressions> }
<orderby>           -> order by <exprAdd> | ε
*/

package csvtool

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	. "fmt"
	"io"
	"os"
	"regexp"
	. "strconv"
	"strings"

	"github.com/oarkflow/errors"
	"github.com/oarkflow/pkg/xopen"

	bt "github.com/google/btree"
)

// recursive descent parser builds parse tree and QuerySpecs
func parseQuery(q *QuerySpecs) (*Node, error) {
	_ = Print
	n := &Node{label: N_QUERY}
	n.tok1 = q
	err := scanTokens(q)
	if err != nil {
		return n, err
	}
	err = openFiles(q)
	if err != nil {
		return n, err
	}

	parseOptions(q)
	n.node1, err = parseSelect(q)
	if err != nil {
		return n, err
	}
	n.node2, err = parseFrom(q)
	if err != nil {
		return n, err
	}
	n.node3, err = parseWhere(q)
	if err != nil {
		return n, err
	}
	n.node4, err = parseGroupby(q)
	if err != nil {
		return n, err
	}
	n.node5, err = parseHaving(q)
	if err != nil {
		return n, err
	}
	q.sortExpr, err = parseOrder(q)
	if err != nil {
		return n, err
	}
	parseLimit(q)

	if q.Tok().id != EOS {
		err = ErrMsg(q.Tok(), "Expected end of query, got "+q.Tok().val)
	}
	if err != nil {
		return n, errors.NewE(err, "unable to parseLimit", "")
	}

	// add 'having' and 'order by' expressions to selections if grouping
	if q.sortExpr != nil && q.groupby {
		nn := n.node1.node1
		for ; nn.node2 != nil; nn = nn.node2 {
		}
		nn.node2 = &Node{
			label: N_SELECTIONS,
			tok3:  1 << 4,
			node1: q.sortExpr,
		}
	}
	findHavingAggregates(q, n, n.node5)

	// process selections
	_, _, _, err = aggCheck(n.node1)
	if err != nil {
		return n, errors.NewE(err, "unable to aggCheck", "")
	}
	_, _, _, err = typeCheck(n.node1)
	if err != nil {
		return n, errors.NewE(err, "unable to typeCheck", "")
	}
	branchShortener(q, n.node1)
	columnNamer(q, n.node1)
	_, f := findAggregateFunctions(q, n.node1)
	if f {
		return n, ErrMsg(q.Tok(), "Cannot have aggregate function inside an aggregate function")
	}

	// process 'from' section
	_, _, _, err = typeCheck(n.node2)
	if err != nil {
		return n, errors.NewE(err, "unable to typeCheck", "")
	}
	_, err = joinExprFinder(q, n.node2, "")
	if err != nil {
		return n, errors.NewE(err, "unable to joinExprFinder", "")
	}
	branchShortener(q, n.node2)

	// process 'where' section
	if e := findAggregateFunction(n.node3); e > 0 {
		return n, ErrMsg(q.Tok(), "Cannot have aggregate function in 'where' clause")
	}
	_, _, _, err = typeCheck(n.node3)
	if err != nil {
		return n, errors.NewE(err, "unable to typeCheck ", "")
	}
	branchShortener(q, n.node3.node1)

	// process groups
	_, _, _, err = typeCheck(n.node4)
	if err != nil {
		return n, errors.NewE(err, "unable to typeCheck process groups", "")
	}
	branchShortener(q, n.node4)

	// process sort expression separately if not grouping
	if !(q.sortExpr != nil && q.groupby) {
		_, sortType, _, er := typeCheck(q.sortExpr)
		if er != nil {
			return n, errors.NewE(err, "unable to typeCheck sort expression separately", "")
		}
		err = enforceType(q.sortExpr, sortType)
		branchShortener(q, q.sortExpr)
	}

	return n, err
}

func CsvStats(q *QuerySpecs) (*CsvDetail, error) {
	_ = Print
	n := &Node{label: N_QUERY}
	n.tok1 = q
	err := scanTokens(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to scanTokens", "")
	}
	err = openFiles(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to openFiles", "")
	}

	parseOptions(q)
	n.node1, err = parseSelect(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to parseSelect", "")
	}
	n.node2, err = parseFrom(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to parseFrom", "")
	}
	n.node3, err = parseWhere(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to parseWhere", "")
	}
	n.node4, err = parseGroupby(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to parseGroupby", "")
	}
	n.node5, err = parseHaving(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to parseHaving", "")
	}
	q.sortExpr, err = parseOrder(q)
	if err != nil {
		return nil, errors.NewE(err, "unable to parseOrder", "")
	}
	parseLimit(q)

	if q.Tok().id != EOS {
		err = ErrMsg(q.Tok(), "Expected end of query, got "+q.Tok().val)
	}
	if err != nil {
		return nil, errors.NewE(err, "unable to parseLimit", "")
	}

	// add 'having' and 'order by' expressions to selections if grouping
	if q.sortExpr != nil && q.groupby {
		nn := n.node1.node1
		for ; nn.node2 != nil; nn = nn.node2 {
		}
		nn.node2 = &Node{
			label: N_SELECTIONS,
			tok3:  1 << 4,
			node1: q.sortExpr,
		}
	}
	findHavingAggregates(q, n, n.node5)

	// process selections
	_, _, _, err = aggCheck(n.node1)
	if err != nil {
		return nil, err
	}
	_, _, _, err = typeCheck(n.node1)
	if err != nil {
		return nil, err
	}
	branchShortener(q, n.node1)
	columnNamer(q, n.node1)
	colDetail := make(map[int]ColumnDetail)
	columnDetails := &CsvDetail{
		TotalColumn: q.colSpec.NewWidth,
		Columns:     colDetail,
	}
	for i := 0; i < q.colSpec.NewWidth; i++ {
		name := q.colSpec.NewNames[i]
		colType := typeMap[q.colSpec.NewTypes[i]]
		colDetail[i+1] = ColumnDetail{
			Name:     name,
			Type:     colType,
			Position: i + 1,
		}
	}
	return columnDetails, nil
}

func GetDataType(i int) string {
	return typeMap[i]
}

func LineCounter(fileName string, noHeader ...bool) (int64, error) {
	r, err := xopen.Ropen(fileName)
	if err != nil {
		return 0, errors.NewE(err, "unable to Ropen", "")
	}
	var count int64
	const lineBreak = '\n'

	buf := make([]byte, bufio.MaxScanTokenSize)

	for {
		bufferSize, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return 0, errors.NewE(err, "unable to Read buffer", "")
		}

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], lineBreak)
			if i == -1 || bufferSize == buffPosition {
				break
			}
			buffPosition += i + 1
			count++
		}
		if err == io.EOF {
			break
		}
	}

	if len(noHeader) > 0 && noHeader[0] {
		return count, nil
	}
	return count - 1, nil
}

// already processed with openFiles()
func parseOptions(q *QuerySpecs) {
	switch q.Tok().Lower() {
	case "c":
	case "h":
	case "nh":
	default:
		return
	}
	q.NextTok()
	parseOptions(q)
}

// node1 is selections
func parseSelect(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_SELECT}
	var err error
	if q.Tok().Lower() != "select" {
		return n, ErrMsg(q.Tok(), "Expected 'select'. Found "+q.Tok().val)
	}
	q.NextTok()
	err = parseTop(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseTop", "")
	}
	countSelected = 0
	n.node1, err = parseSelections(q)
	return n, err
}

// node2 is chain of selections for all infile columns
func selectAll(q *QuerySpecs) (*Node, error) {
	var err error
	n := &Node{label: N_SELECTIONS, tok3: 0}
	firstSelection := n
	var lastSelection *Node
	for j := 1; ; j++ {
		file, ok := q.files["_f"+Sprint(j)]
		if !ok {
			break
		}
		for i := range file.names {
			n.node2 = &Node{label: N_SELECTIONS, tok3: 0}
			n.node1 = &Node{
				label: N_VALUE,
				tok1:  i,
				tok2:  1,
				tok3:  file.types[i],
				tok5:  q.files["_f"+Sprint(j)],
			}
			countSelected++
			lastSelection = n
			n = n.node2
		}
	}
	lastSelection.node2, err = parseSelections(q)
	return firstSelection, err
}

// node1 is expression
// node2 is next selection
// tok1 is destination column indexes
// tok2 will be destination column name
// tok3 is bit array - 1 and 2 are distinct
// tok4 will be aggregate function
// tok5 will be type
func parseSelections(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_SELECTIONS}
	var err error
	var hidden bool
	if q.Tok().id == SP_COMMA {
		q.NextTok()
	}
	n.tok3 = 0
	switch q.Tok().id {
	case SP_STAR:
		q.NextTok()
		return selectAll(q)
	// tok3 bit 1 means distinct, bit 2 means hidden
	case KW_DISTINCT:
		n.tok3 = 1
		q.NextTok()
		if q.Tok().Lower() == "hidden" && !q.Tok().quoted {
			hidden = true
			n.tok3 = 3
			q.NextTok()
		}
		fallthrough
	// expression
	case KW_CASE:
		fallthrough
	case WORD:
		fallthrough
	case SP_MINUS:
		fallthrough
	case SP_LPAREN:
		if !hidden {
			countSelected++
		}
		// alias = expression
		if q.PeekTok().id == SP_EQ {
			if q.Tok().id != WORD {
				return n, ErrMsg(q.Tok(), "Alias must be a word. Found "+q.Tok().val)
			}
			n.tok2 = q.Tok().val
			q.NextTok()
			q.NextTok()
			n.node1, err = parseExprAdd(q)
			// expression
		} else {
			n.node1, err = parseExprAdd(q)
			if q.Tok().Lower() == "as" {
				n.tok2 = q.NextTok().val
				q.NextTok()
			}
		}
		if err != nil {
			return n, errors.NewE(err, "unable to parseSelections", "")
		}
		n.node2, err = parseSelections(q)
		return n, err
	// done with selections
	case KW_FROM:
		if countSelected == 0 {
			return selectAll(q)
		}
		return nil, nil
	default:
		return n, ErrMsg(q.Tok(), "Expected a new selection or 'from' clause. Found "+q.Tok().val)
	}
}

// node1 is exprMult
// node2 is exprAdd
// tok1 is add/minus operator
func parseExprAdd(q *QuerySpecs) (*Node, error) {
	var err error
	n := &Node{label: N_EXPRADD}
	n.node1, err = parseExprMult(q)
	if err != nil {
		return n, err
	}
	switch q.Tok().id {
	case SP_MINUS:
		fallthrough
	case SP_PLUS:
		n.tok1 = q.Tok().id
		q.NextTok()
		n.node2, err = parseExprAdd(q)
	}
	return n, err
}

// node1 is exprNeg
// node2 is exprMult
// tok1 is mult/div operator
func parseExprMult(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_EXPRMULT}
	var err error
	n.node1, err = parseExprNeg(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseExprNeg", "")
	}
	switch q.Tok().id {
	case SP_STAR:
		if q.PeekTok().Lower() == "from" {
			break
		}
		fallthrough
	case SP_MOD:
		fallthrough
	case SP_CARROT:
		fallthrough
	case SP_DIV:
		n.tok1 = q.Tok().id
		q.NextTok()
		n.node2, err = parseExprMult(q)
	}
	return n, err
}

// tok1 is minus operator
// node1 is exprCase
func parseExprNeg(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_EXPRNEG}
	var err error
	if q.Tok().id == SP_MINUS {
		n.tok1 = q.Tok().id
		q.NextTok()
	}
	n.node1, err = parseExprCase(q)
	return n, err
}

// tok1 is [case, word, expr] token - tells if case, terminal value, or (expr)
// tok2 is [when, expr] token - tells what kind of case. predlist, or expr exprlist respectively
// node2.tok3 will be initial 'when' expression type
// node1 is (expression), when predicate list, expression for exprlist
// node2 is expression list to compare to initial expression
// node3 is else expression
func parseExprCase(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_EXPRCASE}
	var err error
	switch q.Tok().id {
	case KW_CASE:
		n.tok1 = KW_CASE

		switch q.NextTok().id {
		// when predicates are true
		case KW_WHEN:
			n.tok2 = KW_WHEN
			n.node1, err = parseCaseWhenPredList(q)
			if err != nil {
				return n, errors.NewE(err, "unable to  parse Case When PredList", "")
			}
		// expression matches expression list
		case WORD:
			fallthrough
		case SP_LPAREN:
			n.tok2 = N_EXPRADD
			n.node1, err = parseExprAdd(q)
			if err != nil {
				return n, errors.NewE(err, "unable to parseExprAdd", "")
			}
			if q.Tok().Lower() != "when" {
				return n, ErrMsg(q.Tok(), "Expected 'when' after case expression. Found "+q.Tok().val)
			}
			n.node2, err = parseCaseWhenExprList(q)
			if err != nil {
				return n, errors.NewE(err, "unable to parse Case When ExprList", "")
			}
		default:
			return n, ErrMsg(q.Tok(), "Expected expression or 'when'. Found "+q.Tok().val)
		}

		switch q.Tok().id {
		case KW_END:
			q.NextTok()
		case KW_ELSE:
			q.NextTok()
			n.node3, err = parseExprAdd(q)
			if err != nil {
				return n, errors.NewE(err, "unable to  parseExprAdd", "")
			}
			if q.Tok().Lower() != "end" {
				return n, ErrMsg(q.Tok(), "Expected 'end' after 'else' expression. Found "+q.Tok().val)
			}
			q.NextTok()
		default:
			return n, ErrMsg(q.Tok(), "Expected 'end' or 'else' after case expression. Found "+q.Tok().val)
		}

	case WORD:
		n.tok1 = WORD
		n.node1, err = parseValue(q)
	case SP_LPAREN:
		n.tok1 = N_EXPRADD
		q.NextTok()
		n.node1, err = parseExprAdd(q)
		if err != nil {
			return n, errors.NewE(err, "unable to parseExprAdd", "")
		}
		if q.Tok().id != SP_RPAREN {
			return n, ErrMsg(q.Tok(), "Expected closing parenthesis. Found "+q.Tok().val)
		}
		q.NextTok()
	default:
		return n, ErrMsg(q.Tok(), "Expected case, value, or expression. Found "+q.Tok().val)
	}
	return n, err
}

// if implement dot notation, put parser here
// tok1 is [value, column index, function id]
// tok2 is [0,1,2] for literal/column/function
// tok3 is type
// tok4 is type in special cases like FN_COUNT
// tok5 is filedata
// node1 is function expression if doing that
var cInt *regexp.Regexp = regexp.MustCompile(`^c\d+$`)

func parseValue(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_VALUE}
	var err error
	tok := q.Tok()
	// see if it's a function
	if fn, ok := functionMap[tok.Lower()]; ok && !tok.quoted && q.PeekTok().id == SP_LPAREN {
		n.tok1 = fn
		n.tok2 = 2
		n.node1, err = parseFunction(q)
		if err != nil {
			return n, errors.NewE(err, "unable to parseFunction", "")
		}
		return n, err
		// any non-function value
	} else {
		// determine file source and value
		S := strings.SplitAfterN(tok.val, ".", 2)
		var fdata *FileData
		var ok bool
		var value string
		if len(S) == 2 && q.aliases {
			alias := strings.TrimRight(S[0], ".")
			fdata, ok = q.files[alias]
			value = S[1]
			if !ok {
				value = tok.val
				fdata = q.files["_f1"]
			}
		} else {
			value = tok.val
			fdata = q.files["_f1"]
		}
		// try column number
		if num, er := Atoi(value); q.intColumn && !tok.quoted && er == nil {
			if num < 1 || num > fdata.width {
				return n, ErrMsg(q.Tok(), "Column number out of bounds:"+Sprint(num))
			}
			n.tok1 = num - 1
			n.tok2 = 1
			n.tok3 = fdata.types[num-1]
			n.tok5 = fdata
		} else if !tok.quoted && cInt.MatchString(value) {
			num, _ := Atoi(value[1:])
			if num < 1 || num > fdata.width {
				return n, ErrMsg(q.Tok(), "Column number out of bounds:"+Sprint(num))
			}
			n.tok1 = num - 1
			n.tok2 = 1
			n.tok3 = fdata.types[num-1]
			n.tok5 = fdata
			// try column name
		} else if n.tok1, err = getColumnIdx(fdata.names, value); err == nil {
			n.tok2 = 1
			n.tok3 = fdata.types[n.tok1.(int)]
			n.tok5 = fdata
			// else must be literal
		} else {
			err = nil
			n.tok2 = 0
			n.tok3 = getNarrowestType(value, 0)
			n.tok1 = value
		}
	}
	q.NextTok()
	return n, err
}

// tok1 says if more predicates
// node1 is case predicate
// node2 is next case predicate list node
func parseCaseWhenPredList(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_CPREDLIST}
	var err error
	n.node1, err = parseCasePredicate(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseCasePredicates", "")
	}
	if q.Tok().Lower() == "when" {
		n.tok1 = 1
		n.node2, err = parseCaseWhenPredList(q)
	}
	return n, err
}

// node1 is predicates
// node2 is expression if true
func parseCasePredicate(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_CPRED}
	var err error
	q.NextTok() // eat when token
	n.node1, err = parsePredicates(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parsePredicates", "")
	}
	if q.Tok().Lower() != "then" {
		return n, ErrMsg(q.Tok(), "Expected 'then' after predicate. Found: "+q.Tok().val)
	}
	q.NextTok() // eat then token
	n.node2, err = parseExprAdd(q)
	return n, err
}

// tok1 is logop
// tok2 is negation
// node1 is predicate comparison
// node2 is next predicates node
func parsePredicates(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_PREDICATES}
	var err error
	n.tok2 = 0
	if q.Tok().id == SP_NEGATE {
		n.tok2 = 1
		q.NextTok()
	}
	n.node1, err = parsePredCompare(q)
	if err != nil {
		return n, errors.NewE(err, "unable parsePredCompare", "")
	}
	if (q.Tok().id & LOGOP) != 0 {
		n.tok1 = q.Tok().id
		q.NextTok()
		n.node2, err = parsePredicates(q)
	}
	return n, err
}

// more strict version of parsePredicates for joins
// tok3 tells how many conditions
func parseJoinPredicates(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_PREDICATES}
	var err error
	n.tok2 = 0
	n.tok3 = 1
	n.node1, err = parseJoinPredCompare(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseJoinPredCompare", "")
	}
	if (q.Tok().id & LOGOP) != 0 {
		return n, ErrMsg(q.Tok(), "Only one join condition per file")
	}
	/*
		if q.Tok().id == KW_AND {
			n.tok3 = 2
			q.NextTok()
			var comparison *Node
			comparison, err = parseJoinPredCompare(q)
			n.node2 = &Node{
				label: N_PREDICATES,
				node1: comparison,
			}
		} else if (q.Tok().id & LOGOP) != 0 {
			return n,ErrMsg(q.Tok(),"Join conditions support one 'and' operator per join")
		}
	*/
	return n, err
}

// more strict version of parsePredCompare for joins
func parseJoinPredCompare(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_PREDCOMP}
	var err error
	n.node1, err = parseExprAdd(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseExprAdd", "")
	}
	if q.Tok().id != SP_EQ {
		return n, ErrMsg(q.Tok(), "Expected = operator. Found: "+q.Tok().val)
	}
	q.NextTok()
	n.node2, err = parseExprAdd(q)
	return n, errors.NewE(err, "unable to parseExprAdd", "")
}

// tok1 is [relop, paren] for comparison or more predicates
// tok2 is negation
// tok3 will be independant type
// tok4 will be join file key
// tok5 will be join struct
// node1 is [expr, predicates]
// node2 is second expr
// node3 is third expr for betweens
func parsePredCompare(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_PREDCOMP}
	var err error
	var negate int
	var olderr error
	if q.Tok().id == SP_NEGATE {
		negate ^= 1
		q.NextTok()
	}
	// more predicates in parentheses
	if q.Tok().id == SP_LPAREN {
		pos := q.tokIdx
		// try parsing as predicate
		q.NextTok()
		n.node1, err = parsePredicates(q)
		if q.Tok().id != SP_RPAREN {
			return n, ErrMsg(q.Tok(), "Expected cosing parenthesis. Found:"+q.Tok().val)
		}
		q.NextTok()
		// if failed, reparse as expression
		if err != nil {
			q.tokIdx = pos
			olderr = err
		} else {
			return n, err
		}
	}
	// comparison
	n.node1, err = parseExprAdd(q)
	if err != nil && olderr != nil {
		return n, olderr
	}
	if err != nil {
		return n, errors.NewE(err, "unable to parsePredCompare", "")
	}
	if q.Tok().id == SP_NEGATE {
		negate ^= 1
		q.NextTok()
	}
	n.tok2 = negate
	if (q.Tok().id & RELOP) == 0 {
		return n, ErrMsg(q.Tok(), "Expected relational operator. Found: "+q.Tok().val)
	}
	n.tok1 = q.Tok().id
	q.NextTok()
	if n.tok1 == KW_LIKE {
		var like interface{}
		re := regexp.MustCompile("%")
		like = re.ReplaceAllString(q.Tok().val, ".*")
		re = regexp.MustCompile("_")
		like = re.ReplaceAllString(like.(string), ".")
		like, err = regexp.Compile("(?i)^" + like.(string) + "$")
		n.node2 = &Node{label: N_VALUE, tok1: liker{like.(*regexp.Regexp)}, tok2: 0, tok3: 0} // like gets 'null' type because it also doesn't effect workflow type
		q.NextTok()
	} else if n.tok1 == KW_IN {
		if q.Tok().id != SP_LPAREN {
			return n, ErrMsg(q.Tok(), "Expected opening parenthesis for expression list. Found: "+q.Tok().val)
		}
		q.NextTok()
		n.node2, err = parseExpressionList(q, true)
		if err != nil {
			return n, errors.NewE(err, "unable to parseExpressionList", "")
		}
		if q.Tok().id != SP_RPAREN {
			return n, ErrMsg(q.Tok(), "Expected closing parenthesis after expression list. Found: "+q.Tok().val)
		}
		q.NextTok()
	} else {
		n.node2, err = parseExprAdd(q)
		if err != nil {
			return n, errors.NewE(err, "unable to parseExprAdd", "")
		}
	}
	if n.tok1 == KW_BETWEEN {
		q.NextTok()
		n.node3, err = parseExprAdd(q)
	}
	return n, err
}

// tok1 int tells that there's another
// node1 is case expression
// node2 is next exprlist node
func parseCaseWhenExprList(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_CWEXPRLIST}
	var err error
	n.node1, err = parseCaseWhenExpr(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseCaseWhenExpr", "")
	}
	if q.Tok().Lower() == "when" {
		n.tok1 = 1
		n.node2, err = parseCaseWhenExprList(q)
	}
	return n, err
}

// node1 is comparison expression
// node2 is result expression
func parseCaseWhenExpr(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_CWEXPR}
	var err error
	q.NextTok() // eat when token
	n.node1, err = parseExprAdd(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseExprAdd", "")
	}
	q.NextTok() // eat then token
	n.node2, err = parseExprAdd(q)
	return n, err
}

// row limit at front of query
func parseTop(q *QuerySpecs) error {
	var err error
	if q.Tok().Lower() == "top" {
		q.quantityLimit, err = Atoi(q.NextTok().val)
		if err != nil {
			return ErrMsg(q.Tok(), "Expected number after 'top'. Found "+q.PeekTok().val)
		}
		q.NextTok()
	}
	return nil
}

// row limit at end of query
func parseLimit(q *QuerySpecs) error {
	var err error
	if q.Tok().Lower() == "limit" {
		q.quantityLimit, err = Atoi(q.NextTok().val)
		if err != nil {
			return ErrMsg(q.Tok(), "Expected number after 'limit'. Found "+q.PeekTok().val)
		}
		q.NextTok()
	}
	return nil
}

// tok1 is file path
// tok2 is alias
// node1 is joins
func parseFrom(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_FROM}
	var err error
	if q.Tok().Lower() != "from" {
		return n, ErrMsg(q.Tok(), "Expected 'from'. Found: "+q.Tok().val)
	}
	tok := strings.Replace(q.NextTok().val, "~/", os.Getenv("HOME")+"/", 1)
	n.tok1 = tok
	_, err = os.Stat(Sprint(n.tok1))
	if err != nil {
		return n, errors.NewE(err, "unable to Stat", "")
	}
	q.NextTok()
	t := q.Tok()
	if t.Lower() == "noheader" || t.Lower() == "nh" {
		t = q.NextTok()
	}
	switch t.id {
	case KW_AS:
		t = q.NextTok()
		if t.id != WORD {
			return n, ErrMsg(q.Tok(), "Expected alias after as. Found: "+t.val)
		}
		fallthrough
	case WORD:
		if _, ok := joinMap[t.Lower()]; ok {
			return n, ErrMsg(q.Tok(), "Join requires file aliases. Found: "+t.val)
		}
		n.tok2 = t.val
		q.NextTok()
	}
	if q.Tok().Lower() == "noheader" || q.Tok().Lower() == "nh" {
		q.NextTok()
	}
	n.node1, err = parseJoin(q)
	return n, err
}

// tok1 is join details (left/outer or inner)
// tok2 is [0,1] for small/big join file
// tok3 is filepath
// tok4 is alias
// node1 is join condition (predicates)
// node2 is next join
func parseJoin(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_JOIN}
	var err error
	switch q.Tok().Lower() {
	case "left":
	case "inner":
	case "outer":
	case "join":
	case "sjoin":
	case "bjoin":
	default:
		return nil, nil
	}
	q.joining = true
	n.tok1 = 0 // 1 is left/outer, 0 is inner
	switch q.Tok().Lower() {
	case "left":
		n.tok1 = 1
		q.NextTok()
	}
	switch q.Tok().Lower() {
	case "inner":
		n.tok1 = 0
		q.NextTok()
	case "outer":
		n.tok1 = 1
		q.NextTok()
	}
	sizeOverride := false
	switch q.Tok().Lower() {
	case "join":
		n.tok2 = 0
	case "sjoin":
		n.tok2 = 0
		sizeOverride = true
	case "bjoin":
		n.tok2 = 1
		sizeOverride = true
		q.bigjoin = true
	default:
		return n, ErrMsg(q.Tok(), "Expected 'join'. Found:"+q.Tok().val)
	}
	// file path
	n.tok3 = q.NextTok().val
	finfo, err := os.Stat(Sprint(n.tok3))
	if err != nil {
		return n, errors.NewE(err, "unable to os Stat", "")
	}
	// see if file is 100+ megabytes
	if !sizeOverride && finfo.Size() > 100000000 {
		n.tok2 = 1
		q.bigjoin = true
	}
	if err = eosError(q); err != nil {
		return n, errors.NewE(err, "unable to eosError", "")
	}
	// alias
	t := q.NextTok()
	if t.Lower() == "noheader" || t.Lower() == "nh" {
		t = q.NextTok()
	}
	switch t.id {
	case KW_AS:
		t = q.NextTok()
		if t.id != WORD {
			return n, ErrMsg(q.Tok(), "Expected alias after as. Found: "+t.val)
		}
		fallthrough
	case WORD:
		n.tok4 = t.val
	default:
		return n, ErrMsg(q.Tok(), "Join requires an alias. Found: "+q.Tok().val)
	}
	if q.PeekTok().Lower() == "noheader" || q.PeekTok().Lower() == "nh" {
		q.NextTok()
	}
	if _, ok := q.files[t.val]; !ok {
		return n, ErrMsg(q.Tok(), "Could not open file "+n.tok3.(string))
	}
	if q.NextTok().Lower() != "on" {
		return n, ErrMsg(q.Tok(), "Expected 'on'. Found: "+q.Tok().val)
	}
	q.NextTok()
	n.node1, err = parseJoinPredicates(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseJoinPredicates", "")
	}
	n.node2, err = parseJoin(q)
	return n, err
}

// node1 is conditions
func parseWhere(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_WHERE}
	var err error
	if q.Tok().Lower() != "where" {
		return n, nil
	}
	q.NextTok()
	n.node1, err = parsePredicates(q)
	return n, err
}

// node1 is conditions
func parseHaving(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_WHERE}
	var err error
	if q.Tok().Lower() != "having" {
		return n, nil
	}
	q.NextTok()
	n.node1, err = parsePredicates(q)
	return n, err
}

// doesn't add to parse tree yet, juset sets q member
func parseOrder(q *QuerySpecs) (*Node, error) {
	if q.Tok().id == EOS {
		return nil, nil
	}
	if q.Tok().Lower() == "order" {
		if q.NextTok().Lower() != "by" {
			return nil, ErrMsg(q.Tok(), "Expected 'by' after 'order'. Found "+q.Tok().val)
		}
		q.NextTok()
		expr, err := parseExprAdd(q)
		if q.Tok().Lower() == "asc" {
			q.NextTok()
			q.sortWay = 2
		}
		if _, ok := q.files["_f3"]; ok && q.joining && !q.groupby {
			return nil, ErrMsg(q.Tok(), "Non-grouping ordered queries can only join 2 files")
		}
		return expr, err
	}
	return nil, nil
}

// tok1 is function id
// tok2 will be intermediate index if aggragate
// tok3 is distinct btree for count, increment step
// tok4 true if using aes, increment start number
// node1 is expression in parens
func parseFunction(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_FUNCTION}
	var err error
	n.tok1 = functionMap[q.Tok().Lower()]
	q.NextTok()
	// count(*)
	if q.NextTok().id == SP_STAR && n.tok1.(int) == FN_COUNT {
		n.node1 = &Node{
			label: N_VALUE,
			tok1:  "1",
			tok2:  0,
			tok3:  1,
		}
		q.NextTok()
		// other functions
	} else {
		if (n.tok1.(int)&AGG_BIT) != 0 && q.Tok().Lower() == "distinct" {
			n.tok3 = bt.New(200)
			q.distinctAgg = true
			q.NextTok()
		}
		switch n.tok1.(int) {
		case FN_COALESCE:
			n.node1, err = parseExpressionList(q, true)
		case FN_ENCRYPT:
			fallthrough
		case FN_DECRYPT:
			// first param is expression to en/decrypt
			n.node1, err = parseExprAdd(q)
			needPrompt := true
			var password string
			if q.Tok().id == SP_COMMA {
				// second param is password
				password = q.NextTok().val
				needPrompt = false
				q.NextTok()
			} else if q.Tok().id != SP_RPAREN {
				return n, ErrMsg(q.Tok(), "Expect closing parenthesis or comma after expression in crypto function. Found: "+q.Tok().val)
			}
			if q.password == "" && needPrompt {
				q.password = promptPassword()
			}
			if password == "" {
				password = q.password
			}
			pass := sha256.Sum256([]byte(password))
			c, _ := aes.NewCipher(pass[:])
			gcm, _ := cipher.NewGCM(c)
			n.tok3 = gcm
			n.tok4 = true
			if err != nil {
				return n, errors.NewE(err, "unable to get user", "")
			}
		case FN_INC:
			if ff, err := ParseFloat(q.Tok().val, 64); err == nil {
				n.tok3 = float(ff)
				n.tok4 = float(1.0)
				q.NextTok()
			} else if q.Tok().id != SP_RPAREN {
				return n, ErrMsg(q.Tok(), "inc() function parameter must be increment amount or empty. Found: "+q.Tok().val)
			} else {
				n.tok3 = float(1.0)
				n.tok4 = float(1.0)
			}
		default:
			n.node1, err = parseExprAdd(q)
		}
	}
	if q.Tok().id != SP_RPAREN {
		return n, ErrMsg(q.Tok(), "Expected closing parenthesis after function. Found: "+q.Tok().val)
	}
	q.NextTok()
	// groupby if aggregate function
	if (n.tok1.(int) & AGG_BIT) != 0 {
		q.groupby = true
	}
	return n, err
}

// node1 is groupExpressions
// tok1 is groups map
func parseGroupby(q *QuerySpecs) (*Node, error) {
	n := &Node{label: N_GROUPBY}
	var err error
	if !(q.Tok().Lower() == "group" && q.PeekTok().Lower() == "by") {
		return nil, nil
	}
	if q.distinctAgg {
		return n, errors.New("Cannot use distinct in aggregate function when using 'group by'")
	}
	q.NextTok()
	q.NextTok()
	n.node1, err = parseExpressionList(q, false)
	n.tok1 = make(map[interface{}]interface{})
	return n, err
}

// node1 is expression
// node2 is expressions
// tok1 [0,1] for map returns row or next map when used for groups
func parseExpressionList(q *QuerySpecs, interdependant bool) (*Node, error) { // bool afg if expression types are interdependant
	label := N_EXPRESSIONS
	if interdependant {
		label = N_DEXPRESSIONS
	}
	n := &Node{label: label}
	var err error
	n.node1, err = parseExprAdd(q)
	if err != nil {
		return n, errors.NewE(err, "unable to parseExprAdd", "")
	}
	switch q.Tok().id {
	case SP_COMMA:
		q.NextTok()
	case WORD:
	case KW_CASE:
	case SP_LPAREN:
	default:
		n.tok1 = 0
		return n, err
	}
	n.tok1 = 1
	n.node2, err = parseExpressionList(q, interdependant)
	return n, err
}
