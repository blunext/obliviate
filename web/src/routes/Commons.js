import scryptAsynch from "scrypt-async"
import nacl from "tweetnacl"

const MODE = import.meta.env.MODE

export const CONSTANTS = {
    VARIABLES_URL: MODE === 'development' ? 'http://localhost:3000/variables' : '/variables',
    SAVE_URL: MODE === 'development' ? 'http://localhost:3000/save' : '/save',
    READ_URL: MODE === 'development' ? 'http://localhost:3000/read' : '/read',
    DELETE_URL: MODE === 'development' ? 'http://localhost:3000/delete' : '/delete',
    costFactor: 15,
    queryIndexWithPassword: 4
}

export const calculateKeyDerived = function (password, salt, logN, callback) {
    try {
        const t1 = getTime()
        scryptAsynch(password, salt, {
                logN: logN,
                r: 8,
                p: 1,
                dkLen: nacl.secretbox.keyLength, // 32
                interruptStep: 0,
                encoding: 'binary' // hex, base64, binary
            },
            function (res) {
                const time = Math.round(getTime() - t1)
                callback(res, time)
            }
        )
    } catch (ex) {
        alert(ex.message)
    }
}

var getTime = (function () {
    if (typeof performance !== "undefined") {
        return performance.now.bind(performance)
    }
    return Date.now.bind(Date)
})()

export const post = function(method, webObject, url, postSuccess, postError) {
    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(webObject),
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`)
            }
            return response.json()
        })
        .then(data => {
            postSuccess(data)
        })
        .catch(err => {
            postError(err)
        })
}