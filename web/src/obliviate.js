import $ from 'jquery';
import nacl from 'tweetnacl';
import 'tweetnacl-util';
import scryptAsynch from 'scrypt-async';
import ClipboardJS from 'clipboard';

const encrypt = {
    secretKey: '',
    message: '',
    salt: '',
    password: false,
    time: 0,
    processEncrypt: function () {
        if ($("#passwordBlock").hasClass("collapsing")) {
            return;
        }
        encrypt.message = $('#message').val();
        if (encrypt.message.length === 0) {
            $("#message").addClass('is-invalid');
            return;
        }
        $("#message").removeClass('is-invalid');

        if ($("#passwordBlock").hasClass("show")) {
            const password = $('#encryptPassword').val();
            if (password.length > 0) {
                encodeButtonAccessibility(false);
                encrypt.password = true;
                encrypt.salt = nacl.randomBytes(nacl.secretbox.keyLength); // the same as key, 32 bytes
                calculateKeyDerived(password, encrypt.salt, scryptLogN, encrypt.scryptCallback);
                $('#encryptPassword').removeClass('is-invalid');
            } else {
                $('#encryptPassword').addClass('is-invalid');
            }
            return;
        } else {
            encodeButtonAccessibility(false);
        }
        encrypt.secretKey = nacl.randomBytes(nacl.secretbox.keyLength);
        encrypt.continue();
    },
    scryptCallback: function (key, time) {
        encrypt.secretKey = key;
        encrypt.time = time;
        encrypt.continue();
    },
    continue: function () {
        // encrypt message with nacl secretbox
        const messageUTF8 = nacl.util.decodeUTF8(encrypt.message);
        const messageNonce = nacl.randomBytes(nacl.secretbox.nonceLength);

        const encryptedMessage = nacl.secretbox(messageUTF8, messageNonce, encrypt.secretKey);

        // nonce will be used as a link anchor
        urlNonce = nacl.util.encodeBase64(messageNonce);

        // store secret key in the message
        const fullMessage = new Uint8Array(encrypt.secretKey.length + encryptedMessage.length);
        if (encrypt.password) {
            fullMessage.set(encrypt.salt);
        } else {
            fullMessage.set(encrypt.secretKey);
        }
        fullMessage.set(encryptedMessage, encrypt.secretKey.length);

        // encrypt message transmission with nacl box
        const transmissionNonce = nacl.randomBytes(nacl.box.nonceLength);
        const transmission = nacl.box(fullMessage, transmissionNonce, serverPublicKey, keys.secretKey);

        const obj = {};
        obj.message = nacl.util.encodeBase64(transmission);
        obj.nonce = nacl.util.encodeBase64(transmissionNonce);
        obj.hash = nacl.util.encodeBase64(nacl.hash(messageNonce));
        obj.publicKey = nacl.util.encodeBase64(keys.publicKey);
        if (encrypt.password) {
            obj.time = encrypt.time;
        }

        post('POST', obj, '/save', encrypt.encodeSuccess, encrypt.encodeError);
    },
    encodeSuccess: function encodeSuccess(result) {
        let index;
        if (encrypt.password) {
            index = queryIndexWithPassword;
        } else {
            index = 3;
        }
        if (!window.location.origin) { // IE fix
            window.location.origin = window.location.protocol + "//" + window.location.hostname +
                (window.location.port === 443 ? "" : ":" + window.location.port);
        }
        const url = window.location.origin + '/?' + urlNonce.substring(0, index) + "#" + urlNonce.substring(index, 32);
        $('#link').val(url);
        showLink();
    },
    encodeError: function (XMLHttpRequest, textStatus, errorThrown) {
        encodeButtonAccessibility(true);
        alert('{{.encryptNetworkError}}');
    }
};

