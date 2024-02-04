<script>
    import nacl from "tweetnacl"
    import * as base64 from '@stablelib/base64'
    import {CONSTANTS, calculateKeyDerived, post} from './Commons.js'

    export let data = {
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
    let messageReadInfo = false

    const hasPassword = window.location.search.substring(1).length === CONSTANTS.queryIndexWithPassword
    let messagePasswordOk = true
    let decodeButton = true
    let decodeButtonSpinner = false
    let loadCypherAction = true
    let urlCryptoData = {urlNonce: new Uint8Array(), hash: ""}
    let salt = new Uint8Array()
    let costFactor = CONSTANTS.costFactor
    let secretKey = new Uint8Array()
    let encodedMessage = new Uint8Array()
    let cypherLoaded = false
    let messagePassword = ""
    let cypherReady = 0

    function decrypt() {
        // console.log("decrypt")
        if (loadCypherAction) {
            loadCypher()
        } else {
            decryptMessage()
        }
    }

    function loadCypher() {
        // console.log("loadCypher")

        decodeButtonAccessibility(false)

        const keys = nacl.box.keyPair()
        const nonce = window.location.search.substring(1) + window.location.hash.substring(1)

        let urlNonce = new Uint8Array()
        try {
            urlNonce = base64.decode(nonce)
        } catch (ex) {
            decodeButtonAccessibility(true)
            alert(data.linkIsCorrupted)
            return
        }

        const hash = base64.encode(nacl.hash(urlNonce))
        urlCryptoData = {urlNonce, hash}

        const obj = {}
        obj.hash = base64.encode(nacl.hash(urlNonce))
        obj.publicKey = base64.encode(keys.publicKey)
        if (hasPassword) {
            obj.password = true
        }

        post('POST', obj, CONSTANTS.READ_URL, decryptTransmission, loadError)

        function decryptTransmission(result) {
            // console.log("decryptTransmission: " + result)

            // decode transmission with box
            const messageWithNonceAsUint8Array = base64.decode(result.message)
            const noncePart = messageWithNonceAsUint8Array.slice(0, nacl.box.nonceLength)
            const messagePart = messageWithNonceAsUint8Array.slice(nacl.box.nonceLength, result.message.length)

            const decrypted = nacl.box.open(messagePart, noncePart, data.serverPublicKey, keys.secretKey)
            if (!decrypted) {
                decodeButtonAccessibility(true)
                alert(data.generalError)
                return
            }
            // decode message with secretbox
            if (hasPassword) {
                salt = decrypted.slice(0, nacl.secretbox.keyLength)
                costFactor = result.costFactor
            } else {
                secretKey = decrypted.slice(0, nacl.secretbox.keyLength)
            }
            encodedMessage = decrypted.slice(nacl.secretbox.keyLength, decrypted.length)
            cypherLoaded = true
        }
    }

    $: if (cypherLoaded) {
        decryptMessage()
    }

    $: if (cypherReady > 0) {
        decryptCypher()
    }

    function decryptMessage() {
        // console.log("decryptMessage")
        decodeButtonAccessibility(false)
        if (hasPassword) {
            if (messagePassword.length > 0) {
                calculateKeyDerived(messagePassword, salt, costFactor, scryptCallback)
            } else {
                messagePasswordOk = false
                decodeButtonAccessibility(true)
                loadCypherAction = false
            }
            return
        }
        cypherReady++

        function scryptCallback(key, time) { // do nothing with time while decrypt
            secretKey = key
            cypherReady++
        }
    }

    function decryptCypher() {
        const messageBytes = nacl.secretbox.open(encodedMessage, urlCryptoData.urlNonce, secretKey)
        if (messageBytes == null) {
            if (hasPassword) {
                loadCypherAction = false
                messagePasswordOk = false
                decodeButtonAccessibility(true)
                return
            }
            decodeButtonAccessibility(true)
            alert(data.generalError)
            return
        }

        const message = new TextDecoder('utf-8').decode(messageBytes)
        messageCallback(message, messagePassword)

        if (hasPassword) {
            const obj = {}
            obj.hash = urlCryptoData.hash
            post('DELETE', obj, CONSTANTS.DELETE_URL, doNothing, deleteError(obj))
        }

    }

    function loadError(err) {
        // console.log("loadError: " + err)
        decodeButtonAccessibility(true)
        if (err.response !== undefined && err.response.status === 404) {
            messageReadInfo = true
            messageReadInfo = false //hide pass
        } else {
            alert(data.decryptNetworkError)
        }
    }

    function doNothing() { // do nothing
    }

    function deleteError(obj) {
        return function (XMLHttpRequest, textStatus, errorThrown) {
            // try to delete again
            window.setTimeout(function () {
                post('DELETE', obj, '/delete?again', doNothing, doNothing)
            }, 1000)
        }
    }

    function decodeButtonAccessibility(state) {
        decodeButton = state
        decodeButtonSpinner = !state
    }

    function updatePassword(event) {
        // console.log("updatePassword: ", event.target.value)
        if (event.target.value.length === 0) {
            messagePasswordOk = false
        } else {
            messagePasswordOk = true
        }
    }

    export let newMessageCallback = () => {
    }
    export let messageCallback = (message, messagePassword) => {
    }
</script>

<div class="container">
    {#if messageReadInfo}
        <div class="row">
            <div class="col-sm">
                <p class="text-secondary">{data.messageRead}</p>
            </div>
        </div>
    {:else}
        {#if hasPassword}
            <div class="row">
                <div class="input-group mb-3">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">{data.password}</span>
                        </div>
                        <input type="text"
                               class="form-control {messagePasswordOk ? '' : 'is-invalid'}"
                               placeholder={data.enterPasswordPlaceholder}
                               on:input={updatePassword}
                               bind:value={messagePassword}
                        />
                    </div>
                </div>
            </div>
        {/if}
        <div class="row">
            <div class="col-sm mb-2">
                <button type="button" on:click={decrypt}
                        class="btn btn-danger btn-lg w-100 {decodeButton ? '' : 'disabled'}">
                    <span class="spinner-border spinner-border-sm {decodeButtonSpinner ? '' : 'd-none'}"/>
                    {data.readMessageButton}
                </button>
            </div>
            <div class="col-sm">
                <button type="button" class="btn btn-primary btn-lg w-100"
                        on:click={newMessageCallback}>{data.newMessageButton}
                </button>
            </div>
        </div>
    {/if}
</div>