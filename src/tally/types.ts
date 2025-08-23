import { ParserOptions } from "xml2js";

export type TallyParams = {
	startDate?: string;
	endDate?: string;
	ledgerName?: string;
	raw?: boolean
}

const xml_configs: ParserOptions = {
	explicitArray: false,   // don’t wrap every field in arrays
	mergeAttrs: true,       // merge XML attributes into parent object
	trim: true,              // clean whitespace

	explicitRoot: false, // do not wrap the root element in an object
}

export default xml_configs;