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
            header: this.props.var.header,
            enterTextMessage: this.props.var.enterTextMessage,
            password: this.props.var.password,
            passwordEncryptPlaceholder: this.props.var.passwordEncryptPlaceholder,
            ieEncryptWarning: this.props.var.ieEncryptWarning,
            secureButton: this.props.var.secureButton,
            infoHeader: this.props.var.infoHeader,
            info: this.props.var.info,
            info1: this.props.var.info1,
            info2: this.props.var.info2,
            info3: this.props.var.info3,
            //---
            message: '', messagePassword: '',
            hasPassword: false,
            secretKey: '',
            salt: '',
            time: 0,
        };
    }

    onChangeMessage = (event) => {
        this.setState({message: event.target.value});
    }
    onChangePassword = (event) => {
        this.setState({messagePassword: event.target.value});
    }

    processEncrypt = (e) => {
        debugger;
        if ($("#passwordBlock").hasClass("collapsing")) {
            return;
        }
        // encrypt.message = $('#message').val();
        if (this.state.message.length === 0) {
            $("#message").addClass('is-invalid');
            return;
        }
        $("#message").removeClass('is-invalid');

        if ($("#passwordBlock").hasClass("show")) {
            // const password = $('#encryptPassword').val();
            if (this.state.password.length > 0) {
                this.encodeButtonAccessibility(false);
                this.scope.hasPassword = true;

                this.setState({salt: nacl.randomBytes(nacl.secretbox.keyLength)});  // the same as key, 32 bytes
                libs.calculateKeyDerived(this.state.password, this.state.salt, libs.scryptLogN, this.scryptCallback);
                $('#encryptPassword').removeClass('is-invalid');
            } else {
                $('#encryptPassword').addClass('is-invalid');
            }
            return;
        } else {
            this.encodeButtonAccessibility(false);
        }
        this.setState({secretKey: nacl.randomBytes(nacl.secretbox.keyLength)})
        this.continue();
    }
    scryptCallback = (key, time) => {
        this.setState({secretKey: key});
        this.setState({time: time});
        this.continue();
    }
    continue = () => {
        debugger;
        // encrypt message with nacl secretbox
        const messageUTF8 = nacl.util.decodeUTF8(this.state.message);
        const messageNonce = nacl.randomBytes(nacl.secretbox.nonceLength);

        const encryptedMessage = nacl.secretbox(messageUTF8, messageNonce, this.state.secretKey);

        // nonce will be used as a link anchor
        urlNonce = nacl.util.encodeBase64(messageNonce);

        // store secret key in the message
        const fullMessage = new Uint8Array(this.state.secretKey.length + encryptedMessage.length);
        if (this.state.hasPassword) {
            fullMessage.set(this.state.salt);
        } else {
            fullMessage.set(this.state.secretKey);
        }
        fullMessage.set(encryptedMessage, this.state.secretKey.length);

        // encrypt message transmission with nacl box
        const transmissionNonce = nacl.randomBytes(nacl.box.nonceLength);
        const transmission = nacl.box(fullMessage, transmissionNonce, serverPublicKey, keys.secretKey);

        const obj = {};
        obj.message = nacl.util.encodeBase64(transmission);
        obj.nonce = nacl.util.encodeBase64(transmissionNonce);
        obj.hash = nacl.util.encodeBase64(nacl.hash(messageNonce));
        obj.publicKey = nacl.util.encodeBase64(keys.publicKey);
        if (this.state.hasPassword) {
            obj.time = this.state.time;
        }

        libs.post('POST', obj, '/save', this.encodeSuccess, this.encodeError);
    }
    encodeButtonAccessibility = (state) => {
        if (state) {
            $("#encodeButton").removeClass('disabled');
            $("#encodeButtonSpinner").addClass('d-none');
        } else {
            $("#encodeButton").addClass('disabled');
            if (!libs.IE()) {
                $("#encodeButtonSpinner").removeClass('d-none');
            }
        }
    }
    encodeSuccess = (result) => {
        let index;
        if (this.state.hasPassword) {
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
        alert('{{.encryptNetworkError}}');
    }
    function = () => {
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
                <div className="form-group

                    {/*d-none */}

                    mt-3 mb-3" id="inputMessageBlock">
                    <label htmlFor="message" className="text-secondary">{this.state.enterTextMessage}</label>
                    <textarea className="form-control mb-3" id="message" rows="4" maxLength="262144"
                              autoFocus defaultValue={this.state.message}
                              onChange={this.onChangeMessage}/>
                    <div className="container">
                        <div className="row">
                            <div className="input-group mb-3 collapse" id="passwordBlock">
                                <div className="input-group">
                                    <div className="input-group-prepend">
                                        <span className="input-group-text">{this.state.password}</span>
                                    </div>
                                    <input type="text" className="form-control" id="encryptPassword"
                                           placeholder={this.state.passwordEncryptPlaceholder}
                                           onChange={this.onChangePassword}/>
                                </div>
                                <div className="col-sm text-danger text-center font-weight-light d-none"
                                     id="ieEncryptWarning">{this.state.ieEncryptWarning}</div>
                            </div>
                        </div>
                        <div className="row">
                            <div className="col-sm mb-2">
                                <button type="button" className="btn btn-success btn-block btn-lg"
                                        data-toggle="collapse"
                                        data-target="#passwordBlock">{this.state.password}
                                </button>
                            </div>
                            <div className="col-sm">
                                <button type="button" className="btn btn-danger btn-block btn-lg" id="encodeButton"
                                        value={this.state.messagePassword}
                                        onClick={this.processEncrypt}>
                                <span className="spinner-border spinner-border-sm d-none"
                                      id="encodeButtonSpinner"/>
                                    {this.state.secureButton}
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
            ready: false,
        };
    }

    componentDidMount() {


        axios.get(libs.API_URL)
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