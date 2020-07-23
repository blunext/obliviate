import scryptAsynch from "scrypt-async";
import nacl from "tweetnacl";
import $ from "jquery";


// TODO: get rid of potpone


const commons = {
    VARIABLES_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/variables' : '/variables',
    SAVE_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/save' : '/save',
    READ_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/read' : '/read',
    DELETE_URL: process.env.NODE_ENV === 'development' ? 'http://localhost:3000/delete' : '/delete',
    scryptLogN: 14,
    queryIndexWithPassword: 4,
    calculateKeyDerived: function (password, salt, logN, callback) {
        // let postpone = this.IE() ? 0 : 750;
        let postpone = 0;
        window.setTimeout(function () {
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
    },
    arraySlice: function (arr, x, y) {
        if (this.IE()) {
            return arr.slice(x, y);
        }
        // IE stuff
        let ret = [];
        for (let i = 0; i < arr.length; i++) {
            if (i >= x && i < y) {
                ret.push(arr[i]);
            }
        }
        return new Uint8Array(ret);
    },
    replaceAll: function (str, find, replace) {
        return str.replace(new RegExp(find, 'g'), replace);
    }

};


var getTime = (function () {
    if (typeof performance !== "undefined") {
        return performance.now.bind(performance);
    }
    return Date.now.bind(Date);
})();

export const libs = commons;
