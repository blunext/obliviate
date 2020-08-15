import React, {useEffect, useState} from 'react';
import nacl from "tweetnacl";
import naclutil from "tweetnacl-util";
import {calculateKeyDerived, commons, post} from "./commons";
import {isIE} from "react-device-detect";

function Decrypt(props) {

    console.log("Decrypt start");

    const hasPassword = window.location.search.substring(1).length === commons.queryIndexWithPassword;

    const [decodeButton, setDecodeButton] = useState(true);
    const [decodeButtonSpinner, setDecodeButtonSpinner] = useState(false);
    const [loadCypherAction, setLoadCypherAction] = useState(true);
    const [urlCryptoData, setUrlCryptoData] = useState({urlNonce: '', hash: []});
    const [salt, setSalt] = useState('');
    const [secretKey, setSecretKey] = useState('');
    const [cypherLoaded, setCypherLoaded] = useState(false);
    const [cypherReady, setCypherReady] = useState(0);
    const [encodedMessage, setEncodedMessage] = useState(false);
    const [messagePassword, setMessagePassword] = useState('');
    const [messagePasswordOk, setMessagePasswordOk] = useState(true);
    const [messageReadInfo, setMessageReadInfo] = useState(false);
    const [costFactor, setCostFactor] = useState(commons.costFactorDefault);

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

        post('POST', obj, commons.READ_URL, decryptTransmission, loadError);

        function decryptTransmission(result) {
            // decode transmission with box
            const messageWithNonceAsUint8Array = naclutil.decodeBase64(result.message);
            const noncePart = messageWithNonceAsUint8Array.slice(0, nacl.box.nonceLength);
            const messagePart = messageWithNonceAsUint8Array.slice(nacl.box.nonceLength, result.message.length);

            const decrypted = nacl.box.open(messagePart, noncePart, props.var.serverPublicKey, keys.secretKey);
            if (!decrypted) {
                decodeButtonAccessibility(true);
                alert(props.var.generalError);
                return
            }
            // decode message with secretbox
            if (hasPassword) {
                setSalt(decrypted.slice(0, nacl.secretbox.keyLength));
                if (result.costFactor !== undefined) { // for backward compatibility
                    setCostFactor(result.costFactor);
                }
            } else {
                setSecretKey(decrypted.slice(0, nacl.secretbox.keyLength));
            }
            setEncodedMessage(decrypted.slice(nacl.secretbox.keyLength, decrypted.length));

            setCypherLoaded(true);
        }
    }

    useEffect(() => {
        if (cypherLoaded) {
            decryptMessage();
        }
    }, [cypherLoaded])

    useEffect(() => {
        if (cypherReady > 0) {
            decryptCypher();
        }
    }, [cypherReady])

    function decryptMessage() {
        decodeButtonAccessibility(false);
        if (hasPassword) {
            if (messagePassword.length > 0) {
                calculateKeyDerived(messagePassword, salt, costFactor, scryptCallback);
            } else {
                setMessagePasswordOk(false);
                decodeButtonAccessibility(true);
                setLoadCypherAction(false);
            }
            return;
        }
        setCypherReady(cypherReady + 1);

        function scryptCallback(key, time) { // do nothing with time while decrypt
            setSecretKey(key);
            setCypherReady(cypherReady + 1);
        }
    }

    function decryptCypher() {
        const messageBytes = nacl.secretbox.open(encodedMessage, urlCryptoData.urlNonce, secretKey);
        if (messageBytes == null) {
            if (hasPassword) {
                setLoadCypherAction(false);
                setMessagePasswordOk(false);
                decodeButtonAccessibility(true);
                return;
            }
            decodeButtonAccessibility(true);
            alert(props.var.generalError);
            return;
        }

        const message = naclutil.encodeUTF8(messageBytes);

        props.messageCallback(message, messagePassword);

        if (hasPassword) {
            const obj = {};
            obj.hash = urlCryptoData.hash;
            post('DELETE', obj, commons.DELETE_URL, doNothing, deleteError(obj));
        }

    }

    function loadError(err) {
        decodeButtonAccessibility(true);
        if (err.response !== undefined && err.response.status === 404) {
            setMessageReadInfo(true);
            setMessagePasswordOk(false); //hide pass
        } else {
            alert(props.var.decryptNetworkError);
        }
    }

    function doNothing() { // do nothing
    }

    function deleteError(obj) {
        return function (XMLHttpRequest, textStatus, errorThrown) {
            // try to delete again
            window.setTimeout(function () {
                post('DELETE', obj, '/delete?again', doNothing, doNothing);
            }, 1000);
        }
    }

    function decodeButtonAccessibility(state) {
        if (state) {
            setDecodeButton(true);
            setDecodeButtonSpinner(false);
        } else {
            setDecodeButton(false);
            if (!isIE) {
                setDecodeButtonSpinner(true);
            }
        }
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
                <div className={messageReadInfo ? "row" : "row d-none"}>
                    <div className="col-sm">
                        <p className="text-secondary">{props.var.messageRead}
                        </p>
                    </div>
                </div>
                <div className="row">
                    <div
                        className={messageReadInfo ? "d-none" : (hasPassword ? "input-group mb-3" : "input-group mb-3 d-none")}>
                        <div className="input-group">
                            <div className="input-group-prepend">
                                <span className="input-group-text">{props.var.password}</span>
                            </div>
                            <input type="text"
                                   className={messagePasswordOk ? "form-control" : "form-control is-invalid"}
                                   placeholder={props.var.passwordDecryptPlaceholder}
                                   onChange={updatePassword}

                            />
                        </div>
                        <div
                            className="col-sm text-danger text-center font-weight-light d-none">{props.var.ieDecryptWarning}</div>
                    </div>
                </div>
                <div className="row">
                    <div className={messageReadInfo ? "col-sm mb-2 d-none" : "col-sm mb-2"}>
                        <button type="button" onClick={decrypt}
                                className={decodeButton ? "btn btn-danger btn-block btn-lg" : "btn btn-danger btn-block btn-lg disabled"}>
                            <span
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