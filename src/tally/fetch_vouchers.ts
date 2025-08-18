import axios from "axios";
import { parseStringPromise } from "xml2js";
import myenv from "../config/env_config";
import xml_configs from "./xml_config";

async function fetchLedgerVouchers() {

	const fetchLedgerVouchersXML = `
		<ENVELOPE>
			<HEADER>
				<VERSION>1</VERSION>
				<TALLYREQUEST>Export</TALLYREQUEST>
				<TYPE>Collection</TYPE>
				<ID>List of Ledger Vouchers</ID>
			</HEADER>
			<BODY>
				<DESC>
					<TDL>
						<TDLMESSAGE>
							<COLLECTION NAME="List of Ledger Vouchers" ISINITIALIZE="Yes">
								<TYPE>Voucher</TYPE>
								<NATIVEMETHOD>NAME</NATIVEMETHOD>
								<NATIVEMETHOD>LEDGERNAME</NATIVEMETHOD>
								<NATIVEMETHOD>AMOUNT</NATIVEMETHOD>
							</COLLECTION>
						</TDLMESSAGE>
					</TDL>
				</DESC>
			</BODY>
		</ENVELOPE>
	`;
	const l = `	<ENVELOPE>
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
</ENVELOPE>`


	const response = await axios.post(myenv.TALLY_URL, l, {
		headers: { "Content-Type": "text/xml" },
	});

	console.log()


	// Convert XML → JSON
	const json = await parseStringPromise(response.data, {
		...xml_configs
	})
	if (2 == 2) {
		return  json
	}



	console.log(json)
	const result = json?.BODY?.DATA?.COLLECTION?.VOUCHER;

	const transactions = result?.map((v: any) => ({
		// name: v.NAME._,
		ledgerName: v.LEDGERNAME._,
		// amount: v.AMOUNT._,

		voucher_type: v.VCHTYPE,
		voucher_type_name: v.VOUCHERTYPENAME,
		voucher_number: v.VOUCHERNUMBER,
		date: v.DATE._,
		effective_date: v.EFFECTIVEDATE?._


	})) || []


	return transactions

}
export default fetchLedgerVouchers;