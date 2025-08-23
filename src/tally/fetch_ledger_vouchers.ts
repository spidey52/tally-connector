import axios from "axios";
import dayjs from "dayjs";
import { Builder, parseStringPromise } from "xml2js";
import myenv from "../config/env_config";
import xml_configs, { TallyParams } from "./types";


export type LedgerVoucherXml = {
	DSPVCHDATE: string[]
	DSPVCHLEDACCOUNT: string[]
	DSPVCHTYPE: string[]
	DSPVCHDRAMT: string[]
	DSPVCHCRAMT: string[]
	DSPEXPLVCHNUMBER: string[]
	DSPVCHLEDBALANCE: string[]
}

type LedgerVoucherJson = {
	date: string
	ledger_account: string;
	vch_type: string;
	debit_amount: string;
	credit_amount: string;
	voucher_no: string;
	ledger_balance: string;
}

const transformLedgerVoucher = (xml: LedgerVoucherXml) => {
	const len = xml.DSPVCHDATE.length;

	const result: LedgerVoucherJson[] = [];

	for (let i = 0; i < len; i++) {
		result.push({
			date: xml.DSPVCHDATE[i],
			ledger_account: xml.DSPVCHLEDACCOUNT[i],
			vch_type: xml.DSPVCHTYPE[i],
			debit_amount: xml.DSPVCHDRAMT[i],
			credit_amount: xml.DSPVCHCRAMT[i],
			voucher_no: xml.DSPEXPLVCHNUMBER[i],
			ledger_balance: xml.DSPVCHLEDBALANCE[i],
		});
	}

	return result;
}

async function fetchLedgerVouchers(params: TallyParams = {}) {
	const startDate = params.startDate || dayjs("2025-04-01").format('YYYYMMDD');
	const endDate = params.endDate || dayjs().format('YYYYMMDD');

	console.log(startDate, endDate, params.ledgerName);
	const jsonVal = {
		ENVELOPE: {
			HEADER: {
				TALLYREQUEST: "Export Data"
			},
			BODY: {
				EXPORTDATA: {
					REQUESTDESC: {
						STATICVARIABLES: {
							SVFROMDATE: startDate,
							SVTODATE: endDate,
							SVEXPORTFORMAT: "$$SysName:XML",
							LEDGERNAME: params.ledgerName || "SALES",
							SHOWRUNBALANCE: "YES",
							EXPLODEVNUM: "YES",
						},
						REPORTNAME: "Ledger Vouchers"
					}
				}
			}
		}
	}
	const builder = new Builder(xml_configs);
	let fetchLedgerVouchersXML = builder.buildObject(jsonVal);


	const response = await axios.post(myenv.TALLY_URL, fetchLedgerVouchersXML, {
		headers: { "Content-Type": "text/xml" },
	});

	const json = await parseStringPromise(response.data, {
		...xml_configs
	})

	// console.log(response.data)
// return json;

	return transformLedgerVoucher(json)
}
export default fetchLedgerVouchers;