const decrypt = {
    secretKey: '',
    salt: '',
    encodedMessage: '',
    password: false,
    hash: '',
    loadCypher: function () {
        decodeButtonAccessibility(false);
        const nonce = window.location.search.substring(1) + window.location.hash.substring(1);
        try {
            urlNonce = nacl.util.decodeBase64(nonce);
        } catch (ex) {
            decodeButtonAccessibility(true);
            alert('{{.linkIsCorrupted}}');
            return;
        }
        decrypt.hash = nacl.util.encodeBase64(nacl.hash(urlNonce));
        const obj = {};
        obj.hash = decrypt.hash;
        obj.publicKey = nacl.util.encodeBase64(keys.publicKey);
        if (decrypt.password) {
            obj.password = true;
        }

        post('POST', obj, '/read', decrypt.decryptTransmission, decrypt.loadError);
    },
    decryptTransmission: function (result) {
        // decode transmission with box
        const messageWithNonceAsUint8Array = nacl.util.decodeBase64(result.message);
        const noncePart = arraySlice(messageWithNonceAsUint8Array, 0, nacl.box.nonceLength);
        const messagePart = arraySlice(messageWithNonceAsUint8Array, nacl.box.nonceLength, result.message.length);

        const decrypted = nacl.box.open(messagePart, noncePart, serverPublicKey, keys.secretKey);
        if (!decrypted) {
            $('#decodedMessage').html("{{.generalError}}");
            showDecodedMessage();
            return
        }
        // decode message with secretbox
        if (decrypt.password) {
            decrypt.salt = arraySlice(decrypted, 0, nacl.secretbox.keyLength);
        } else {
            decrypt.secretKey = arraySlice(decrypted, 0, nacl.secretbox.keyLength);
        }
        decrypt.encodedMessage = arraySlice(decrypted, nacl.secretbox.keyLength, decrypted.length);
        decrypt.decryptMessage();
    },
    decryptMessage: function () {
        $("#decryptPassword").removeClass('is-invalid');
        decodeButtonAccessibility(false);
        if (decrypt.password) {
            const password = $('#decryptPassword').val();
            if (password.length > 0) {
                calculateKeyDerived(password, decrypt.salt, scryptLogN, decrypt.scryptCallback);
            } else {
                $("#decryptPassword").addClass('is-invalid');
                decodeButtonAccessibility(true);
                decrypt.changeAction();
            }
            return;
        }
        decrypt.continue();
    },
    scryptCallback: function (key, time) { // do nothing with time while decrypt
        decrypt.secretKey = key;
        decrypt.continue();
    },
    continue: function () {
        const messageBytes = nacl.secretbox.open(decrypt.encodedMessage, urlNonce, decrypt.secretKey);
        if (messageBytes == null) {
            if (decrypt.password) {
                $("#decryptPassword").addClass('is-invalid');
                decrypt.changeAction();
                decodeButtonAccessibility(true);
                return;
            }
            $('#decodedMessage').html("{{.generalError}}");
            showDecodedMessage(); // TODO: remove "Decoded message:" header
            return;
        }

        const message = nacl.util.encodeUTF8(messageBytes);

        const escape = document.createElement('textarea');
        escape.textContent = message;

        const str = replaceAll(escape.innerHTML, '\n', '<br/>');
        $('#decodedMessage').html(str);
        showDecodedMessage();

        if (decrypt.password) {
            const obj = {};
            obj.hash = decrypt.hash;
            post('DELETE', obj, '/delete', decrypt.deleteSuccess, decrypt.deleteError(obj));
        }

    },
    loadError: function (XMLHttpRequest, textStatus, errorThrown) {
        if (XMLHttpRequest.status === 404) {
            $("#decodeButtonBlock").addClass('d-none');
            $("#decryptPasswordBlock").addClass('d-none');
            $("#errorForDecodedMessage").removeClass('d-none');
            decodeButtonAccessibility(true);
        } else {
            decodeButtonAccessibility(true);
            alert('{{.decryptNetworkError}}')
        }
    },
    changeAction: function () {
        $("#decodeButton").off('click');
        $("#decodeButton").click(function (e) {
            decrypt.decryptMessage();
        });
    },
    deleteSuccess: function () { // do nothing
    },
    deleteError: function (obj) {
        return function (XMLHttpRequest, textStatus, errorThrown) {
            // try to delete again
            window.setTimeout(function () {
                post('DELETE', obj, '/delete?again', decrypt.deleteSuccess, decrypt.deleteErrorTryAgain);
            }, 1000);
        }
    },
    deleteErrorTryAgain: function (XMLHttpRequest, textStatus, errorThrown) {// do nothing
    }
};

// -------

function again() {
    keys = nacl.box.keyPair();
    showEnterMessage();
}

function post(method, webObject, url, postSuccess, postError) {
    $.ajax({
        url: url,
        type: method,
        dataType: "json",
        data: JSON.stringify(webObject),
        success: postSuccess,
        error: postError
    });
}

