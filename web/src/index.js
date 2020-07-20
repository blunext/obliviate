import 'bootstrap/dist/css/bootstrap.min.css';
// import $ from 'jquery';
// import Popper from 'popper.js';
import 'bootstrap/dist/js/bootstrap.bundle.min';

import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import './obliviate.js';

const header = 'aaa';
const enterTextMessage = 'aaa';
const password = 'aaa';
const passwordEncryptPlaceholder = 'aaa';
const ieEncryptWarning = 'aaa';
const secureButton = 'aaa';
// const copyLink = 'aaa';
// const copyLinkButtoncopyLink = 'aaa';
// const copyLinkButton = 'aaa';
// const newMessageButton = 'aaa';
// const decodedMessage = 'aaa';
// const messageRead = 'aaa';
// const passwordDecryptPlaceholder = 'aaa';
// const ieDecryptWarning = 'aaa';
// const readMessageButton = 'aaa';
const infoHeader = 'aaa';
const info = 'aaa';
const info1 = 'aaa';
const info2 = 'aaa';
const info3 = 'aaa';

class Encrypt extends React.Component {
    constructor(props) {
        super(props);
        this.state = {message: '', password: ''};
        this.onChangeMessage = this.onChangeMessage.bind(this);
        this.onChangePassword = this.onChangePassword.bind(this);
        this.processEncrypt = this.processEncrypt.bind(this);
    }

    onChangeMessage(event) {
        this.setState({message: event.target.value});
    }

    onChangePassword(event) {
        this.setState({password: event.target.value});
    }

    processEncrypt(event) {
        alert("klik, " + this.state.message + " " + this.state.password)
    }


    render() {
        return (
            <div>
                <h4 className="text-secondary text-center mt-2">{header}</h4>
                <div className="container border border-primary">
                    <div className="form-group

                    {/*d-none */}

                    mt-3 mb-3" id="inputMessageBlock">
                        <label htmlFor="message" className="text-secondary">{enterTextMessage}</label>
                        <textarea className="form-control mb-3" id="message" rows="4" maxLength="262144"
                                  autoFocus defaultValue={this.state.message}
                                  onChange={this.onChangeMessage}></textarea>
                        <div className="container">
                            <div className="row">
                                <div className="input-group mb-3 collapse" id="passwordBlock">
                                    <div className="input-group">
                                        <div className="input-group-prepend">
                                            <span className="input-group-text">{password}</span>
                                        </div>
                                        <input type="text" className="form-control" id="encryptPassword"
                                               placeholder={passwordEncryptPlaceholder}
                                               onChange={this.onChangePassword}/>
                                    </div>
                                    <div className="col-sm text-danger text-center font-weight-light d-none"
                                         id="ieEncryptWarning">{ieEncryptWarning}</div>
                                </div>
                            </div>
                            <div className="row">
                                <div className="col-sm mb-2">
                                    <button type="button" className="btn btn-success btn-block btn-lg"
                                            data-toggle="collapse"
                                            data-target="#passwordBlock">{password}
                                    </button>
                                </div>
                                <div className="col-sm">
                                    <button type="button" className="btn btn-danger btn-block btn-lg" id="encodeButton"
                                            value={this.state.password}
                                            onClick={this.processEncrypt}>
                                <span className="spinner-border spinner-border-sm d-none"
                                      id="encodeButtonSpinner"></span>
                                        {secureButton}
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

                <div className="container mt-3">
                    <div className="row">
                        <div className="col-sm-2">
                        </div>
                        <div className="col">
                            <hr/>
                        </div>
                        <div className="col-auto text-secondary"><small>{infoHeader}</small></div>
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
                                    {info} <a href="https://github.com/blunext/obliviate" target="_blank"
                                              rel="noopener noreferrer">GitHub</a>.
                                    {info1} <a href="mailto:info@securenote.io" target="_blank"
                                               rel="noopener noreferrer">{info2}</a>. {info3}
                                </small>
                            </p>
                        </div>
                        <div className="col-sm-2">
                        </div>
                    </div>
                </div>
            </div>
        );
    }


}

ReactDOM.render(<Encrypt/>, document.getElementById('root'));

