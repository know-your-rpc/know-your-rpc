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
}

// use this: https://chainid.network/chains.json
// https://github.com/DefiLlama/chainlist/blob/main/constants/chainIds.json
async function main() {
    const rpcsFromChainsJson = await fetchFromChainsJson();

    const rpcsFromChainList = await getRpcsFromChainList();

    const rpcs = {};

    const allChainIds = new Set([...Object.keys(rpcsFromChainsJson), ...Object.keys(rpcsFromChainList)]);

    for (const chainId of allChainIds) {
        if (rpcsFromChainList[chainId] && rpcsFromChainsJson[chainId]) {
            rpcs[chainId] = [...new Set([...rpcsFromChainList[chainId], ...rpcsFromChainsJson[chainId]]).values()].map(url => ({ url }));
        } else if (rpcsFromChainList[chainId]) {
            rpcs[chainId] = rpcsFromChainList[chainId].map(rpc => ({ url: rpc }));
        } else {
            rpcs[chainId] = rpcsFromChainsJson[chainId].map(rpc => ({ url: rpc }));
        }
    }

    fs.writeFileSync("./public.json", JSON.stringify({
        subscriptions: {},
        rpcInfo: { ...extraChains, ...rpcs }
    }, null, 2))
}


async function fetchFromChainsJson() {
    const chainsJsonResponse = await fetch("https://chainid.network/chains.json");
    if (!chainsJsonResponse.ok) {
        throw new Error("Request to chains.json failed");
    }
    const data = await chainsJsonResponse.json();

    const structuredRpcs = {};

    for (const chain of data) {
        let filteredRpcs = chain.rpc.filter(rpc => rpc.startsWith("https://") && !rpc.includes("${"));

        // we ignore rpcs with only single rpc (too many)
        if (filteredRpcs.length < 2) {
            console.log(`No rpcs found for chainId=${chain.chainId}`);
            continue;
        }

        structuredRpcs[chain.chainId] = filteredRpcs;
    }

    return structuredRpcs;
}

async function getRpcsFromChainList() {
    const chainListUrlsResponse = await fetch(CHAIN_LIST_URL);

    if (!chainListUrlsResponse.ok) {
        console.error("Request to github failed");
        throw new Error("Request to github failed");
    }

    const fileContent = await chainListUrlsResponse.text();

    const jsObject = extractJsObject(fileContent);
    const chainsConfigs = parseJsObject(jsObject);

    const structuredRpcs = mapToStandardizedStructure(chainsConfigs);
    return structuredRpcs;
}

function mapToStandardizedStructure(chainsConfigs) {
    const structuredRpcs = {};

    for (const [chainId, { rpcs }] of Object.entries(chainsConfigs)) {

        let rpcsStructured = rpcs.map(objOrStr => {
            if (typeof objOrStr === 'string') {
                return objOrStr;
            } else {
                return objOrStr.url;
            }

            // remove wss 
        }).filter(url => url.startsWith("https://"));

        if (rpcsStructured.length === 0) {
            console.log(`No rpcs found for chainId=${chainId}`);
            continue;
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