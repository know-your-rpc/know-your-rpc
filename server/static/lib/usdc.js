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
    _usdcContract = new this.web3.eth.Contract(ABI, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")


    async requestTransfer(to, value) {
        const accounts = await this.web3.eth.requestAccounts();

        return this._usdcContract.methods.transfer(to, value).send({ from: accounts[0] });
    }

}

export const usdc = new USDC();