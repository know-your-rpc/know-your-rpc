const fs = require("fs");

async function fetchTopNetworks() {
  const chains = await fetch("https://api.llama.fi/v2/chains").then((r) =>
    r.json(),
  );

  const sortedChains = chains.sort((a, b) => b.tvl - a.tvl);

  return sortedChains
    .filter((chain) => !!chain.chainId)
    .map((chain) => ({ chainId: String(chain.chainId), name: chain.name }));
}

function getSupportedChains() {
  const SUPPORTED_CHAINS_PATH = "../writer/config/supported_chain.go";
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

  const currentSupportedChains = currentSupportedChainsLines
    .slice(startLineIndex + 1, endLineIndex)
    .map((line) => JSON.parse(line.slice(0, -1)));

  return currentSupportedChains;
}

module.exports = {
  fetchTopNetworks,
  getSupportedChains,
};
