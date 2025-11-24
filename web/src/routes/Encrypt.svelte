<script>
    import {onMount} from "svelte"
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

    let message = ""
    let messageOk = true
    let textarea
    export let messagePassword = ""
    let hasPassword = messagePassword !== ""
    let passwordOk = true
    let buttonEncode = true
    let encodeSpinner = false
    let secretKey = new Uint8Array()
    let salt = new Uint8Array()
    let time = 0
    let urlNonce = ""

    function handleInputChange() {
        messageOk = message !== ""
    }

    function onChangePassword() {
        passwordOk = messagePassword !== ""
    }

    function onPasswordToggle() {
        hasPassword = !hasPassword
    }

    function processEncrypt() {
        if (message.length === 0) {
            messageOk = false
            return
        }

        if (hasPassword) {
            if (messagePassword.length > 0) {
                encodeButtonAccessibility(false)
                salt = nacl.randomBytes(nacl.secretbox.keyLength)  // the same as key, 32 bytes
                calculateKeyDerived(messagePassword, salt, CONSTANTS.costFactor, scryptCallback)
            } else {
                passwordOk = false
            }
            return
        } else {
            encodeButtonAccessibility(false)
        }
        secretKey = nacl.randomBytes(nacl.secretbox.keyLength)
        continueProcessing()
    }

    function scryptCallback(key, processingTime) {
        secretKey = key
        time = processingTime
        continueProcessing()
    }

    function continueProcessing() {
        const ephemeralKeys = nacl.box.keyPair()

        const messageNonce = nacl.randomBytes(nacl.secretbox.nonceLength)
        urlNonce = base64.encode(messageNonce)

        const encryptedMessage = nacl.secretbox(new TextEncoder().encode(message), messageNonce, secretKey)

        // store secret key in the message
        const fullMessage = new Uint8Array(secretKey.length + encryptedMessage.length)
        if (hasPassword) {
            fullMessage.set(salt)
        } else {
            fullMessage.set(secretKey)
        }
        fullMessage.set(encryptedMessage, secretKey.length)

        // encrypt message transmission with nacl box using ephemeral keys
        const transmissionNonce = nacl.randomBytes(nacl.box.nonceLength)
        const transmission = nacl.box(fullMessage, transmissionNonce, data.serverPublicKey, ephemeralKeys.secretKey)

        const obj = {}
        obj.message = base64.encode(transmission)
        obj.nonce = base64.encode(transmissionNonce)
        obj.hash = base64.encode(nacl.hash(messageNonce))
        obj.publicKey = base64.encode(ephemeralKeys.publicKey)
        if (hasPassword) {
            obj.time = time
            obj.costFactor = CONSTANTS.costFactor
        }
        post('POST', obj, CONSTANTS.SAVE_URL, encodeSuccess, encodeError)

        // Destroy ALL sensitive data in memory
        ephemeralKeys.secretKey.fill(0)
        ephemeralKeys.publicKey.fill(0)
        secretKey.fill(0)
        if (hasPassword) {
            salt.fill(0)
        }
        transmissionNonce.fill(0)
        messageNonce.fill(0)
        fullMessage.fill(0)
        transmission.fill(0)

        // Clear plaintext message
        message = ""
    }

    function encodeSuccess(result) {
        let index
        if (hasPassword) {
            index = CONSTANTS.queryIndexWithPassword
        } else {
            index = 3
        }
        const url = window.location.origin + '/?' + urlNonce.substring(0, index) + "#" + urlNonce.substring(index, 32)
        showLinkCallback(url)
    }

    function encodeError(err) {
        encodeButtonAccessibility(true)
        alert(data.encryptNetworkError)
    }

    function encodeButtonAccessibility(state) {
        buttonEncode = state
        encodeSpinner = !state
    }

    onMount(async () => {
        // console.log("encrypt: messagePassword: " + messagePassword)
        setTimeout(() => {
            textarea.focus();
        }, 0);
    })


    export let showLinkCallback = (url) => {
    }

</script>

<p class="text-secondary mb-2">{data.enterTextMessage}</p>
<div class="row">
    <div class="col">
        <textarea class={messageOk ? "form-control mb-3" : "form-control mb-3 is-invalid"}
                  rows="4" maxLength="262144"
                  bind:value={message}
                  bind:this={textarea}
                  on:input={handleInputChange}/>
    </div>
</div>
<div class="container">
    <div class="row">
        <div class={hasPassword ? "mb-3" : "mb-3 collapse"}>
            <div class="input-group">
                <span class="input-group-text">{data.password}</span>
                <input type="text"
                       class={passwordOk ? "form-control" : "form-control is-invalid"}
                       placeholder={data.enterPasswordPlaceholder}
                       bind:value={messagePassword}
                       on:input={onChangePassword}/>
            </div>
        </div>
    </div>
    <div class="row">
        <div class="col-sm mb-2">
            <button type="button" class="btn btn-success btn-lg w-100"
                    on:click={onPasswordToggle}>{data.password}
            </button>
        </div>
        <div class="col-sm">
            <button type="button"
                    class={buttonEncode ? "btn btn-danger btn-lg w-100" : "btn btn-danger btn-lg w-100 disabled"}
                    on:click={processEncrypt}>
                <span class={encodeSpinner ? "spinner-border spinner-border-sm" : "spinner-border spinner-border-sm d-none"}/>
                {data.secureButton}
            </button>
        </div>
    </div>
</div>