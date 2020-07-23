import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap/dist/js/bootstrap.bundle.min';
import React, {Suspense, useEffect, useRef, useState} from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import * as serviceWorker from './serviceWorker';
import axios from 'axios';
import {libs} from './commons'
import naclutil from "tweetnacl-util";

const Encrypt = React.lazy(() => import('./encrypt'));
// import ShowLink from "./showlink";
const ShowLink = React.lazy(() => import('./showlink'));

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

function Main() {
    const [ready, setReady] = useState(false);
    const [link, setLink] = useState('');
    const vars = useRef({});

    useEffect(() => {
        axios.get(libs.VARIABLES_URL)
            .then(res => {
                vars.current.serverPublicKey = naclutil.decodeBase64(res.data.PublicKey);
                vars.current.header = res.data.header;
                vars.current.enterTextMessage = res.data.enterTextMessage;
                vars.current.password = res.data.password;
                vars.current.passwordEncryptPlaceholder = res.data.passwordEncryptPlaceholder;
                vars.current.ieEncryptWarning = res.data.ieEncryptWarning;
                vars.current.secureButton = res.data.secureButton;
                vars.current.infoHeader = res.data.infoHeader;
                vars.current.info = res.data.info;
                vars.current.info1 = res.data.info1;
                vars.current.info2 = res.data.info2;
                vars.current.info3 = res.data.info3;
                vars.current.encryptNetworkError = res.data.encryptNetworkError;
                vars.current.copyLink = res.data.copyLink;
                vars.current.copyLinkButton = res.data.copyLinkButton;
                vars.current.newMessageButton = res.data.newMessageButton;
                setReady(true);
            })
            .catch(err => {
                // TODO: -----
                console.log(err);
            });
    }, [])

    function linkCallback(url) {
        setLink(url);
    }

    function againCallback() {
        setLink('');
    }

    if (!ready) {
        return (
            <div className="loader">Loading...</div>
        )
    } else {
        return (
            <>
                <h4 className="text-secondary text-center mt-2">{vars.current.header}</h4>
                <div className="container border border-primary">
                    <div className="form-group mt-3 mb-3" id="inputMessageBlock">
                        <Suspense fallback={<div className="loader">Loading...</div>}>
                            {link === '' ? <Encrypt var={vars.current} linkCallback={linkCallback}/> : null}
                        </Suspense>
                        <Suspense fallback={<div className="loader">Loading...</div>}>
                            {link !== '' ?
                                <ShowLink var={vars.current} link={link} againCallback={againCallback}/> : null}
                        </Suspense>
                    </div>
                </div>
                <div className="container mt-3">
                    <div className="row">
                        <div className="col-sm-2">
                        </div>
                        <div className="col">
                            <hr/>
                        </div>
                        <div className="col-auto text-secondary"><small>{vars.current.infoHeader}</small></div>
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
                                    {vars.current.info} <a href="https://github.com/blunext/obliviate"
                                                           target="_blank"
                                                           rel="noopener noreferrer">GitHub</a>.
                                    {vars.current.info1} <a href="mailto:info@securenote.io" target="_blank"
                                                            rel="noopener noreferrer">{vars.current.info2}</a>. {vars.current.info3}
                                </small>
                            </p>
                        </div>
                        <div className="col-sm-2">
                        </div>
                    </div>
                </div>
            </>
        )
    }
}

ReactDOM.render(<Main/>, document.getElementById('root'));

serviceWorker.unregister();