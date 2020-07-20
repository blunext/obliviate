import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap/dist/js/bootstrap.bundle.min';
import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import './obliviate.js';
import * as serviceWorker from './serviceWorker';
import axios from 'axios';
import {config} from './constants'

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
            message: '', messagePassword: '',
        };
    }

    onChangeMessage = (event) => {
        this.setState({message: event.target.value});
    }

    onChangePassword = (event) => {
        this.setState({messagePassword: event.target.value});
    }

    processEncrypt = (event) => {
        alert("klik, " + this.state.message + " " + this.state.messagePassword)
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
                {/*<div className="form-group d-none mt-3 mb-3" id="linkBlock">*/}
                {/*    <label htmlFor="link" className="text-secondary">{copyLinkButtoncopyLink}</label>*/}
                {/*    <textarea className="form-control mb-3" id="link" rows="1"></textarea>*/}
                {/*    <div className="container">*/}
                {/*        <div className="row">*/}
                {/*            <div className="col-sm mb-2">*/}
                {/*                <button type="button" className="btn btn-warning btn-block btn-lg"*/}
                {/*                        data-clipboard-action="copy"*/}
                {/*                        data-clipboard-target="#link">{copyLinkButton}*/}
                {/*                </button>*/}
                {/*            </div>*/}
                {/*            <div className="col-sm">*/}
                {/*                <button type="button" className="btn btn-primary btn-block btn-lg"*/}
                {/*                        onClick="again();">{newMessageButton}*/}
                {/*                </button>*/}
                {/*            </div>*/}
                {/*        </div>*/}
                {/*    </div>*/}
                {/*</div>*/}
                {/*<div className="form-group d-none mt-3 mb-3" id="presentationBlock">*/}
                {/*    <div className="container">*/}
                {/*        <div className="row">*/}
                {/*            <div className="col-sm text-secondary">*/}
                {/*                <p className="text-center">{decodedMessage}:</p>*/}
                {/*                <div className="border-top my-3"></div>*/}
                {/*                <p id="decodedMessage"></p>*/}
                {/*            </div>*/}
                {/*        </div>*/}
                {/*        <div className="row">*/}
                {/*            <div className="col-sm">*/}
                {/*                <button type="button" className="btn btn-primary btn-block btn-lg"*/}
                {/*                        onClick="again();">{newMessageButton}*/}
                {/*                </button>*/}
                {/*            </div>*/}
                {/*        </div>*/}
                {/*    </div>*/}
                {/*</div>*/}
                {/*<div className="form-group d-none mt-3 mb-3" id="decodeBlock">*/}
                {/*    <div className="container">*/}
                {/*        <div className="row d-none" id="errorForDecodedMessage">*/}
                {/*            <div className="col-sm">*/}
                {/*                <p className="text-secondary">{messageRead}*/}
                {/*                </p>*/}
                {/*            </div>*/}
                {/*        </div>*/}
                {/*        <div className="row">*/}
                {/*            <div className="input-group mb-3" id="decryptPasswordBlock">*/}
                {/*                <div className="input-group">*/}
                {/*                    <div className="input-group-prepend">*/}
                {/*                        <span className="input-group-text">{password}</span>*/}
                {/*                    </div>*/}
                {/*                    <input type="text" className="form-control" id="decryptPassword"*/}
                {/*                           placeholder="{passwordDecryptPlaceholder}"/>*/}
                {/*                </div>*/}
                {/*                <div className="col-sm text-danger text-center font-weight-light d-none"*/}
                {/*                     id="ieDecryptWarning">{ieDecryptWarning}</div>*/}
                {/*            </div>*/}
                {/*        </div>*/}
                {/*        <div className="row">*/}
                {/*            <div className="col-sm mb-2" id="decodeButtonBlock">*/}
                {/*                <button type="button" className="btn btn-danger btn-block btn-lg"*/}
                {/*                        id="decodeButton">*/}
                {/*                    <span className="spinner-border spinner-border-sm d-none"*/}
                {/*                          id="decodeButtonSpinner"></span>*/}
                {/*                    {readMessageButton}*/}
                {/*                </button>*/}
                {/*            </div>*/}
                {/*            <div className="col-sm">*/}
                {/*                <button type="button" className="btn btn-primary btn-block btn-lg"*/}
                {/*                        onClick="again();">{newMessageButton}*/}
                {/*                </button>*/}
                {/*            </div>*/}
                {/*        </div>*/}
                {/*    </div>*/}
                {/*</div>*/}
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
        axios.get(config.API_URL)
            .then(res => {
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