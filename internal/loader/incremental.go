package loader

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
)

var lastTallyAlterIdXml = `<?xml version="1.0" encoding="utf-8"?>
<ENVELOPE>
	<HEADER>
		<VERSION>1</VERSION>
		<TALLYREQUEST>Export</TALLYREQUEST>
		<TYPE>Data</TYPE>
		<ID>MyReport</ID>
	</HEADER>
	<BODY>
		<DESC>
			<STATICVARIABLES>
				<SVEXPORTFORMAT>ASCII (Comma Delimited)</SVEXPORTFORMAT>
			</STATICVARIABLES>
			<TDL>
				<TDLMESSAGE>
					<REPORT NAME="MyReport">
						<FORMS>MyForm</FORMS>
					</REPORT>
					<FORM NAME="MyForm">
						<PARTS>MyPart</PARTS>
					</FORM>
					<PART NAME="MyPart">
						<LINES>MyLine</LINES>
						<REPEAT>MyLine : MyCollection</REPEAT>
						<SCROLLED>Vertical</SCROLLED>
					</PART>
					<LINE NAME="MyLine">
						<FIELDS>FldAlterMaster,FldAlterTransaction</FIELDS>
					</LINE>
					<FIELD NAME="FldAlterMaster">
						<SET>$AltMstId</SET>
					</FIELD>
					<FIELD NAME="FldAlterTransaction">
						<SET>$AltVchId</SET>
					</FIELD>
					<COLLECTION NAME="MyCollection">
						<TYPE>Company</TYPE>
						<FILTER>FilterActiveCompany</FILTER>
					</COLLECTION>
					<SYSTEM TYPE="Formulae" NAME="FilterActiveCompany">
						<CONTENT>$$IsEqual:Dummy:$Name</CONTENT>
					</SYSTEM>
				</TDLMESSAGE>
			</TDL>
		</DESC>
	</BODY>
</ENVELOPE>`

func LastTallyAlterId() (int, int, error) {
	result, err := CallTallyApi(context.TODO(), []byte(lastTallyAlterIdXml))

	if err != nil {
		// handle error
		log.Println("Error calling Tally API:", err)
		return 0, 0, err
	}
	cleanedString := strings.ReplaceAll(result, `"`, "")
	log.Println("Last Alter IDs from Tally API:", cleanedString)

	splittedIds := strings.Split(cleanedString, ",")

	if len(splittedIds) < 2 {
		log.Println("Insufficient data received from Tally API")
		return 0, 0, fmt.Errorf("insufficient data received from Tally API")
	}

	masterAlterID, err := strconv.Atoi(splittedIds[0])
	if err != nil {
		log.Println("Error converting masterAlterID:", err)
		return 0, 0, err
	}

	transactionID, err := strconv.Atoi(splittedIds[1])
	if err != nil {
		log.Println("Error converting transactionID:", err)
		return 0, 0, err
	}

	log.Printf("Last Alter IDs - Master: %d, Transaction: %d", masterAlterID, transactionID)

	return masterAlterID, transactionID, nil
}
