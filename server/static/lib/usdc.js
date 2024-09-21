const ABI = [{
    "type": "function",
    "name": "transfer",
    "inputs": [
        {
            "name": "to",
            "type": "address"
        },
        {
            "name": "value",
            "type": "uint256"
        }
    ],
    "outputs": [
        {
            "name": "",
            "type": "bool"
        }
    ]
}];

class USDC {
    web3 = new Web3(window.ethereum);
    // 0xa9059cbb00000000000000000000000069df8f2010843da5bfe3df08ab769940764bb64f00000000000000000000000000000000000000000000000000000000041cdb40
    _usdcContract = new this.web3.eth.Contract(ABI, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")


    async requestTransfer(to, value) {
        const accounts = await window.ethereum.request({ method: "eth_requestAccounts" });
        // Ensure the user is on the Ethereum mainnet
        const chainId = await this.web3.eth.getChainId();
        if (chainId !== 1) {
            try {
                await window.ethereum.request({
                    method: 'wallet_switchEthereumChain',
                    params: [{ chainId: '0x1' }], // chainId must be in hexadecimal numbers
                });
            } catch (switchError) {
                throw new Error("Please switch to the Ethereum mainnet to perform this transaction.");
            }
        }
        return this._usdcContract.methods.transfer(to, value).send({ from: accounts[0] });
    }

}

export const usdc = new USDC();