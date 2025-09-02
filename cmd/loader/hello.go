package main

// import (
// 	"encoding/xml"
// 	"fmt"
// 	"regexp"
// 	"strings"
// 	"tally-connector/config"
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
// 	Report     Report      `xml:"REPORT"`
// 	Form       Form        `xml:"FORM"`
// 	Parts      []Part      `xml:"PART"`
// 	Lines      []Line      `xml:"LINE"`
// 	Fields     []Field     `xml:"FIELD"`
// 	Collection Collection  `xml:"COLLECTION"`
// 	Systems    []SystemDef `xml:"SYSTEM,omitempty"`
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
// 	Name    string `xml:"NAME,attr"`
// 	Fields  string `xml:"FIELDS"`
// 	Explode string `xml:"EXPLODE,omitempty"`
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

// type SystemDef struct {
// 	Type string `xml:"TYPE,attr"`
// 	Name string `xml:"NAME,attr"`
// 	Expr string `xml:",chardata"`
// }

// // --- helpers ---

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

// func join(arr []string, sep string) string {
// 	return strings.Join(arr, sep)
// }

// // --- main generator ---

// func GenerateXMLfromYAML(tblConfig TableConfigYAML, cfg TallyConfig) (string, error) {
// 	// Build real fields
// 	fields := make([]Field, len(tblConfig.Fields))
// 	fieldNames := ""
// 	for i, f := range tblConfig.Fields {
// 		fieldName := fmt.Sprintf("Fld%02d", i+1)
// 		if i > 0 {
// 			fieldNames += ","
// 		}
// 		fieldNames += fieldName

// 		fields[i] = Field{
// 			Name:   fieldName,
// 			Set:    generateFieldXML(f),
// 			XMLTag: fmt.Sprintf("F%02d", i+1),
// 		}
// 	}

// 	// Always add FldBlank for intermediate parts
// 	fields = append([]Field{
// 		{
// 			Name:   "FldBlank",
// 			Set:    "\"\"",
// 			XMLTag: "BLANK",
// 		},
// 	}, fields...)

// 	// Split collection path like Ledger.Entries.StockItems
// 	lstRoutes := strings.Split(tblConfig.Collection, ".")
// 	targetCollection := lstRoutes[0]
// 	if len(lstRoutes) > 1 {
// 		lstRoutes = append([]string{"MyCollection"}, lstRoutes[1:]...)
// 	} else {
// 		lstRoutes = []string{"MyCollection"}
// 	}

// 	parts := make([]Part, len(lstRoutes))
// 	lines := make([]Line, len(lstRoutes))

// 	for i, route := range lstRoutes {
// 		partName := fmt.Sprintf("MyPart%02d", i+1)
// 		lineName := fmt.Sprintf("MyLine%02d", i+1)

// 		parts[i] = Part{
// 			Name:     partName,
// 			Lines:    lineName,
// 			Repeat:   fmt.Sprintf("%s : %s", lineName, route),
// 			Scrolled: "Vertical",
// 		}

// 		if i < len(lstRoutes)-1 {
// 			// intermediate line uses FldBlank
// 			lines[i] = Line{
// 				Name:    lineName,
// 				Fields:  "FldBlank",
// 				Explode: fmt.Sprintf("MyPart%02d", i+2),
// 			}
// 		} else {
// 			// last line uses real fields
// 			lines[i] = Line{
// 				Name:   lineName,
// 				Fields: fieldNames,
// 			}
// 		}
// 	}

// 	// Build system filters if needed
// 	var systems []SystemDef
// 	filters := make([]string, len(tblConfig.Filters))
// 	for j, flt := range tblConfig.Filters {
// 		systems = append(systems, SystemDef{
// 			Type: "Formulae",
// 			Name: fmt.Sprintf("Fltr%02d", j+1),
// 			Expr: flt,
// 		})
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
// 					SVFromDate:       cfg.FromDate,
// 					SVToDate:         cfg.ToDate,
// 					SVCurrentCompany: cfg.Company,
// 				},
// 				TDL: TDL{
// 					TDLMessage: TDLMessage{
// 						Report: Report{
// 							Name:  "TallyDatabaseLoaderReport",
// 							Forms: "MyForm",
// 						},
// 						Form: Form{
// 							Name:  "MyForm",
// 							Parts: parts[0].Name, // root part
// 						},
// 						Parts:  parts,
// 						Lines:  lines,
// 						Fields: fields,
// 						Collection: Collection{
// 							Name:   "MyCollection",
// 							Type:   targetCollection,
// 							Fetch:  join(tblConfig.Fetch, ","),
// 							Filter: join(filters, ","),
// 						},
// 						Systems: systems,
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
