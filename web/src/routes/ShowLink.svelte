<script>
    import ClipboardJS from 'clipboard'
    import {onMount, afterUpdate} from "svelte"

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
    export let link = ""
    let textarea

    onMount(() => {
        autoResize(textarea)
    })

    function autoResize(element) {
        element.style.height = '0px'
        const computed = window.getComputedStyle(element)
        const height = parseInt(computed.getPropertyValue('border-top-width'), 10)
            + parseInt(computed.getPropertyValue('border-bottom-width'), 10)
            + element.scrollHeight
        element.style.height = height + 'px'
    }

    new ClipboardJS('.btn')

    export let newMessageCallback = () => {
    }
</script>

<label for="link" class="text-secondary">{data.copyLink}</label>
<textarea class="form-control mb-3 bg-primary-subtle" id="link"
          bind:value={link}
          bind:this={textarea}
          readOnly/>

<div class="container">
    <div class="row">
        <div class="col-sm mb-2">
            <button type="button" class="btn btn-warning btn-lg w-100"
                    data-clipboard-action="copy"
                    data-clipboard-target="#link">{data.copyLinkButton}
            </button>
        </div>
        <div class="col-sm">
            <button type="button" class="btn btn-primary btn-lg w-100"
                    on:click={newMessageCallback}>{data.newMessageButton}
            </button>
        </div>
    </div>
</div>
