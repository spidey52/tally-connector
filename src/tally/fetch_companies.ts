import axios from "axios";
import { parseStringPromise } from "xml2js";
import myenv from "../config/env_config";
import xml_configs from "./xml_config";

// XML request to get list of companies
const getCompaniesXML = `
   <ENVELOPE>
    <HEADER>
      <VERSION>1</VERSION>
      <TALLYREQUEST>Export</TALLYREQUEST>
      <TYPE>Collection</TYPE>
      <ID>List of Companies</ID>
    </HEADER>
    <BODY>
      <DESC>
        <TDL>
          <TDLMESSAGE>
            <COLLECTION NAME="List of Companies" ISMODIFY="No">
              <TYPE>Company</TYPE>
              <FETCH>NAME</FETCH>
            </COLLECTION>
          </TDLMESSAGE>
        </TDL>
      </DESC>
    </BODY>
  </ENVELOPE>
`;

async function fetchCompanies() {
  try {
    const response = await axios.post(myenv.TALLY_URL, getCompaniesXML, {
      headers: { "Content-Type": "text/xml" },
    });

    console.log("📩 Raw XML Response:\n", response.data);

    // Convert XML → JSON
    const json = await parseStringPromise(response.data, {
      ...xml_configs
    });
    console.log("✅ Parsed JSON Response:\n", JSON.stringify(json, null, 2));
    return json;
  } catch (err: any) {
    console.error("❌ Error connecting to Tally:", err.message);
  }
}


export default fetchCompanies;