const { fetchTopNetworks } = require("./top-networks");
const fs = require("fs");

const SUPPORTED_CHAINS_PATH = "../writer/config/supported_chain.go";

async function main() {
  const topNetworks = (await fetchTopNetworks())
    .slice(0, 50)
    .map((c) => ({ ChainId: c.chainId, Name: c.name }));

  const currentSupportedChainsLines = fs
    .readFileSync(SUPPORTED_CHAINS_PATH)
    .toString()
    .replaceAll("ChainId", `"ChainId"`)
    .replaceAll("Name", `"Name"`)
    .split("\n");

  const startLineIndex = currentSupportedChainsLines.findIndex((v) =>
    v.includes("$$START$$"),
  );
  const endLineIndex = currentSupportedChainsLines.findIndex((v) =>
    v.includes("$$END$$"),
  );

  const currentSupportedChainIds = currentSupportedChainsLines
    .slice(startLineIndex + 1, endLineIndex)
    .map((line) => JSON.parse(line.slice(0, -1)));

  const topNetworksDedup = topNetworks.filter(
    (c) => !currentSupportedChainIds.some((c2) => c2.ChainId === c.ChainId),
  );

  const merged = [...currentSupportedChainIds, ...topNetworksDedup];

  merged.forEach((m) =>
    console.log(`{ChainId: "${m.ChainId}", Name: "${m.Name}"},`),
  );
}

void main();
