import scryptAsynch from "scrypt-async";
import nacl from "tweetnacl";
import $ from "jquery";


// TODO: get rid of potpone


const commons = {
    VARIABLES_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/variables' : '/variables',
    SAVE_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/save' : '/save',
    READ_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/read' : '/read',
    DELETE_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/delete' : '/delete',
    costFactorDefault: 14,
    costFactor: 15, // CPU/memory cost parameter
    queryIndexWithPassword: 4,
    calculateKeyDerived: function (password, salt, logN, callback) {
        try {
            const t1 = getTime();
            scryptAsynch(password, salt, {
                    logN: logN,
                    r: 8,
                    p: 1,
                    dkLen: nacl.secretbox.keyLength, // 32
                    interruptStep: 0,
                    encoding: 'binary' // hex, base64, binary
                },
                function (res) {
                    const time = Math.round(getTime() - t1);
                    callback(res, time);
                }
            );
        } catch (ex) {
            alert(ex.message); //TODO: process the exception
        }
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
    },
};


var getTime = (function () {
    if (typeof performance !== "undefined") {
        return performance.now.bind(performance);
    }
    return Date.now.bind(Date);
})();

export const libs = commons;
