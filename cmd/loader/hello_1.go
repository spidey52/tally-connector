package main

// import (
// 	"encoding/xml"
// 	"fmt"
// 	"regexp"
// 	"tally-connector/cmd/loader/config"
// )

// type TableConfigYAML struct {
// 	Name       string
// 	Collection string
// 	Fields     []config.Field
// 	Fetch      []string
// 	Filters    []string
// }

// type TallyConfig struct {
// 	Company  string
// 	FromDate string
// 	ToDate   string
// }

// type Envelope struct {
// 	XMLName xml.Name `xml:"ENVELOPE"`
// 	Header  Header   `xml:"HEADER"`
// 	Body    Body     `xml:"BODY"`
// }

// type Header struct {
// 	Version      int    `xml:"VERSION"`
// 	TallyRequest string `xml:"TALLYREQUEST"`
// 	Type         string `xml:"TYPE"`
// 	ID           string `xml:"ID"`
// }

// type Body struct {
// 	Desc Desc `xml:"DESC"`
// }

// type Desc struct {
// 	StaticVariables StaticVariables `xml:"STATICVARIABLES"`
// 	TDL             TDL             `xml:"TDL"`
// }

// type StaticVariables struct {
// 	SVExportFormat   string `xml:"SVEXPORTFORMAT"`
// 	SVFromDate       string `xml:"SVFROMDATE"`
// 	SVToDate         string `xml:"SVTODATE"`
// 	SVCurrentCompany string `xml:"SVCURRENTCOMPANY,omitempty"`
// }

// type TDL struct {
// 	TDLMessage TDLMessage `xml:"TDLMESSAGE"`
// }

// type TDLMessage struct {
// 	Report     Report     `xml:"REPORT"`
// 	Form       Form       `xml:"FORM"`
// 	Part       Part       `xml:"PART"`
// 	Line       Line       `xml:"LINE"`
// 	Fields     []Field    `xml:"FIELD"`
// 	Collection Collection `xml:"COLLECTION"`
// }

// type Report struct {
// 	Name  string `xml:"NAME,attr"`
// 	Forms string `xml:"FORMS"`
// }

// type Form struct {
// 	Name  string `xml:"NAME,attr"`
// 	Parts string `xml:"PARTS"`
// }

// type Part struct {
// 	Name     string `xml:"NAME,attr"`
// 	Lines    string `xml:"LINES"`
// 	Repeat   string `xml:"REPEAT"`
// 	Scrolled string `xml:"SCROLLED"`
// }

// type Line struct {
// 	Name   string `xml:"NAME,attr"`
// 	Fields string `xml:"FIELDS"`
// }

// type Field struct {
// 	Name   string `xml:"NAME,attr"`
// 	Set    string `xml:"SET"`
// 	XMLTag string `xml:"XMLTAG"`
// }

// type Collection struct {
// 	Name   string `xml:"NAME,attr"`
// 	Type   string `xml:"TYPE"`
// 	Fetch  string `xml:"FETCH,omitempty"`
// 	Filter string `xml:"FILTER,omitempty"`
// }

// func generateFieldXML(field config.Field) string {
// 	if matched, _ := regexp.MatchString(`^(\.\.)?[a-zA-Z0-9_]+$`, field.Field); matched {

// 		switch field.Type {
// 		case "text":
// 			return fmt.Sprintf("$%s", field.Field)

// 		case "logical":
// 			return fmt.Sprintf("if $%s then 1 else 0", field.Field)

// 		case "date":
// 			return fmt.Sprintf("if $$IsEmpty:$%s then $$StrByCharCode:241 else $$PyrlYYYYMMDDFormat:$%s:\"-\"", field.Field, field.Field)

// 		case "number":
// 			return fmt.Sprintf("if $$IsEmpty:$%s then \"0\" else $$String:$%s", field.Field, field.Field)

// 		case "amount":
// 			return fmt.Sprintf("$$StringFindAndReplace:(if $$IsDebit:$%s then -$$NumValue:$%s else $$NumValue:$%s):\"(-)\":\"-\"", field.Field, field.Field, field.Field)

// 		case "quantity":
// 			return fmt.Sprintf("$$StringFindAndReplace:(if $$IsInwards:$%s then $$Number:$$String:$%s:\"TailUnits\" else -$$Number:$$String:$%s:\"TailUnits\"):\"(-)\":\"-\"", field.Field, field.Field, field.Field)

// 		case "rate":
// 			return fmt.Sprintf("if $$IsEmpty:$%s then 0 else $$Number:$%s", field.Field, field.Field)

// 		default:
// 			return field.Field
// 		}

// 	}

// 	return field.Field

// }

// func GenerateXMLfromYAML(tblConfig TableConfigYAML, config TallyConfig) (string, error) {
// 	fields := make([]Field, len(tblConfig.Fields))
// 	fieldNames := ""
// 	for i, f := range tblConfig.Fields {
// 		fieldName := fmt.Sprintf("Fld%02d", i+1)
// 		fieldNames += fieldName
// 		if i < len(tblConfig.Fields)-1 {
// 			fieldNames += ","
// 		}
// 		fields[i] = Field{
// 			Name:   fieldName,
// 			Set:    generateFieldXML(f),
// 			XMLTag: fmt.Sprintf("F%02d", i+1),
// 		}
// 	}

// 	envelope := Envelope{
// 		Header: Header{
// 			Version:      1,
// 			TallyRequest: "Export",
// 			Type:         "Data",
// 			ID:           "TallyDatabaseLoaderReport",
// 		},
// 		Body: Body{
// 			Desc: Desc{
// 				StaticVariables: StaticVariables{
// 					SVExportFormat:   "XML (Data Interchange)",
// 					SVFromDate:       config.FromDate,
// 					SVToDate:         config.ToDate,
// 					SVCurrentCompany: config.Company,
// 				},
// 				TDL: TDL{
// 					TDLMessage: TDLMessage{
// 						Report: Report{
// 							Name:  "TallyDatabaseLoaderReport",
// 							Forms: "MyForm",
// 						},
// 						Form: Form{
// 							Name:  "MyForm",
// 							Parts: "MyPart01",
// 						},
// 						Part: Part{
// 							Name:     "MyPart01",
// 							Lines:    "MyLine01",
// 							Repeat:   "MyLine01 : MyCollection",
// 							Scrolled: "Vertical",
// 						},
// 						Line: Line{
// 							Name:   "MyLine01",
// 							Fields: fieldNames,
// 						},
// 						Fields: fields,
// 						Collection: Collection{
// 							Name:   "MyCollection",
// 							Type:   tblConfig.Collection,
// 							Fetch:  join(tblConfig.Fetch, ","),
// 							Filter: join(tblConfig.Filters, ","),
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	output, err := xml.MarshalIndent(envelope, "", "  ")
// 	if err != nil {
// 		return "", err
// 	}
// 	return xml.Header + string(output), nil
// }

// func join(arr []string, sep string) string {
// 	result := ""
// 	for i, s := range arr {
// 		result += s
// 		if i < len(arr)-1 {
// 			result += sep
// 		}
// 	}
// 	return result
// }
