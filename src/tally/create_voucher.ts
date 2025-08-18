import axios from "axios";
import myenv from "../config/env_config";

async function createVoucher() {
   const xml = `
<ENVELOPE>
   <HEADER>
      <TALLYREQUEST>Import Data</TALLYREQUEST>
   </HEADER>
   <BODY>
      <IMPORTDATA>
         <REQUESTDESC>
            <REPORTNAME>Vouchers</REPORTNAME>
            <STATICVARIABLES>
               <SVCURRENTCOMPANY>Dummy</SVCURRENTCOMPANY>
            </STATICVARIABLES>
         </REQUESTDESC>
         <REQUESTDATA>
            <TALLYMESSAGE xmlns:UDF="TallyUDF">
               <VOUCHER VCHTYPE="Payment" ACTION="Create">
                  <DATE>20250801</DATE>
                  <GUID>12345-67890-ABCDE</GUID>
                  <NARRATION>Payment to Supplier</NARRATION>
                  <VOUCHERTYPENAME>Payment</VOUCHERTYPENAME>
                  <VOUCHERNUMBER>3</VOUCHERNUMBER>
                  <PARTYLEDGERNAME>Cash</PARTYLEDGERNAME>
                  <PERSISTEDVIEW>Accounting Voucher View</PERSISTEDVIEW>

                  <ALLLEDGERENTRIES.LIST>
                     <LEDGERNAME>Cash</LEDGERNAME>
                     <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
                     <AMOUNT>-8000</AMOUNT>
                  </ALLLEDGERENTRIES.LIST>

                  <ALLLEDGERENTRIES.LIST>
                     <LEDGERNAME>Cash</LEDGERNAME>
                     <ISDEEMEDPOSITIVE>No</ISDEEMEDPOSITIVE>
                     <AMOUNT>8000</AMOUNT>
                  </ALLLEDGERENTRIES.LIST>

               </VOUCHER>
            </TALLYMESSAGE>
         </REQUESTDATA>
      </IMPORTDATA>
   </BODY>
</ENVELOPE>

  `;

   try {
      const res = await axios.post(myenv.TALLY_URL, xml, {
         headers: { "Content-Type": "text/xml" },
      });

      console.log("Response:", res.data);
   } catch (err) {
      console.error("Error:", err);
   }
}

export default createVoucher;