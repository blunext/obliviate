import React from "react";
import {libs} from "./commons";
import nacl from "tweetnacl";
import $ from "jquery";
import naclutil from "tweetnacl-util";

class Encrypt extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            message: '',
            messagePassword: '',
            messageOk: true,
            passwordOk: true,
            buttonEncode: true,
            encodeSpinner: false,
        };
        this.hasPassword = false;
        this.secretKey = ''; //TODO: zmienic nazwę
        this.salt = '';
        this.time = 0;
        this.warningForIE = libs.IE();
        this.keys = nacl.box.keyPair();
        this.urlNonce = '';

        console.log("Encrypt start");
    }

    onChangeMessage = (event) => {
        this.setState({message: event.target.value});
        if (event.target.value.length === 0) {
            this.setState({messageOk: false});
        } else {
            this.setState({messageOk: true});
        }
    }
    onChangePassword = (event) => {
        this.setState({messagePassword: event.target.value});
        if (event.target.value.length === 0) {
            this.setState({passwordOk: false});
        } else {
            this.setState({passwordOk: true});
        }
    }

    processEncrypt = (e) => {
        if ($("#passwordBlock").hasClass("collapsing")) {
            return;
        }
        if (this.state.message.length === 0) {
            this.setState({messageOk: false});
            return;
        }

        if ($("#passwordBlock").hasClass("show")) {
            if (this.state.messagePassword.length > 0) {
                this.encodeButtonAccessibility(false);
                this.hasPassword = true;

                this.salt = nacl.randomBytes(nacl.secretbox.keyLength);  // the same as key, 32 bytes
                libs.calculateKeyDerived(this.state.messagePassword, this.salt, libs.scryptLogN, this.scryptCallback);
            } else {
                this.setState({passwordOk: false});
            }
            return;
        } else {
            this.encodeButtonAccessibility(false);
        }
        this.secretKey = nacl.randomBytes(nacl.secretbox.keyLength);
        this.continue();
    }
    scryptCallback = (key, time) => {
        this.secretKey = key;
        this.time = time;
        this.continue();
    }
    continue = () => {
        // encrypt message with nacl secretbox
        const messageUTF8 = naclutil.decodeUTF8(this.state.message);
        const messageNonce = nacl.randomBytes(nacl.secretbox.nonceLength);

        const encryptedMessage = nacl.secretbox(messageUTF8, messageNonce, this.secretKey);

        // nonce will be used as a link anchor
        this.urlNonce = naclutil.encodeBase64(messageNonce);

        // store secret key in the message
        const fullMessage = new Uint8Array(this.secretKey.length + encryptedMessage.length);
        if (this.hasPassword) {
            fullMessage.set(this.salt);
        } else {
            fullMessage.set(this.secretKey);
        }
        fullMessage.set(encryptedMessage, this.secretKey.length);

        // encrypt message transmission with nacl box
        const transmissionNonce = nacl.randomBytes(nacl.box.nonceLength);
        const transmission = nacl.box(fullMessage, transmissionNonce, this.props.var.serverPublicKey, this.keys.secretKey);

        const obj = {};
        obj.message = naclutil.encodeBase64(transmission);
        obj.nonce = naclutil.encodeBase64(transmissionNonce);
        obj.hash = naclutil.encodeBase64(nacl.hash(messageNonce));
        obj.publicKey = naclutil.encodeBase64(this.keys.publicKey);
        if (this.hasPassword) {
            obj.time = this.time;
        }

        libs.post('POST', obj, libs.SAVE_URL, this.encodeSuccess, this.encodeError);
    }
    encodeButtonAccessibility = (state) => {
        if (state) {
            this.setState({buttonEncode: true})
            this.setState({encodeSpinner: false})
        } else {
            this.setState({buttonEncode: false})
            if (!libs.IE()) {
                this.setState({encodeSpinner: false})
            }
        }
    }
    encodeSuccess = (result) => {
        let index;
        if (this.hasPassword) {
            index = libs.queryIndexWithPassword;
        } else {
            index = 3;
        }
        if (!window.location.origin) { // IE fix
            window.location.origin = window.location.protocol + "//" + window.location.hostname +
                (window.location.port === 443 ? "" : ":" + window.location.port);
        }
        const url = window.location.origin + '/?' + this.urlNonce.substring(0, index) + "#" + this.urlNonce.substring(index, 32);
        this.props.linkCallback(url);
    }
    encodeError = (XMLHttpRequest, textStatus, errorThrown) => {
        this.encodeButtonAccessibility(true);
        alert(this.props.var.encryptNetworkError);
    }

    render() {
        return (
            <>
                <label htmlFor="message" className="text-secondary">{this.props.var.enterTextMessage}</label>
                <textarea className={this.state.messageOk ? "form-control mb-3" : "form-control mb-3 is-invalid"}
                          id="message"
                          rows="4" maxLength="262144"
                          autoFocus defaultValue={this.props.var.message}
                          onChange={this.onChangeMessage}/>
                <div className="container">
                    <div className="row">
                        <div className="input-group mb-3 collapse" id="passwordBlock">
                            <div className="input-group">
                                <div className="input-group-prepend">
                                    <span className="input-group-text">{this.props.var.password}</span>
                                </div>
                                <input type="text"
                                       className={this.state.passwordOk ? "form-control" : "form-control is-invalid"}
                                       id="encryptPassword"
                                       placeholder={this.props.var.passwordEncryptPlaceholder}
                                       onChange={this.onChangePassword}/>
                            </div>
                            <div
                                className={this.warningForIE ? "col-sm text-danger text-center font-weight-light" : "col-sm text-danger text-center font-weight-light d-none"}>
                                {this.props.var.ieEncryptWarning}</div>
                        </div>
                    </div>
                    <div className="row">
                        <div className="col-sm mb-2">
                            <button type="button" className="btn btn-success btn-block btn-lg"
                                    data-toggle="collapse"
                                    data-target="#passwordBlock">{this.props.var.password}
                            </button>
                        </div>
                        <div className="col-sm">
                            <button type="button" className=
                                {this.state.buttonEncode ? "btn btn-danger btn-block btn-lg" : "btn btn-danger btn-block btn-lg disabled"}
                                    id="encodeButton"
                                    value={this.state.messagePassword}
                                    onClick={this.processEncrypt}>
                                <span
                                    className={this.state.encodeSpinner ? "spinner-border spinner-border-sm" : "spinner-border spinner-border-sm d-none"}
                                    id="encodeButtonSpinner"/>
                                {this.props.var.secureButton}
                            </button>
                        </div>
                    </div>
                </div>
            </>
        );
    }
}

export default Encrypt;