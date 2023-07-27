<script>
    import Encrypt from "./Encrypt.svelte"
    import ShowLink from "./ShowLink.svelte"
    import Decrypt from "./Decrypt.svelte"
    import ShowMessage from "./ShowMessage.svelte"
    import {onMount} from 'svelte'

    export let data
    const parts = {ENCRYPT: 0, LINK: 1, DECRYPT: 2, SHOW: 3}
    let visible = parts.ENCRYPT
    let link = ""
    let hash = ""
    let message = ""

    function showLinkCallback(url = "") {
        // console.log("linkCallback: ", url)
        link = url
        visible = parts.LINK
    }

    function newMessageCallback() {
        // console.log("newMessageCallback")
        hash = ""
        visible = parts.ENCRYPT
    }

    function messageCallback(msg) {
        // console.log("messageCallback: ", msg)
        message = msg
        visible = parts.SHOW
    }

    onMount(() => {
        if (window.onSvelteReady) {
            window.onSvelteReady()
        }
        hash = window.location.hash
        if (hash) {
            visible = parts.DECRYPT
        }
    })
</script>

<h4 class="text-secondary text-center mt-2">{data.header}</h4>
<div class="container border border-primary-subtle rounded">
    <div class="mt-2 mb-2">
        {#if visible === parts.ENCRYPT}
            <Encrypt {data} {showLinkCallback}/>
        {:else if visible === parts.LINK}
            <ShowLink {data} {link} {newMessageCallback}/>
        {:else if visible === parts.DECRYPT}
            <Decrypt {data} {hash} {newMessageCallback} {messageCallback}/>
        {:else if visible === parts.SHOW}
            <ShowMessage {data} {message} {newMessageCallback}/>
        {/if}
    </div>

    <div class="container mt-3">
        <div class="row">
            <div class="col-sm-1"/>
            <div class="col">
                <hr/>
            </div>
            <div class="col-auto fw-lighter"><small>{data.infoHeader}</small></div>
            <div class="col">
                <hr/>
            </div>
            <div class="col-sm-1"/>
        </div>
        <div class="row">
            <div class="col-sm-1"/>
            <div class="col">
                <p class="fw-light">
                    <small>{data.info}
                        <a href="https://github.com/blunext/obliviate" target="_blank"
                           rel="noopener noreferrer">GitHub</a>{data.info1}
                        <a href="mailto:info@securenote.io" target="_blank"
                           rel="noopener noreferrer">{data.info2}</a>{data.info3}
                    </small>
                </p>
            </div>
            <div class="col-sm-1"/>
        </div>
    </div>
</div>
