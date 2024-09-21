const fs = require('fs');

const CHAIN_LIST_URL = "https://raw.githubusercontent.com/DefiLlama/chainlist/main/constants/extraRpcs.js"
const START_STRING = "export const extraRpcs = "
const END_STRING = ";";

const extraChains = {
    "1329": [
        {
            "url": "https://evm-rpc.sei-apis.com"
        }
    ],
    "810180": [
        {
            "url": "https://rpc.zklink.io"
        }
    ],
    "7560": [
        {
            "url": "https://rpc.cyber.co"
        },
        {
            "url": "https://cyber.alt.technology"
        }
    ]
}

// https://github.com/DefiLlama/chainlist/blob/main/constants/chainIds.json
async function main() {
    const chainListUrlsResponse = await fetch(CHAIN_LIST_URL);

    if (!chainListUrlsResponse.ok) {
        console.error("Request to github failed");
        process.exit(1);
    }

    const fileContent = await chainListUrlsResponse.text();

    const jsObject = extractJsObject(fileContent);
    const chainsConfigs = parseJsObject(jsObject);

    const structuredRpcs = mapToStandardizedStructure(chainsConfigs);

    fs.writeFileSync("./public.json", JSON.stringify({
        subscriptions: {},
        rpcInfo: { ...extraChains, ...structuredRpcs }
    }, null, 2))
}


function mapToStandardizedStructure(chainsConfigs) {
    const structuredRpcs = {};

    for (const [chainId, { rpcs }] of Object.entries(chainsConfigs)) {

        let rpcsStructured = rpcs.map(objOrStr => {
            if (typeof objOrStr === 'string') {
                return { url: objOrStr };
            } else {
                return { url: objOrStr.url };
            }

            // remove wss 
        }).filter(({ url }) => url.startsWith("https://"));

        if (rpcsStructured.length === 0) {
            console.log(`No rpcs found for chainId=${chainId}, adding dummy rpc`);
            rpcsStructured = [{ url: "https://dummy.rpc" }]
        }

        structuredRpcs[chainId] = rpcsStructured;
    }
    return structuredRpcs
}

function extractJsObject(fileContent) {
    const startIndex = fileContent.indexOf(START_STRING) + START_STRING.length;
    const endIndex = startIndex + fileContent.slice(startIndex).indexOf(END_STRING);
    const jsObject = fileContent.slice(startIndex, endIndex);
    return jsObject;
}

function parseJsObject(jsObject) {
    const privacyStatement = {};
    let xd;
    const chainsConfigs = eval('xd = ' + jsObject);
    return chainsConfigs;
}


main()