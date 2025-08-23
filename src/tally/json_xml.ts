import xml2js from "xml2js";


const convertJsonToXml = (json: any) => {
	const builder = new xml2js.Builder();
	return builder.buildObject(json);
};


export default convertJsonToXml;
