import { serve } from '@hono/node-server'
import { Hono } from 'hono'

import { logger } from 'hono/logger'
import fetchCompanies from './tally/fetch_companies'
import fetchLedgers from './tally/fetch_ledgers'
import fetchLedgerVouchers from './tally/fetch_vouchers'
const app = new Hono()

app.use(logger())


app.get('/fetch-companies', async (c) => {
	// Logic for fetching companies

	const companies = await fetchCompanies()

	return c.json({ companies })
})

app.get('/fetch-ledgers', async (c) => {
	const ledgers = await fetchLedgers()

	return c.json({ ledgers })
})
/*
	1. https://help.tallysolutions.com/integration-methods-and-technologies/
	2.  

*/

app.get('/create-voucher', (c) => {
	// Logic for creating a voucher
	return c.json({ message: 'Voucher created' })
})

app.get('/fetch-vouchers', async (c) => {
	// Logic for fetching vouchers
	const vouchers = await fetchLedgerVouchers()
	return c.json({ vouchers })
})

const main = async () => {
	const PORT = 9999
	serve({
		fetch: app.fetch,
		port: PORT
	})
	console.log(`Server running at http://localhost:${PORT}`)
}

main()
fetchLedgerVouchers()