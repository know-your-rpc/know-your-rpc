export async function getRequest(url, queryParams) {
    const response = await fetch(url + "?" + new URLSearchParams(queryParams), { method: "GET" })

    if (response.ok) {
        return await response.json();
    }

    throw new Error(`Request failed statusCode=${response.status}`)
}

