import axios from "axios";
import xml2js from "xml2js";
import myenv from "../config/env_config";
import xml_configs from "./types";


async function callApi(xml: string, options: {} = {}) {
	try {
		console.log(xml)
		const res = await axios.post(myenv.TALLY_URL, xml, {
			headers: { "Content-Type": "text/xml" },
		});

		const xmlResponse = res.data;

		const jsonResponse = await xml2js.parseStringPromise(xmlResponse, {
			...xml_configs
		});
		console.log(xmlResponse)


		return jsonResponse

	} catch (err) {
		console.error("Error:", err);
	}
}

export default callApi;
