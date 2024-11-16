const { fetchTopNetworks, getSupportedChains } = require("./top-networks");
const fs = require("fs");

async function main() {
  const topNetworks = (await fetchTopNetworks())
    .slice(0, 50)
    .map((c) => ({ ChainId: c.chainId, Name: c.name }));

  const currentSupportedChain = getSupportedChains();

  const topNetworksDedup = topNetworks.filter(
    (c) => !currentSupportedChain.some((c2) => c2.ChainId === c.ChainId),
  );

  const merged = [...currentSupportedChain, ...topNetworksDedup];

  merged.forEach((m) =>
    console.log(`{ChainId: "${m.ChainId}", Name: "${m.Name}"},`),
  );
}

void main();
