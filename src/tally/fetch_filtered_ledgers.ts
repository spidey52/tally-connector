import axios from "axios";
import { parseStringPromise } from "xml2js";
import myenv from "../config/env_config";

async function parseLedgerXML(xml: string) {
	const result = await parseStringPromise(xml, {
		explicitArray: false,   // don’t wrap every field in arrays
		mergeAttrs: true,       // merge XML attributes into parent object
		trim: true              // clean whitespace
	});

	// Navigate to ledger list
	const ledgers = result?.ENVELOPE?.BODY?.DATA?.COLLECTION?.LEDGER;

	if (!ledgers) {
		console.log("No ledgers found");
		return;
	}

	// Sometimes single ledger comes as object, convert to array
	const ledgerArray = Array.isArray(ledgers) ? ledgers : [ledgers];

	// Map to simplified objects
	const parsed = ledgerArray.map((l: any) => ({
		name: l.NAME,
		parent: l.PARENT?._,   // "Bank Accounts", "Cash-in-Hand" etc
		closingBalance: l.CLOSINGBALANCE?._,
		openingBalance: l.OPENINGBALANCE?._,
	}));

	return parsed;
}

async function fetchLedgers() {
	const fetchLedgersXML = `
	<ENVELOPE>
  <HEADER>
    <VERSION>1</VERSION>
    <TALLYREQUEST>Export</TALLYREQUEST>
    <TYPE>Data</TYPE>
    <ID>Ledger Vouchers</ID>
  </HEADER>
  <BODY>
    <EXPORTDATA>
      <REQUESTDESC>
        <REPORTNAME>Ledger Vouchers</REPORTNAME>
        <STATICVARIABLES>
          <!-- Replace with your exact company name -->
          <SVCURRENTCOMPANY>Dummy</SVCURRENTCOMPANY>
          <!-- Replace with the ledger whose account statement you want -->
          <LEDGERNAME>DPAT9327</LEDGERNAME>
          <!-- Date range in YYYYMMDD format -->
          <SVFROMDATE>20250401</SVFROMDATE>
          <SVTODATE>20260331</SVTODATE>
          <!-- Force XML output -->
          <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
        </STATICVARIABLES>
      </REQUESTDESC>
    </EXPORTDATA>
  </BODY>
</ENVELOPE>
	`;

	const response = await axios.post(myenv.TALLY_URL, fetchLedgersXML, {
		headers: { "Content-Type": "text/xml" },
	});



	// Convert XML → JSON
	const json = await parseLedgerXML(response.data);
	console.log("✅ Parsed JSON Response:\n", JSON.stringify(json, null, 2));
	return json;
}

export default fetchLedgers;