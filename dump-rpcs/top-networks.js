async function fetchTopNetworks() {
  const chains = await fetch("https://api.llama.fi/v2/chains").then((r) =>
    r.json(),
  );

  const sortedChains = chains.sort((a, b) => b.tvl - a.tvl);

  return sortedChains
    .filter((chain) => !!chain.chainId)
    .map((chain) => ({ chainId: String(chain.chainId), name: chain.name }));
}

module.exports = {
  fetchTopNetworks,
};
