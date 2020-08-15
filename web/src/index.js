import 'react-app-polyfill/ie11';
import 'react-app-polyfill/stable';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap/dist/js/bootstrap.bundle.min';
import React, {useEffect, useRef, useState} from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import * as serviceWorker from './serviceWorker';
import axios from 'axios';
import {commons} from './commons'
import naclutil from "tweetnacl-util";
import Encrypt from "./encrypt";
import ShowLink from "./showlink";
import Decrypt from "./decrypt";
import Show from "./show";

function Main() {
    const vars = useRef({});

    const parts = {ENCRYPT: 0, LINK: 1, DECRYPT: 2, SHOW: 3};
    const [ready, setReady] = useState(false);
    const [link, setLink] = useState('');
    const [visible, setVisible] = useState(parts.ENCRYPT);
    const [message, setMessage] = useState('');

    console.log("Main start");

    useEffect(() => {

        if (window.location.hash) {
            setVisible(parts.DECRYPT);
        }

        axios.get(commons.VARIABLES_URL)
            .then(res => {
                vars.current = res.data;
                vars.current.serverPublicKey = naclutil.decodeBase64(res.data.PublicKey);

                document.title = vars.current.title;
                document.description = vars.current.description;

                setReady(true);
            })
            .catch(err => {
                // TODO: -----
                console.log(err);
            });


    }, [])

    function linkCallback(url) {
        setLink(url);
        setVisible(parts.LINK);
    }

    function againCallback() {
        setLink('');
        setVisible(parts.ENCRYPT);
    }

    function messageCallback(message) {
        setMessage(message);
        setVisible(parts.SHOW);
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
                    <div className="form-group mt-3 mb-3">
                        {visible === parts.ENCRYPT ?
                            <Encrypt var={vars.current} linkCallback={linkCallback}/> : null}
                        {visible === parts.LINK ?
                            <ShowLink var={vars.current} link={link} againCallback={againCallback}/> : null}
                        {visible === parts.DECRYPT ?
                            <Decrypt var={vars.current} messageCallback={messageCallback}
                                     againCallback={againCallback}/> : null}
                        {visible === parts.SHOW ?
                            <Show var={vars.current} message={message} againCallback={againCallback}/> : null}
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