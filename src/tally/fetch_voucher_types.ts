import axios from "axios";
import { parseStringPromise } from "xml2js";
import myenv from "../config/env_config";

async function fetchVoucherTypes() {

	const fetchVoucherTypesXML = `
		<ENVELOPE>
			<HEADER>
				<VERSION>1</VERSION>
				<TALLYREQUEST>Export</TALLYREQUEST>
				<TYPE>Collection</TYPE>
				<ID>List of Voucher Types</ID>
			</HEADER>
			<BODY>
				<DESC>
					<TDL>
						<TDLMESSAGE>
							<COLLECTION NAME="List of Voucher Types" ISINITIALIZE="Yes">
								<TYPE>VoucherType</TYPE>
								<NATIVEMETHOD>NAME</NATIVEMETHOD>
							</COLLECTION>
						</TDLMESSAGE>
					</TDL>
				</DESC>
			</BODY>
		</ENVELOPE>
	`;

	const response = await axios.post(myenv.TALLY_URL, fetchVoucherTypesXML, {
		headers: { "Content-Type": "text/xml" },
	});

	// console.log("📩 Raw XML Response:\n", response.data);

	// Convert XML → JSON
	const json = await parseStringPromise(response.data, {
	});

	console.log("✅ Parsed JSON Response:\n", JSON.stringify(json, null, 2));
}