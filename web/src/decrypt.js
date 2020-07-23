import React from "react";


function Decrypt(props) {

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
                            <input type="text" className="form-control" id="decryptPassword"
                                   placeholder={props.var.passwordDecryptPlaceholder}/>
                        </div>
                        <div className="col-sm text-danger text-center font-weight-light d-none"
                             id="ieDecryptWarning">{props.var.ieDecryptWarning}</div>
                    </div>
                </div>
                <div className="row">
                    <div className="col-sm mb-2" id="decodeButtonBlock">
                        <button type="button" className="btn btn-danger btn-block btn-lg"
                                id="decodeButton">
                            <span className="spinner-border spinner-border-sm d-none" id="decodeButtonSpinner"/>
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