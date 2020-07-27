import React, {useEffect, useState} from 'react';
import nacl from "tweetnacl";
import naclutil from "tweetnacl-util";
import {libs} from "./commons";
import $ from "jquery";

function Decrypt(props) {

    console.log("Decrypt start");

    const hasPassword = window.location.search.substring(1).length === libs.queryIndexWithPassword;

    const [decodeButton, setDecodeButton] = useState(true);
    const [decodeButtonSpinner, setDecodeButtonSpinner] = useState(false);
    const [loadCypherAction, setLoadCypherAction] = useState(true);
    const [urlCryptoData, setUrlCryptoData] = useState({urlNonce: '', hash: []});
    const [salt, setSalt] = useState('');
    const [secretKey, setSecretKey] = useState('');
    const [cypherLoaded, setCypherLoaded] = useState(false);
    const [cypherReady, setCypherReady] = useState(false);
    const [encodedMessage, setEncodedMessage] = useState(false);
    const [messagePassword, setMessagePassword] = useState('');
    const [messagePasswordOk, setMessagePasswordOk] = useState(true);

    function decrypt() {
        if (loadCypherAction) {
            loadCypher();
        } else {
            decryptMessage();
        }
    }

    function loadCypher() {
        console.log("loadCypher");

        decodeButtonAccessibility(false);

        const keys = nacl.box.keyPair();
        const nonce = window.location.search.substring(1) + window.location.hash.substring(1);

        let urlNonce = '';
        try {
            urlNonce = naclutil.decodeBase64(nonce);
        } catch (ex) {
            decodeButtonAccessibility(true);
            alert(props.var.linkIsCorrupted);
            return;
        }

        const hash = naclutil.encodeBase64(nacl.hash(urlNonce));
        setUrlCryptoData({urlNonce: urlNonce, hash: hash});

        const obj = {};
        obj.hash = naclutil.encodeBase64(nacl.hash(urlNonce));
        obj.publicKey = naclutil.encodeBase64(keys.publicKey);
        if (hasPassword) {
            obj.password = true;
        }

        libs.post('POST', obj, libs.READ_URL, decryptTransmission, loadError);

        function decryptTransmission(result) {
            // decode transmission with box
            const messageWithNonceAsUint8Array = naclutil.decodeBase64(result.message);
            const noncePart = libs.arraySlice(messageWithNonceAsUint8Array, 0, nacl.box.nonceLength);
            const messagePart = libs.arraySlice(messageWithNonceAsUint8Array, nacl.box.nonceLength, result.message.length);

            const decrypted = nacl.box.open(messagePart, noncePart, props.var.serverPublicKey, keys.secretKey);
            if (!decrypted) {
                $('#decodedMessage').html("{{.generalError}}");
                showDecodedMessage();
                return
            }
            // decode message with secretbox
            if (hasPassword) {
                setSalt(libs.arraySlice(decrypted, 0, nacl.secretbox.keyLength));
            } else {
                setSecretKey(libs.arraySlice(decrypted, 0, nacl.secretbox.keyLength));
            }
            setEncodedMessage(libs.arraySlice(decrypted, nacl.secretbox.keyLength, decrypted.length));

            setCypherLoaded(true);
        }
    }

    useEffect(() => {
        if (cypherLoaded) {
            decryptMessage();
        }
    }, [cypherLoaded])

    useEffect(() => {
        if (cypherReady) {
            decryptCypher();
        }
    }, [cypherReady])

    function decryptMessage() {
        debugger;
        // $("#decryptPassword").removeClass('is-invalid');
        decodeButtonAccessibility(false);
        if (hasPassword) {
            // const password = $('#decryptPassword').val();
            if (messagePassword.length > 0) {
                libs.calculateKeyDerived(messagePassword, salt, libs.scryptLogN, scryptCallback);
            } else {
                setMessagePasswordOk(false);
                decodeButtonAccessibility(true);
                setLoadCypherAction(false);
            }
            return;
        }
        setCypherReady(true);

        function scryptCallback(key, time) { // do nothing with time while decrypt
            setSecretKey(key);
            setCypherReady(true);
        }
    }

    function decryptCypher() {
        debugger;
        const messageBytes = nacl.secretbox.open(encodedMessage, urlCryptoData.urlNonce, secretKey);
        if (messageBytes == null) {
            if (hasPassword) {
                $("#decryptPassword").addClass('is-invalid');
                // loadCypherAction = false;
                setLoadCypherAction(false);
                decodeButtonAccessibility(true);
                return;
            }
            $('#decodedMessage').html(props.var.generalError);
            showDecodedMessage(); // TODO: remove "Decoded message:" header
            return;
        }

        const message = naclutil.encodeUTF8(messageBytes);

        props.messageCallback(message);

        if (hasPassword) {
            const obj = {};
            obj.hash = urlCryptoData.hash;
            libs.post('DELETE', obj, libs.DELETE_URL, deleteSuccess, deleteError(obj));
        }

    }

    function loadError(XMLHttpRequest, textStatus, errorThrown) {
        if (XMLHttpRequest.status === 404) {
            $("#decodeButtonBlock").addClass('d-none');
            $("#decryptPasswordBlock").addClass('d-none');
            $("#errorForDecodedMessage").removeClass('d-none');
            decodeButtonAccessibility(true);
        } else {
            decodeButtonAccessibility(true);
            alert(props.var.decryptNetworkError);
        }
    }

    function deleteSuccess() { // do nothing
    }

    function deleteError(obj) {
        return function (XMLHttpRequest, textStatus, errorThrown) {
            // try to delete again
            window.setTimeout(function () {
                libs.post('DELETE', obj, '/delete?again', deleteSuccess, deleteErrorTryAgain);
            }, 1000);
        }
    }

    function deleteErrorTryAgain(XMLHttpRequest, textStatus, errorThrown) {  // do nothing
    }

    function decodeButtonAccessibility(state) {
        if (state) {
            setDecodeButton(true);
            setDecodeButtonSpinner(false);
        } else {
            setDecodeButton(false);
            if (!libs.IE()) {
                setDecodeButtonSpinner(true);
            }
        }
    }

    function showDecodedMessage() {
        decodeButtonAccessibility(true);
    }

    function updatePassword(event) {
        setMessagePassword(event.target.value);
        if (event.target.value.length === 0) {
            setMessagePasswordOk(false);
        } else {
            setMessagePasswordOk(true);
        }
    }


    return (
        <>
            <div className="container">
                <div className="row d-none" id="errorForDecodedMessage">
                    <div className="col-sm">
                        <p className="text-secondary">{props.var.messageRead}
                        </p>
                    </div>
                </div>
                <div className="row">
                    <div className="input-group mb-3" id="decryptPasswordBlock">
                        <div className="input-group">
                            <div className="input-group-prepend">
                                <span className="input-group-text">{props.var.password}</span>
                            </div>
                            <input type="text"
                                   className={messagePasswordOk ? "form-control" : "form-control is-invalid"}
                                   id="decryptPassword"
                                   placeholder={props.var.passwordDecryptPlaceholder}
                                   onChange={updatePassword}

                            />
                        </div>
                        <div className="col-sm text-danger text-center font-weight-light d-none"
                             id="ieDecryptWarning">{props.var.ieDecryptWarning}</div>
                    </div>
                </div>
                <div className="row">
                    <div className="col-sm mb-2" id="decodeButtonBlock">
                        <button type="button" id="decodeButton" onClick={decrypt}
                                className={decodeButton ? "btn btn-danger btn-block btn-lg" : "btn btn-danger btn-block btn-lg disabled"}>
                            <span id="decodeButtonSpinner"
                                  className={decodeButtonSpinner ? "spinner-border spinner-border-sm" : "spinner-border spinner-border-sm d-none"}/>
                            {props.var.readMessageButton}
                        </button>
                    </div>
                    <div className="col-sm">
                        <button type="button" className="btn btn-primary btn-block btn-lg"
                                onClick={props.againCallback}>{props.var.newMessageButton}
                        </button>
                    </div>
                </div>
            </div>
        </>
    )
}

export default Decrypt;