import { serve } from '@hono/node-server';
import { cors } from 'hono/cors';

import { createRoute, OpenAPIHono, z } from '@hono/zod-openapi';
import { Scalar } from '@scalar/hono-api-reference';
import dayjs from 'dayjs';
import { logger } from 'hono/logger';
import callApi from './tally/call_api';
import fetchCompanies from './tally/fetch_companies';
import fetchLedgerVouchers from './tally/fetch_ledger_vouchers';
import fetchLedgers from './tally/fetch_ledgers';

const app = new OpenAPIHono({})
app.use('*', cors())

app.use(logger())

app.doc("/docs", {
    openapi: "3.0.0",
    info: {
        title: "Tally Connector API",
        version: "1.0.0"
    },
})

app.get('/scalar', Scalar({
    url: '/docs',
    pageTitle: "Tally Connector Api"
}))


app.get('/fetch-companies', async (c) => {
    const companies = await fetchCompanies()
    return c.json({ companies })
})

const createGenericRoutes = (params: {
    url: string,
    method: 'get' | 'post' | 'put' | 'delete',
    query: z.ZodObject<any>
    description?: string
}) => {

    const route = createRoute({
        method: params.method,
        path: params.url,
        request: {
            query: params.query,
        },
        responses: {
            200: {
                description: params.description || "",
                content: {
                    "application/json": {
                        schema: z.object({
                            result: z.array(z.object({})),
                            total: z.number()
                        })
                    }
                }
            }
        }
    })

    return route
}

const defaultTallyQuery = z.object({
    startDate: z.string().default(dayjs().startOf('month').format('YYYY-MM-DD')),
    endDate: z.string().default(dayjs().format('YYYY-MM-DD')),
    ledgerName: z.string().min(3).max(100).default("DPAT9327"),
    raw: z.string().default("false")
})

const fetchLedgerRoutes = createGenericRoutes({
    url: "fetch-ledgers",
    method: 'get',
    query: defaultTallyQuery,
})

const fetchVoucherRoutes = createGenericRoutes({
    url: "fetch-vouchers",
    method: 'get',
    query: defaultTallyQuery,
})

const dayBookRoutes = createGenericRoutes({
    url: "fetch-day-book",
    method: 'get',
    query: defaultTallyQuery,
})

app.openapi(fetchLedgerRoutes, async (c) => {
    const { endDate, ledgerName, startDate, raw } = c.req.valid('query')
    const ledgers = await fetchLedgers({
        startDate,
        endDate,
        ledgerName,
        raw: raw === "true"
    })
    return c.json({ result: ledgers, total: ledgers.length })
})

app.openapi(fetchVoucherRoutes, async (c) => {
    const { endDate, ledgerName, startDate } = c.req.valid('query')
    const vouchers = await fetchLedgerVouchers({ ledgerName, startDate, endDate })
    return c.json({ result: vouchers, total: vouchers.length })
})

