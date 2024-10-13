import { getRequest, postRequest } from "./utils.js";

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



export async function payForSubscription() {
    const web3 = new Web3(window.ethereum);

    const transferParams = await getRequest("/api/payment/params");
    const _usdcContract = new web3.eth.Contract(ABI, transferParams.expectedToken);
    const accounts = await window.ethereum.request({ method: "eth_requestAccounts" });
    // Ensure the user is on the Ethereum mainnet
    const chainId = await web3.eth.getChainId();
    if (chainId !== transferParams.chainId) {
        try {
            await window.ethereum.request({
                method: 'wallet_switchEthereumChain',
                params: [{ chainId: `0x${transferParams.chainId.toString(16)}` }], // chainId must be in hexadecimal numbers
            });
        } catch (switchError) {
            throw new Error("Please switch to the Ethereum mainnet to perform this transaction.");
        }
    }
    console.log(transferParams);
    const txReceipt = await _usdcContract.methods.transfer(transferParams.expectedTo, transferParams.expectedValue).send({ from: accounts[0] });

    await postRequest("/api/payment/acknowledge", { txHash: txReceipt.transactionHash });
}