function replaceAll(str, find, replace) {
    return str.replace(new RegExp(find, 'g'), replace);
}

function arraySlice(arr, x, y) {
    if (!IE()) {
        return arr.slice(x, y);
    }
    // IE stuff
    let ret = [];
    for (let i = 0; i < arr.length; i++) {
        if (i >= x && i < y) {
            ret.push(arr[i]);
        }
    }
    return new Uint8Array(ret);
}

function calculateKeyDerived(password, salt, logN, callback) {
    let postpone = IE() ? 0 : 750;
    window.setTimeout(function () {
        try {
            const t1 = getTime();
            scryptAsynch(password, salt, {
                    logN: logN,
                    r: 8,
                    p: 1,
                    dkLen: nacl.secretbox.keyLength, // 32
                    interruptStep: 0,
                    encoding: 'binary' // hex, base64, binary
                },
                function (res) {
                    const time = Math.round(getTime() - t1);
                    callback(res, time);
                }
            );
        } catch (ex) {
            alert(ex.message); //TODO: process the exception
        }
    }, postpone); // it freezes the UI so postpone invocation
}

function IE() {
    return window.document.documentMode;
}

//debugger;

// ----- init
new ClipboardJS('.btn');
// const serverPublicKey = nacl.util.decodeBase64('{{.PublicKey}}');
const serverPublicKey = "nacl.util.decodeBase64('{{.PublicKey}}')";
let keys = nacl.box.keyPair();
let urlNonce = '';
const queryIndexWithPassword = 4;
const scryptLogN = 14;

const isMobile = window.matchMedia("only screen and (max-width: 760px)").matches;
if (isMobile) {
    $("#link").attr('rows', 2);
}

if (window.location.hash) {
    decrypt.password = window.location.search.substring(1).length === queryIndexWithPassword;
    showDecodeButton();
} else {
    showEnterMessage();
}

// necessary for .off('click')
$("#decodeButton").click(function (e) {
    decrypt.loadCypher();
});

if (IE()) {
    $("#ieEncryptWarning").removeClass('d-none');
    $("#ieDecryptWarning").removeClass('d-none');
}

var getTime = (function () {
    if (typeof performance !== "undefined") {
        return performance.now.bind(performance);
    }
    return Date.now.bind(Date);
})();

//--

function showEnterMessage() {
    $("#inputMessageBlock").removeClass('d-none');
    $("#linkBlock").addClass('d-none');
    $("#decodeBlock").addClass('d-none');
    $("#presentationBlock").addClass('d-none');

    $("#passwordBlock").removeClass('show');

    $("#message").focus();
    $('#link').val("");
    $('#encryptPassword').val("");
}

function showLink() {
    $("#inputMessageBlock").addClass('d-none');
    $("#linkBlock").removeClass('d-none');
    $("#decodeBlock").addClass('d-none');
    $("#presentationBlock").addClass('d-none');

    $("#message").val("");
    encodeButtonAccessibility(true);
}

function showDecodeButton() {
    if (decrypt.password) {
        $("#decryptPasswordBlock").removeClass('d-none');
    } else {
        $("#decryptPasswordBlock").addClass('d-none');
    }
    $("#inputMessageBlock").addClass('d-none');
    $("#linkBlock").addClass('d-none');
    $("#decodeBlock").removeClass('d-none');
    $("#presentationBlock").addClass('d-none');
}

function showDecodedMessage() {
    $("#inputMessageBlock").addClass('d-none');
    $("#linkBlock").addClass('d-none');
    $("#decodeBlock").addClass('d-none');
    $("#presentationBlock").removeClass('d-none');

    decodeButtonAccessibility(true);
}

function encodeButtonAccessibility(state) {
    if (state) {
        $("#encodeButton").removeClass('disabled');
        $("#encodeButtonSpinner").addClass('d-none');
    } else {
        $("#encodeButton").addClass('disabled');
        if (!IE()) {
            $("#encodeButtonSpinner").removeClass('d-none');
        }
    }
}

function decodeButtonAccessibility(state) {
    if (state) {
        $("#decodeButton").removeClass('disabled');
        $("#decodeButtonSpinner").addClass('d-none');
    } else {
        $("#decodeButton").addClass('disabled');
        if (!IE()) {
            $("#decodeButtonSpinner").removeClass('d-none');
        }
    }
}