app.openapi(dayBookRoutes, async (c) => {
    const { endDate, ledgerName, startDate } = c.req.valid('query')
    const json = {
        "ENVELOPE": {
            "HEADER": {
                "TALLYREQUEST": "Export Data"
            },
            "BODY": {
                "EXPORTDATA": {
                    "REQUESTDESC": {
                        "STATICVARIABLES": {
                            "SVFROMDATE": startDate,
                            "SVTODATE": endDate,
                            "VOUCHERTYPENAME": ledgerName,
                            "VOUCHERNUMBER": "1",
                        },
                        "REPORTNAME": "Voucher Register"
                    }
                }
            }
        }
    }

    let xml = `<?xml version="1.0" encoding="utf-8"?>
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Data</TYPE>
        <ID>MyReportLedgerVouchers</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVFROMDATE>${startDate}</SVFROMDATE>
                <SVTODATE>${endDate}</SVTODATE>
                <SVEXPORTFORMAT>ASCII (Comma Delimited)</SVEXPORTFORMAT>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <REPORT NAME="MyReportLedgerVouchers">
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
                        <FIELDS>FldDate,FldVoucherType,FldVoucherNumber,FldLedger,FldAmount,FldNarration</FIELDS>
                    </LINE>
                    <FIELD NAME="FldDate">
                        <SET>$Date</SET>
                    </FIELD>
                    <FIELD NAME="FldVoucherType">
                        <SET>$VoucherTypeName</SET>
                    </FIELD>
                    <FIELD NAME="FldVoucherNumber">
                        <SET>$$StringFindAndReplace:$VoucherNumber:'"':'""'</SET>
                    </FIELD>
                    <FIELD NAME="FldLedger">
                        <SET>$$StringFindAndReplace:$FldLedger:'"':'""'</SET>
                    </FIELD>
                    <FIELD NAME="FldAmount">
                        <SET>$FldAmount</SET>
                    </FIELD>
                    <FIELD NAME="FldNarration">
                        <SET>$$StringFindAndReplace:$Narration:'"':'""'</SET>
                    </FIELD>
                    <COLLECTION NAME="MyCollection">
                        <TYPE>Voucher</TYPE>
                        <FETCH>Narration,AllLedgerEntries</FETCH>
                        <FILTER>FilterCancelledVouchers,FilterOptionalVouchers,FilterVch</FILTER>
                    </COLLECTION>
                    <SYSTEM TYPE="Formulae" NAME="FilterVch">NOT $$IsEmpty:($$FilterValue:$LedgerName:AllLedgerEntries:First:FilterVchLedger)</SYSTEM>
                    <SYSTEM TYPE="Formulae" NAME="FilterVchLedger">$$IsEqual:$LedgerName:"${ledgerName}"</SYSTEM>
                    <SYSTEM TYPE="Formulae" NAME="FilterVchLedgerNot">NOT $$IsEqual:$LedgerName:"${ledgerName}"</SYSTEM>
                    <SYSTEM TYPE="Formulae" NAME="FldAmount">if $$IsDr:$$FilterAmtTotal:AllLedgerEntries:FilterVchLedger:$Amount then (-$$FilterAmtTotal:AllLedgerEntries:FilterVchLedger:$Amount) else ($$FilterAmtTotal:AllLedgerEntries:FilterVchLedger:$Amount)</SYSTEM>
                    <SYSTEM TYPE="Formulae" NAME="FldLedger">if $$FilterCount:AllLedgerEntries:FilterVchLedgerNot > 1 then ($$FullList:AllLedgerEntries:$LedgerName) else ($$FilterValue:$LedgerName:AllLedgerEntries:First:FilterVchLedgerNot)</SYSTEM>
                    <SYSTEM TYPE="Formulae" NAME="FilterCancelledVouchers">NOT $IsCancelled</SYSTEM>
                    <SYSTEM TYPE="Formulae" NAME="FilterOptionalVouchers">NOT $IsOptional</SYSTEM>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>`
    // const daybook = await callApi(convertJsonToXml(json));
    const daybook = await callApi(xml);
    // const response = (await callApi(convertJsonToXml(voucherDetails)));

    return c.json({ result: [], total: 0, response: daybook })
})


// app.get('/call-api', async (c) => {
// 	const format = c.req.query('format') as 'html' | 'xml' || 'xml';
// 	const startDate = dayjs("2025-04-01").format('YYYYMMDD');
// 	const endDate = dayjs("2025-08-06").format('YYYYMMDD');
// 	const ledgerName = "DPAT9327";

// 	const xml = `	<ENVELOPE>
// <HEADER>
// <TALLYREQUEST>Export Data</TALLYREQUEST>
// </HEADER>
// <BODY>
// <EXPORTDATA>
// <REQUESTDESC>
// <STATICVARIABLES>
// <!--Specify the Period here-->
// <SVFROMDATE>${startDate}</SVFROMDATE>
// <SVTODATE>${endDate}</SVTODATE>
// <!--Specify the Voucher-type here-->
// <VOUCHERTYPENAME>SALES</VOUCHERTYPENAME>
// </STATICVARIABLES>
// <!--Specify the Report Name here-->
// <REPORTNAME>Voucher Register</REPORTNAME>
// </REQUESTDESC>
// </EXPORTDATA>
// </>
// </ENVELOPE>
// `
// 	const response = await callApi(xml, format)


// 	return c.render(response)

// })


// app.get('/create-voucher', (c) => {
// 	return c.json({ message: 'Voucher created' })
// })

// const fetchVoucherRoutes = createRoute({
// 	method: 'get',
// 	path: "/fetch-vouchers",
// 	request: {
// 		query: z.object({
// 			ledgerName: z.string().min(2).max(100),
// 			startDate: z.string().min(10).max(10),
// 			endDate: z.string().min(10).max(10),
// 		})
// 	},

// 	responses: {
// 		200: {
// 			description: "List of vouchers",
// 			content: {
// 				"application/json": {
// 					schema: z.object({
// 						vouchers: z.array(z.any())
// 					})
// 				},
// 			}
// 		}
// 	}
// })





const main = async () => {
    const PORT = 9999
    serve({
        fetch: app.fetch,
        port: PORT
    })
    console.log(`Server running at http://localhost:${PORT}`)
}

main()



