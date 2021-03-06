package util

import "fmt"
import "io"
import "github.com/reiver/go-whitespace"
import "bytes"
import "encoding/json"
import "os"
import "bufio"
import "log"
import "strconv"

type RawInputHandler struct {
	output io.Writer
	buff   []rune
}

type SumoOperator interface {
	Process(map[string]interface{})
}

type SumoAggOperator interface {
	Process(map[string]interface{})
	Flush()
}

type ParseError string

func (e ParseError) Error() string {
	return string(e)
}

func ConnectToStdIn(operator SumoOperator) {
	fi, _ := os.Stdin.Stat() // get the FileInfo struct describing the standard input.
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		ConnectToReader(operator, os.Stdin)
	} else {
		fmt.Println("No input")
		return
	}
}

func ConnectToReader(operator SumoOperator, reader io.Reader) {
	bio := bufio.NewReader(reader)
	var line, hasMoreInLine, err = bio.ReadLine()
	var buf []byte
	for err != io.EOF {
		buf = append(buf, line...)
		if !hasMoreInLine && len(buf) > 0 {
			var rawMsg interface{}
			err := json.Unmarshal(buf, &rawMsg)
			buf = []byte{}
			if err != nil {
				log.Println("Error parsing json")
			} else {
				mapMessage, ok := rawMsg.(map[string]interface{})
				if ok {
					operator.Process(mapMessage)
				} else {
					log.Println("Unexpected JSON")
				}
			}
		}
		line, hasMoreInLine, err = bio.ReadLine()
	}
}

const Plus = "PLUS"
const StartRelation = "StartRelation"
const EndRelation = "EndRelation"
const Raw = "_raw"
const Type = "_type"
const Meta = "Meta"
const Relation = "Relation"

func IsPlus(inp map[string]interface{}) bool {
	tpe, ok := inp[Type].(string)
	return ok && tpe == Plus
}

func IsStartRelation(inp map[string]interface{}) bool {
	tpe, ok := inp[Type].(string)
	return ok && tpe == StartRelation
}

func IsEndRelation(inp map[string]interface{}) bool {
	tpe, ok := inp[Type].(string)
	return ok && tpe == EndRelation
}

func IsRelation(inp map[string]interface{}) bool {
	tpe, ok := inp[Type].(string)
	return ok && tpe == Relation
}

func IsMeta(inp map[string]interface{}) bool {
	tpe, ok := inp[Type].(string)
	return ok && tpe == Meta
}

func CreateStartRelation() map[string]interface{} {
	return map[string]interface{}{Type: StartRelation}
}

func CreateEndRelation() map[string]interface{} {
	return map[string]interface{}{Type: EndRelation}
}

func CreateRelation(inp map[string]interface{}) map[string]interface{} {
	inp[Type] = Relation
	return inp
}

func CreateMeta(inp map[string]interface{}) map[string]interface{} {
	inp[Type] = Meta
	return inp
}

func ExtractRaw(inp map[string]interface{}) string {
	raw, ok := inp[Raw].(string)
	if ok {
		return raw
	} else {
		return ""
	}
}

func (handler *RawInputHandler) Process(inp []byte) {
	runes := bytes.Runes(inp)
	// If not whitespace, flush, append
	if len(runes) > 0 && !whitespace.IsWhitespace(runes[0]) {
		handler.Flush()
		handler.buff = append(handler.buff, runes...)
	} else {
		// If it is whitespace, just append with a newline
		handler.buff = append(handler.buff, '\n')
		handler.buff = append(handler.buff, runes...)
	}
}

func NewRawInputHandler(inp io.Writer) *RawInputHandler {
	return &RawInputHandler{inp, []rune{}}
}

func NewRawInputHandlerStdout() *RawInputHandler {
	return NewRawInputHandler(os.Stdout)
}

func (handler *RawInputHandler) Flush() {
	m := make(map[string]interface{})
	m[Raw] = string(handler.buff)
	m[Type] = Plus
	handler.buff = []rune{}
	b, err := json.Marshal(m)
	//fmt.Printf(b)
	if err != nil {
		fmt.Printf("ERROR!", err)
	} else {
		handler.output.Write(b)
		handler.output.Write([]byte{'\n'})
	}
}

type JsonWriter struct {
	writer io.Writer
}

func NewJsonWriter() *JsonWriter {
	return &JsonWriter{os.Stdout}
}

func (writer *JsonWriter) Write(inp map[string]interface{}) {
	jsonBytes, err := json.Marshal(inp)
	//fmt.Printf(b)
	if err != nil {
		fmt.Printf("ERROR!", err)
	} else {
		writer.writer.Write(jsonBytes)
		writer.writer.Write([]byte{'\n'})
	}
}

func CoerceNumber(v interface{}) (float64, error) {
	return strconv.ParseFloat(fmt.Sprint(v), 64)
}
