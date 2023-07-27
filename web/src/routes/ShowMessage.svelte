<script>
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
    export let message = ""

    export let newMessageCallback = () => {
    }
</script>

<div class="container">
    <div class="row">
        <div class="col-sm text-secondary">
            <p class="text-center">{data.decodedMessage}:</p>
            <div class="border-top my-3"/>
            <p class="br-line">{message}</p>
        </div>
    </div>
    <div class="row">
        <div class="col-sm">
            <button type="button" class="btn btn-primary btn-lg w-100"
                    on:click={newMessageCallback}>{data.newMessageButton}
            </button>
        </div>
    </div>
</div>

<style>
    .br-line {
        white-space: pre-wrap;
    }
</style>