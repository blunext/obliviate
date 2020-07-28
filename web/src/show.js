import React from 'react';

function Show(props) {
    console.log("Show start");

    return (
        <>
            <div className="container">
                <div className="row">
                    <div className="col-sm text-secondary">
                        <p className="text-center">{props.var.decodedMessage}:</p>
                        <div className="border-top my-3"/>
                        <p className="br-line">{props.message}</p>
                    </div>
                </div>
                <div className="row">
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

export default Show;