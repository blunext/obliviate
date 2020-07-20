// Constants.js
import scryptAsynch from "scrypt-async";
import nacl from "tweetnacl";
import $ from "jquery";

const commons = {
    API_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/variables' : '/variables',
    scryptLogN: 14,
    getTime: (function () {
        if (typeof performance !== "undefined") {
            return performance.now.bind(performance);
        }
        return Date.now.bind(Date);
    })(),
    calculateKeyDerived: function (password, salt, logN, callback) {
        let postpone = this.IE() ? 0 : 750;
        window.setTimeout(function () {
            try {
                const t1 = this.getTime();
                scryptAsynch(password, salt, {
                        logN: logN,
                        r: 8,
                        p: 1,
                        dkLen: nacl.secretbox.keyLength, // 32
                        interruptStep: 0,
                        encoding: 'binary' // hex, base64, binary
                    },
                    function (res) {
                        const time = Math.round(this.getTime() - t1);
                        callback(res, time);
                    }
                );
            } catch (ex) {
                alert(ex.message); //TODO: process the exception
            }
        }, postpone); // it freezes the UI so postpone invocation
    },
    IE: function () {
        return window.document.documentMode;
    },
    post: function (method, webObject, url, postSuccess, postError) {
        $.ajax({
            url: url,
            type: method,
            dataType: "json",
            data: JSON.stringify(webObject),
            success: postSuccess,
            error: postError
        });
    }

};
export const libs = commons;
