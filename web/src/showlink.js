import React from "react";
import ClipboardJS from 'clipboard';


function ShowLink(props) {
    const isMobile = window.matchMedia("only screen and (max-width: 760px)").matches;
    new ClipboardJS('.btn');
    console.log("ShowLink start");

    return (
        <>
            <label htmlFor="link" className="text-secondary">{props.var.copyLink}</label>
            <textarea className="form-control mb-3" id="link" rows={isMobile ? 2 : 1} value={props.link} readOnly/>
            <div className="container">
                <div className="row">
                    <div className="col-sm mb-2">
                        <button type="button" className="btn btn-warning btn-block btn-lg"
                                data-clipboard-action="copy"
                                data-clipboard-target="#link">{props.var.copyLinkButton}
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

export default ShowLink;