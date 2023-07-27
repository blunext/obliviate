import {CONSTANTS} from './Commons.js'
import * as base64 from '@stablelib/base64'

export const prerender = true

let vars = {
    serverPublicKey: new Uint8Array(),
    copyLink: "",
    copyLinkButton: "",
    decodedMessage: "",
    decryptNetworkError: "",
    description: "",
    encryptNetworkError: "",
    enterTextMessage: "",
    generalError: "",
    header: "",
    info: "",
    info1: "",
    info2: "",
    info3: "",
    infoHeader: "",
    linkIsCorrupted: "",
    messageRead: "",
    newMessageButton: "",
    password: "",
    enterPasswordPlaceholder: "",
    readMessageButton: "",
    secureButton: "",
    title: ""
}

/** @type {import('./$types').PageLoad} */
export async function load({fetch}) {
    try {
        const res = await fetch(CONSTANTS.VARIABLES_URL)
        if (!res.ok) {
            throw new Error("Server error, status: " + res.status)
        }

        const data = await res.json()
        vars = data
        vars.serverPublicKey = base64.decode(data.PublicKey)
        return vars
    } catch (error) {
        console.error("Network error: ", error)
        alert("Something went wrong. Try again.")
    }
}