import axios from "axios";
import dayjs from "dayjs";
import { parseStringPromise } from "xml2js";
import myenv from "../config/env_config";
import { TallyParams } from "./types";

async function parseLedgerXML(xml: string, raw: boolean) {
	const result = await parseStringPromise(xml, {
		explicitArray: false,   // don’t wrap every field in arrays
		mergeAttrs: true,       // merge XML attributes into parent object
		trim: true              // clean whitespace
	});

	// Navigate to ledger list
	const ledgers = result?.ENVELOPE?.BODY?.DATA?.COLLECTION?.LEDGER;

	// if (2 == 2) {
	// 	return ledgers;
	// }
	if (raw) {
		return ledgers;
	}

	if (!ledgers) {
		console.log("No ledgers found");
		return [];
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



async function fetchLedgers(params: TallyParams = {}) {
	const startDate = params.startDate || dayjs().startOf('month').format('YYYYMMDD');
	const endDate = params.endDate || dayjs().format('YYYYMMDD');

	const fetchLedgersXML = `
		<ENVELOPE>
			<HEADER>
				<VERSION>1</VERSION>
				<TALLYREQUEST>Export</TALLYREQUEST>
				<TYPE>Collection</TYPE>
				<ID>List of Ledgers</ID>
			</HEADER>
			<BODY>
				<DESC>

				<STATICVARIABLES>
					<SVEXPORTFORMAT>$$SysName:xml</SVEXPORTFORMAT>

					<SVFROMDATE TYPE="Date">${startDate}</SVFROMDATE>
					<SVTODATE TYPE="Date">${endDate}</SVTODATE>
				</STATICVARIABLES>

					<TDL>
						<TDLMESSAGE>
							<COLLECTION NAME="List of Ledgers" ISINITIALIZE="Yes">
								<TYPE>Ledger</TYPE>
								<NATIVEMETHOD>NAME</NATIVEMETHOD>
								<NATIVEMETHOD>PARENT</NATIVEMETHOD>
								<NATIVEMETHOD>OPENINGBALANCE</NATIVEMETHOD>
								<NATIVEMETHOD>CLOSINGBALANCE</NATIVEMETHOD>
							</COLLECTION>
						</TDLMESSAGE>
					</TDL>
				</DESC>
			</BODY>
		</ENVELOPE>
	`;

	// <ENVELOPE>
	// 	<HEADER>
	// 	<TALLYREQUEST>Export Data </TALLYREQUEST>
	// 		</HEADER>
	// 		< BODY >
	// 		<EXPORTDATA>
	// 		<REQUESTDESC>
	// 		<STATICVARIABLES>
	// 		<!--Specify the FROM DATE here-- >
	// 			<SVFROMDATE>20080401 </SVFROMDATE>
	// 			< !--Specify the TO DATE here-- >
	// 				<SVTODATE>20090331 </SVTODATE>
	// 				< SVEXPORTFORMAT > $$SysName: XML </>
	// 					< !--Specify the LedgerName here-- >
	// 						<LEDGERNAME>CASH </LEDGERNAME>
	// 						</STATICVARIABLES>
	// 						< !--Report Name-- >
	// 							<REPORTNAME>Ledger Vouchers </REPORTNAME>
	// 								</REQUESTDESC>
	// 								</EXPORTDATA>
	// 								</BODY>
	// 								</ENVELOPE>




	const response = await axios.post(myenv.TALLY_URL, fetchLedgersXML, {
		headers: { "Content-Type": "text/xml" },
	});

	// Convert XML → JSON
	const json = await parseLedgerXML(response.data, params.raw || false);
	console.log("✅ Parsed JSON Response:\n", JSON.stringify(json, null, 2));
	return json;
}

export default fetchLedgers;