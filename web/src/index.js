import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap/dist/js/bootstrap.bundle.min';
import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import './obliviate.js';
import * as serviceWorker from './serviceWorker';
import axios from 'axios';
import {libs} from './commons'
import $ from "jquery";
import nacl from "tweetnacl";
import naclutil from "tweetnacl-util";
import ClipboardJS from "clipboard";

new ClipboardJS('.btn');
let serverPublicKey = '';
let keys = nacl.box.keyPair();

let urlNonce = '';
const queryIndexWithPassword = 4;

const isMobile = window.matchMedia("only screen and (max-width: 760px)").matches;
if (isMobile) {
    $("#link").attr('rows', 2);
}

// if (window.location.hash) {
//     decrypt.password = window.location.search.substring(1).length === queryIndexWithPassword;
//     showDecodeButton();
// } else {
//     showEnterMessage();
// }

// // necessary for .off('click')
// $("#decodeButton").click(function (e) {
//     decrypt.loadCypher();
// });

if (libs.IE()) {
    $("#ieEncryptWarning").removeClass('d-none');
    $("#ieDecryptWarning").removeClass('d-none');
}

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
        this.secretKey = '';
        this.salt = '';
        this.time = 0;
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
        urlNonce = naclutil.encodeBase64(messageNonce);

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
        const transmission = nacl.box(fullMessage, transmissionNonce, serverPublicKey, keys.secretKey);

        const obj = {};
        obj.message = naclutil.encodeBase64(transmission);
        obj.nonce = naclutil.encodeBase64(transmissionNonce);
        obj.hash = naclutil.encodeBase64(nacl.hash(messageNonce));
        obj.publicKey = naclutil.encodeBase64(keys.publicKey);
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
        this.showLink();
    }
    encodeError = (XMLHttpRequest, textStatus, errorThrown) => {
        this.encodeButtonAccessibility(true);
        alert(this.props.var.encryptNetworkError);
    }

    showLink = () => {
        $("#inputMessageBlock").addClass('d-none');
        $("#linkBlock").removeClass('d-none');
        $("#decodeBlock").addClass('d-none');
        $("#presentationBlock").addClass('d-none');

        $("#message").val("");
        this.encodeButtonAccessibility(true);
    }

    render() {
        return (
            <div className="container border border-primary">
                <div className="form-group mt-3 mb-3" id="inputMessageBlock">
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
                                <div className="col-sm text-danger text-center font-weight-light d-none"
                                     id="ieEncryptWarning">{this.props.var.ieEncryptWarning}</div>
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
                </div>
            </div>
        );
    }
}

class Main extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            header: '',
            enterTextMessage: '',
            password: '',
            passwordEncryptPlaceholder: '',
            ieEncryptWarning: '',
            secureButton: '',
            infoHeader: '',
            info: '',
            info1: '',
            info2: '',
            info3: '',
            encryptNetworkError: '',
            ready: false,
        };
    }

    componentDidMount() {
        axios.get(libs.VARIABLES_URL)
            .then(res => {
                serverPublicKey = naclutil.decodeBase64(res.data.PublicKey);
                this.setState({
                    header: res.data.header,
                    enterTextMessage: res.data.enterTextMessage,
                    password: res.data.password,
                    passwordEncryptPlaceholder: res.data.passwordEncryptPlaceholder,
                    ieEncryptWarning: res.data.ieEncryptWarning,
                    secureButton: res.data.secureButton,
                    infoHeader: res.data.infoHeader,
                    info: res.data.info,
                    info1: res.data.info1,
                    info2: res.data.info2,
                    info3: res.data.info3,
                    encryptNetworkError: res.data.encryptNetworkError,
                    ready: true,
                });
            });
    }

    render() {
        if (!this.state.ready) {
            return (
                <div className="loader">Loading...</div>
            )
        } else {
            return (
                <div>
                    <h4 className="text-secondary text-center mt-2">{this.state.header}</h4>
                    <Encrypt var={this.state}/>
                    <div className="container mt-3">
                        <div className="row">
                            <div className="col-sm-2">
                            </div>
                            <div className="col">
                                <hr/>
                            </div>
                            <div className="col-auto text-secondary"><small>{this.state.infoHeader}</small></div>
                            <div className="col">
                                <hr/>
                            </div>
                            <div className="col-sm-2">
                            </div>
                        </div>
                        <div className="row">
                            <div className="col-sm-2">
                            </div>
                            <div className="col-sm-8">
                                <p className="text-secondary">
                                    <small>
                                        {this.state.info} <a href="https://github.com/blunext/obliviate" target="_blank"
                                                             rel="noopener noreferrer">GitHub</a>.
                                        {this.state.info1} <a href="mailto:info@securenote.io" target="_blank"
                                                              rel="noopener noreferrer">{this.state.info2}</a>. {this.state.info3}
                                    </small>
                                </p>
                            </div>
                            <div className="col-sm-2">
                            </div>
                        </div>
                    </div>
                </div>
            )
        }
    }
}

ReactDOM.render(<Main/>, document.getElementById('root'));

serviceWorker.unregister();